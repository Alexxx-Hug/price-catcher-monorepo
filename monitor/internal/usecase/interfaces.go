package usecase

import (
	"context"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/entity"
)

type PriceMonitor interface {
	Execute(ctx context.Context) error
}

type ProductClient interface {
	FetchTrackedProducts(ctx context.Context) ([]entity.Product, error) // list of products to check
}

type Parser interface {
	ParsePrice(ctx context.Context, url string) (int, error) // parsing product price by url
}

type NotificationSender interface {
	SendProductEvent(ctx context.Context, event entity.ProductEvent) error // send event to kafka
}
