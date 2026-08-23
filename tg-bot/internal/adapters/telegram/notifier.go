package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Notifier struct {
	api *tgbotapi.BotAPI
}

func NewNotifier(api *tgbotapi.BotAPI) *Notifier {
	return &Notifier{
		api: api,
	}
}

func (n *Notifier) SendMessage(ctx context.Context, telegramUserID int64, text string) error {
	msg := tgbotapi.NewMessage(telegramUserID, text)

	if _, err := n.api.Send(msg); err != nil {
		return fmt.Errorf("send telegram notification: %w", err)
	}

	return nil
}
