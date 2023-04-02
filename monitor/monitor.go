package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/11me/pulsy/notifier"
	"github.com/11me/pulsy/message"
)

type state int

const (
	INIT state = iota
	ERROR
	PENDING
	OK
)

func (s state) String() string {
	switch s {
	case ERROR:
		return "ERROR"
	case OK:
		return "OK"
	case PENDING:
		return "PENDING"
	default:
		return "UNKNOWN STATE"
	}
}


type Monitor struct {
	URL      string
	Timeout  time.Duration
	Retry    int
	Interval time.Duration
	status   string
	state    state
	errorMsg string
}

type Watcher struct {
	Monitors  []*Monitor
	Notifiers []notifier.Notifier
	Writers   []io.Writer
	done      chan struct{}
	wg        *sync.WaitGroup
}

func (w *Watcher) Watch() {
	w.wg = &sync.WaitGroup{}
	defer w.wg.Wait()

	go w.listenForInterrupt()

	if w.Writers == nil {
		w.Writers[0] = os.Stdout
	}

	w.done = make(chan struct{})
	for _, m := range w.Monitors {
		w.wg.Add(1)
		go w.watchMonitor(m)
	}
}

func (w *Watcher) watchMonitor(m *Monitor) {
	retryCounter := 0
	m.state = INIT

	client := http.Client{
		Timeout: m.Timeout,
	}

	defer w.wg.Done()
	for {

		if w.isDone() {
			return
		}

		tStart := time.Now()
		res, err := client.Get(m.URL)
		tElapsed := time.Since(tStart)

		lastState := m.state
        // if in the last state we encounter an error
        // do not bombard the service with requests
        // give it a breath
        if lastState == ERROR || lastState == PENDING {
            time.Sleep(time.Second * 3)
        }

		if err != nil {
			retryCounter++
			msg := message.Message{
				Timestamp: tStart.Format(time.RFC3339),
				Status:    ERROR.String(),
				Latency:   time.Duration(tElapsed.Milliseconds()),
				URL:       m.URL,
				Message:   err.Error(),
			}
			messageBytes, _ := json.Marshal(&msg)
			if retryCounter > m.Retry {
				m.state = ERROR
				if lastState != ERROR {
					w.callNotifiers(msg)
				}
			} else {
				m.state = PENDING
			}

			w.callWriters(messageBytes)
			continue
		}

		if res.StatusCode != http.StatusOK {
			retryCounter++
			msg := message.Message{
				Timestamp: tStart.Format(time.RFC3339),
				Status:    ERROR.String(),
				Latency:   time.Duration(tElapsed.Milliseconds()),
				URL:       m.URL,
				Message:   res.Status,
			}
			messageBytes, _ := json.Marshal(&msg)
			if retryCounter > m.Retry {
				m.state = ERROR
				if lastState != ERROR {
					w.callNotifiers(msg)
				}
			} else {
				m.state = PENDING
			}
			w.callWriters(messageBytes)
			continue
		}
		retryCounter = 0
		m.state = OK
		msg := message.Message{
			Timestamp: tStart.Format(time.RFC3339),
			Status:    OK.String(),
			Latency:   time.Duration(tElapsed.Milliseconds()),
			URL:       m.URL,
			Message:   res.Status,
		}
		msgBytes, _ := json.Marshal(&msg)
		if lastState == ERROR {
			w.callNotifiers(msg)
		}
		w.callWriters(msgBytes)

		<-time.Tick(m.Interval)
	}
}

func (w *Watcher) isDone() bool {
	select {
	case <-w.done:
		return true
	default:
		return false
	}
}

func (w *Watcher) listenForInterrupt() {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig
	fmt.Fprintf(os.Stdout, "Interrupt signal received, stopping...\n")
	w.Stop()
}

func (w *Watcher) callNotifiers(m message.Message) {
	for _, notifier := range w.Notifiers {
		if err := notifier.Notify(m); err != nil {
			fmt.Fprintf(os.Stderr, "[NOTIFICATION FAILURE]: %s", err.Error())
		}
	}
}

func (w *Watcher) callWriters(message []byte) {
	for _, writer := range w.Writers {
		writer.Write(message)
	}
}

func (w *Watcher) Stop() {
	close(w.done)
	w.wg.Wait()
}
