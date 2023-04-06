package main

import (
	"log"

	"github.com/11me/pulsy/config"
	"github.com/11me/pulsy/monitor"
	"github.com/11me/pulsy/writer"
	"github.com/11me/pulsy/writer/sqlwriter"
)

func main() {
	if err := config.ReadConfig(); err != nil {
		log.Fatalln(err)
	}

	monitors := config.LoadMonitors()
	notifiers := config.LoadNotifiers()

	sqlWriter := sqlwriter.NewSQLWriter()
	defer func() {
		if err := sqlWriter.Close(); err != nil {
			log.Println(err)
		}
	}()
	consoleWriter := &writer.ConsoleWriter{}
	watcher := monitor.Watcher{
		Monitors:  monitors,
		Notifiers: notifiers,
		Writers: []writer.Writer{
			consoleWriter,
			sqlWriter,
		},
	}
	watcher.Watch()
}
