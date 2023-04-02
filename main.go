package main

import (
	"io"
	"log"

	"github.com/11me/pulsy/config"
	"github.com/11me/pulsy/monitor"
	"github.com/11me/pulsy/writer"
)

func main() {
	if err := config.ReadConfig(); err != nil {
		log.Fatalln(err)
	}

	monitors := config.LoadMonitors()
	notifiers := config.LoadNotifiers()

	consoleWriter := &writer.ConsoleWriter{}

	watcher := monitor.Watcher{
		Monitors:  monitors,
		Notifiers: notifiers,
		Writers: []io.Writer{
			consoleWriter,
		},
	}
	watcher.Watch()
}
