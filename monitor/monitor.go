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

type MessageFormat struct {
	Timestamp string        `json:"@timestamp"`
	Status    string        `json:"status"`
	Latency   time.Duration `json:"latency_ms"`
	URL       string        `json:"url"`
	Message   string        `json:"message"`
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
		if err != nil {
			retryCounter++
			message := &MessageFormat{
				Timestamp: tStart.Format(time.RFC3339),
				Status:    ERROR.String(),
				Latency:   time.Duration(tElapsed.Milliseconds()),
				URL:       m.URL,
				Message:   err.Error(),
			}
			messageBytes, _ := json.Marshal(message)
			if retryCounter > m.Retry {
				m.state = ERROR
				if lastState != ERROR {
					w.callNotifiers(messageBytes)
				}
			} else {
				m.state = PENDING
			}

			w.callWriters(messageBytes)
			continue
		}

		if res.StatusCode != http.StatusOK {
			retryCounter++
			message := &MessageFormat{
				Timestamp: tStart.Format(time.RFC3339),
				Status:    ERROR.String(),
				Latency:   time.Duration(tElapsed.Milliseconds()),
				URL:       m.URL,
				Message:   res.Status,
			}
			messageBytes, _ := json.Marshal(message)
			if retryCounter > m.Retry {
				m.state = ERROR
				if lastState != ERROR {
					w.callNotifiers(messageBytes)
				}
			} else {
				m.state = PENDING
			}
			w.callWriters(messageBytes)
			continue
		}
		retryCounter = 0
		m.state = OK
		message := &MessageFormat{
			Timestamp: tStart.Format(time.RFC3339),
			Status:    OK.String(),
			Latency:   time.Duration(tElapsed.Milliseconds()),
			URL:       m.URL,
			Message:   res.Status,
		}
		messageBytes, _ := json.Marshal(message)
		if lastState == ERROR {
			w.callNotifiers(messageBytes)
		}
		w.callWriters(messageBytes)

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

func (w *Watcher) callNotifiers(message []byte) {
	for _, notifier := range w.Notifiers {
		if err := notifier.Notify(message); err != nil {
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
