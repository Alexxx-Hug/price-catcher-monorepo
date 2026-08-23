package service

import (
	"context"

	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/models"
	eventdto "github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/models/eventDTO"
)

type ProductParser interface {
	ParseProduct(ctx context.Context, url string) (*models.Product, error)
}

type ProductCheckedProducer interface {
	SendProductChecked(ctx context.Context, event eventdto.ProductCheckedEvent) error
}
