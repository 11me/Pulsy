package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/11me/pulsy/message"
)

type DiscordNotifier struct {
	Webhook string
}

type DiscordPayload struct {
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Content   string         `json:"content,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordEmbed struct {
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	URL         string         `json:"url,omitempty"`
	Color       int            `json:"color,omitempty"`
	Fields      []DiscordField `json:"fields,omitempty"`
}

type DiscordField struct {
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}

func (d *DiscordNotifier) Notify(m message.Message) error {
	payload := &DiscordPayload{
		Username:  "Pulsy",
		AvatarURL: "https://img.freepik.com/free-vector/frequency-icon_53876-25527.jpg?w=320",
		Content:   m.Message,
	}
	bodyBytes, _ := json.Marshal(payload)
	body := bytes.NewReader(bodyBytes)

	res, err := http.Post(d.Webhook, "application/json", body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to send notification %s", string(bodyBytes))
	}
	return nil
}
