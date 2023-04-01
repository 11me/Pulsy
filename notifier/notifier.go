package notifier

import (
	"fmt"
	"os"

	"github.com/11me/pulsy/notifier/telegram"
)

type Notifier interface {
	Notify(message []byte) error
}
type DefaultNotifier struct{}

func (n *DefaultNotifier) Notify(message []byte) error {
    fmt.Fprintf(os.Stderr, "[DEFAULT NOTIFIER]: %s", message)
	return nil
}

type NotifierFactory func(options map[string]interface{}) Notifier

func MakeNotifierFactory(name string) NotifierFactory {
    switch name {
    case "telegram":
        return func(options map[string]interface{}) Notifier {
            return &telegram.TelegramNotifier{
                ChatID: options["chat_id"].(string),
                Token: options["token"].(string),
            }
        }
    default:
        return func(options map[string]interface{}) Notifier {
            return &DefaultNotifier{}
        }
    }
}
