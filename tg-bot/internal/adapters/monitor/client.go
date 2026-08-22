package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/Alexxx-Hug/price-catcher-monorepo/gen/go/monitor"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models"
	"google.golang.org/grpc"
)

type Client struct {
	client  monitor.MonitorServiceClient
	timeout time.Duration
}

func NewClient(conn *grpc.ClientConn, timeout time.Duration) *Client {
	return &Client{
		client:  monitor.NewMonitorServiceClient(conn),
		timeout: timeout,
	}
}

func (c *Client) ParseProduct(ctx context.Context, url string) (*models.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	response, err := c.client.ParseProduct(ctx, &monitor.ParseProductRequest{
		Url: url,
	})
	if err != nil {
		return nil, fmt.Errorf("parse product grpc call: %w", err)
	}

	sizes := make([]models.ProductSize, 0, len(response.GetSizes()))

	for _, size := range response.GetSizes() {
		sizes = append(sizes, models.ProductSize{
			OptionID:   size.GetOptionId(),
			SizeName:   size.GetSizeName(),
			OrigName:   size.GetOrigName(),
			PriceMinor: int(size.GetPriceMinor()),
			Quantity:   int(size.GetQuantity()),
		})
	}

	return &models.Product{
		NmID:          response.GetNmId(),
		Name:          response.GetName(),
		Brand:         response.GetBrand(),
		URL:           response.GetUrl(),
		TotalQuantity: int(response.GetTotalQuantity()),
		Sizes:         sizes,
	}, nil
}
