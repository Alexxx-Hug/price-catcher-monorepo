package parser

import (
	"context"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models"
)

// mock parser
type ProductParser struct{}

func (p *ProductParser) ParseProduct(ctx context.Context, url string) (*models.Product, error) {
	return &models.Product{
		NmID:          123,
		Name:          "mock_product",
		Brand:         "mock_brand",
		URL:           "mock_url",
		TotalQuantity: 10,
		Sizes: []models.ProductSize{
			{OptionID: 111, SizeName: "42", OrigName: "42", PriceMinor: 199900, Quantity: 3},
			{OptionID: 222, SizeName: "43", OrigName: "43", PriceMinor: 209900, Quantity: 2},
		},
	}, nil
}
