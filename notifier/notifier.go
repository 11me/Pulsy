package notifier

import (
	"fmt"
	"os"

	"github.com/11me/pulsy/message"
	"github.com/11me/pulsy/notifier/discord"
	"github.com/11me/pulsy/notifier/telegram"
)


type Notifier interface {
	Notify(m message.Message) error
}
type DefaultNotifier struct{}

func (n *DefaultNotifier) Notify(m message.Message) error {
    fmt.Fprintf(os.Stderr, "[DEFAULT NOTIFIER]: %v", m)
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
    case "discord":
        return func(options map[string]interface{}) Notifier {
            return &discord.DiscordNotifier{
                Webhook: options["webhook"].(string),
            }
        }
    default:
        return func(options map[string]interface{}) Notifier {
            return &DefaultNotifier{}
        }
    }
}
