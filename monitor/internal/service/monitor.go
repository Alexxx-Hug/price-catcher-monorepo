package service

import (
	"context"
	"fmt"

	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/models"
)

type MonitorService struct {
	parser ProductParser
}

func NewMonitorService(parser ProductParser) *MonitorService {
	return &MonitorService{
		parser: parser,
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
