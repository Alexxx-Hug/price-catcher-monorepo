package service

import (
	"context"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models/events"
)

type ProductParser interface {
	ParseProduct(ctx context.Context, url string) (*models.Product, error)
}

type UserActionProducer interface {
	SendUserAction(ctx context.Context, event events.UserActionEvent) error
}

type SubscriptionProvider interface {
	ListUserSubscriptions(ctx context.Context, telegramUserID int64) ([]models.Subscription, error)
}
