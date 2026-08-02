package usecase

import (
	"context"
	"fmt"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/entity"
	"time"
)

type Monitor struct {
	parser             Parser
	productClient      ProductClient
	notificationSender NotificationSender
}

func NewMonitor(parser Parser, productClient ProductClient, notificationSender NotificationSender) *Monitor {
	return &Monitor{
		parser:             parser,
		productClient:      productClient,
		notificationSender: notificationSender,
	}
}

func (m *Monitor) Execute(ctx context.Context) error {
	products, err := m.productClient.FetchTrackedProducts(ctx)
	if err != nil {
		return fmt.Errorf("usecase: failed to fetch tracked products: %w", err)
	}

	if len(products) == 0 {
		return nil
	}

	for _, product := range products {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("usecase: execution interrupted: %w", err)
		}

		currentPrice, err := m.parser.ParsePrice(ctx, product.URL)
		if err != nil {
			continue
		}

		if currentPrice == product.Price {
			continue
		}

		var changeType entity.PriceChangeType
		var delta int

		if currentPrice < product.Price {
			changeType = entity.PriceDropped
			delta = product.Price - currentPrice
		} else {
			changeType = entity.PriceRaised
			delta = currentPrice - product.Price
		}

		event := entity.ProductEvent{
			ProductID:   product.ID,
			URL:         product.URL,
			OldPrice:    product.Price,
			NewPrice:    currentPrice,
			ChangeType:  changeType,
			Delta:       delta,
			TriggeredAt: time.Now(),
		}

		if err := m.notificationSender.SendProductEvent(ctx, event); err != nil {
			continue
		}
	}

	return nil
}
