package message

import "time"


type Message struct {
	Timestamp string        `json:"@timestamp"`
	Status    string        `json:"status"`
	Latency   time.Duration `json:"latency_ms"`
	URL       string        `json:"url"`
	Message   string        `json:"message"`
}

