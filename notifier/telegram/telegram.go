package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func (t *TelegramNotifier) Notify(message []byte) error {
    telegramApi := TELEGRAM_API + t.Token + "/sendMessage"
    payload := telegramPayload{
        ChatID: t.ChatID,
        Text: string(message),
    }
    b, _ := json.Marshal(&payload)
    bReader := bytes.NewReader(b)

    res, err := http.Post(telegramApi, "application/json", bReader)
    if err != nil {
        return err
    }
    if res.StatusCode != http.StatusOK {
        msgBytes, err := io.ReadAll(res.Body)
        if err != nil {
            return err
        }
        return fmt.Errorf("failed to send notification %s", string(msgBytes))
    }
    
	return nil
}
