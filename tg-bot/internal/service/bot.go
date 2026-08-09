package service

import (
	"context"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models"
)

type BotUseCase struct {
	parser   ProductParser
	producer UserActionProducer

	userStates map[int64]models.UserState
	pending    map[int64]models.PendingSubscription
}

// return message to user: "Пришли ссылку на товар"
func (u *BotUseCase) StartAddSubscription(ctx context.Context, telegramUserID int64) string

// send request to monitor service and return text with all size of product
func (u *BotUseCase) HandleProductURL(ctx context.Context, TelegramUserID int64, url string) (string, error)

// when user choose the size service will return him a message: "принял в обработку"
func (u *BotUseCase) SelectSize(ctx context.Context, telegramUserID int64, optionID int64) (string, error)