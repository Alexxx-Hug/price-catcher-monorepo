package parser

import (
	"context"

	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/models"
)

type FakeParser struct {
}

func (p *FakeParser) ParseProduct(ctx context.Context, url string) (*models.Product, error) {
	return &models.Product{
		NmID:          123,
		Name:          "test",
		Brand:         "test_brand",
		URL:           url,
		TotalQuantity: 4,
		Sizes: []models.ProductSize{
			{OptionID: 123, SizeName: "test_size", OrigName: "test_size", PriceMinor: 1000, Quantity: 4},
		},
	}, nil

}
