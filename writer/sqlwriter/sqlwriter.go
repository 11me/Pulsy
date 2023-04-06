package sqlwriter

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/11me/pulsy/message"
	_ "github.com/mattn/go-sqlite3"
)

type sqlWriter struct {
	db        *sql.DB
	semaphore chan struct{}
}

func NewSQLWriter() *sqlWriter {
	conn, err := sql.Open("sqlite3", "./monitors.db")
	sem := make(chan struct{}, 1)

	if err != nil {
		log.Fatalln(err)
	}
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS monitors
      (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          timestamp TEXT,
          status TEXT,
          latency_ms INTEGER,
          url TEXT,
          message TEXT
      )`)

	return &sqlWriter{
		db:        conn,
		semaphore: sem,
	}
}

func (w *sqlWriter) Write(ctx context.Context, b []byte) error {
	w.semaphore <- struct{}{}
	defer func() { <-w.semaphore }()

	var m message.Message
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	_, err = w.db.ExecContext(ctx, "INSERT INTO monitors (timestamp, status, latency_ms, url, message) VALUES (?, ?, ?, ?, ?)",
		m.Timestamp, m.Status, m.Latency, m.URL, m.Message)

	return nil
}

func (w *sqlWriter) Close() error {
    return w.db.Close()
}
