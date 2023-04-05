package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/11me/pulsy/message"
)

const TELEGRAM_API = "https://api.telegram.org/bot"

type TelegramNotifier struct {
	Token string
    ChatID string
}

type telegramPayload struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func (t *TelegramNotifier) Notify(m message.Message) error {
    telegramApi := TELEGRAM_API + t.Token + "/sendMessage"
    msgFormat := "%s: %s \n\nStatus: %s \nLatency: %d ms \nURL: %s"
    msg := fmt.Sprintf(msgFormat,  "❗️Alert", m.Message, m.Status, m.Latency, m.URL)
    if m.Status == "OK" {
        msg = fmt.Sprintf(msgFormat,  "✅ OK", m.Message, m.Status, m.Latency, m.URL)
    }

    payload := telegramPayload{
        ChatID: t.ChatID,
        Text: string(msg),
    }
    payloadBytes, _ := json.Marshal(&payload)
    payloadBytesReader := bytes.NewReader(payloadBytes)

    res, err := http.Post(telegramApi, "application/json", payloadBytesReader)
    if err != nil {
        return err
    }
    if res.StatusCode != http.StatusOK {
        bodyBytes, err := io.ReadAll(res.Body)
        if err != nil {
            return err
        }
        return fmt.Errorf("failed to send telegram notification %s %s", res.Status, string(bodyBytes))
    }
    
	return nil
}
