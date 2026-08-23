package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/models"
	eventdto "github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/models/eventDTO"
)

type MonitorService struct {
	parser   ProductParser
	producer ProductCheckedProducer
}

func NewMonitorService(parser ProductParser, producer ProductCheckedProducer) *MonitorService {
	return &MonitorService{
		parser:   parser,
		producer: producer,
	}
}

type MonitorServiceInterface interface {
	ParseProduct(ctx context.Context, url string) (*models.Product, error)
}

func (s *MonitorService) ParseProduct(ctx context.Context, url string) (*models.Product, error) {
	if url == "" {
		return nil, fmt.Errorf("failed: url is empty")
	}

	product, err := s.parser.ParseProduct(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse product: %w", err)
	}

	return product, nil
}

func (s *MonitorService) ProcessCheckPrice(ctx context.Context, event eventdto.TaskCheckPricesEvent) error {
	product, err := s.parser.ParseProduct(ctx, event.URL)
	if err != nil {
		return fmt.Errorf("process check price: %w", err)
	}

	if product == nil {
		return fmt.Errorf("failed to get product")
	}

	for _, size := range product.Sizes {
		if event.OptionID == size.OptionID {
			var output eventdto.ProductCheckedEvent = eventdto.ProductCheckedEvent{
				TaskID:        event.TaskID,
				ProductID:     event.ProductID,
				ProductSizeID: event.ProductSizeID,
				OptionID:      size.OptionID,
				PriceMinor:    size.PriceMinor,
				Quantity:      size.Quantity,
				InStock:       size.Quantity > 0,
				CheckedAt:     time.Now(),
			}

			return s.producer.SendProductChecked(ctx, output)
		}
	}

	errorMessage := fmt.Sprintf("product size option_id=%d not found", event.OptionID)

	output := eventdto.ProductCheckedEvent{
		TaskID:        event.TaskID,
		ProductID:     event.ProductID,
		ProductSizeID: event.ProductSizeID,
		OptionID:      event.OptionID,
		CheckedAt:     time.Now(),
		Error:         &errorMessage,
	}

	return s.producer.SendProductChecked(ctx, output)

}
