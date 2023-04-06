package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/11me/pulsy/message"
	"github.com/11me/pulsy/notifier"
	"github.com/11me/pulsy/writer"
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
	Monitors      []*Monitor
	Notifiers     []notifier.Notifier
	Writers       []writer.Writer
	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	wg            *sync.WaitGroup
}

func (w *Watcher) Watch() {
	ctx, cancel := context.WithCancel(context.Background())
	w.ctx = ctx
	w.ctxCancelFunc = cancel

	w.wg = &sync.WaitGroup{}
	defer w.wg.Wait()

	go w.listenForInterrupt()

	if w.Writers == nil {
		w.Writers[0] = &writer.ConsoleWriter{}
	}

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
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stdout, "[RECOVER]: %v\n", r)
			fmt.Fprintf(os.Stdout, "[RECOVER]: %s: %v\n", "restarting the monitor", m)
			w.wg.Add(1)
			go w.watchMonitor(m)
		}
	}()
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
			time.Sleep(m.Interval)
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
			msgBytes, _ := json.Marshal(&msg)
			if retryCounter > m.Retry {
				m.state = ERROR
				if lastState != ERROR {
					w.callNotifiers(msg)
				}
			} else {
				m.state = PENDING
			}

			w.callWriters(msgBytes)
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
			msgBytes, _ := json.Marshal(&msg)
			if retryCounter > m.Retry {
				m.state = ERROR
				if lastState != ERROR {
					w.callNotifiers(msg)
				}
			} else {
				m.state = PENDING
			}
			w.callWriters(msgBytes)
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

		time.Sleep(m.Interval)
	}
}

func (w *Watcher) isDone() bool {
	select {
	case <-w.ctx.Done():
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
		go writer.Write(w.ctx, message)
	}
}

func (w *Watcher) Stop() {
    w.ctxCancelFunc()
	w.wg.Wait()
}
