package service

import (
	"context"

	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/models"
)

type ProductParser interface {
	ParseProduct(ctx context.Context, url string) (*models.Product, error)
}
