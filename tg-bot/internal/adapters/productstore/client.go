package productstore

import (
	"context"
	"fmt"
	"time"

	productstorepb "github.com/Alexxx-Hug/price-catcher-monorepo/gen/go/productstore"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models"
	"google.golang.org/grpc"
)

type Client struct {
	client  productstorepb.SubscriptionServiceClient
	timeout time.Duration
}

func NewClient(conn *grpc.ClientConn, timeout time.Duration) *Client {
	return &Client{
		client:  productstorepb.NewSubscriptionServiceClient(conn),
		timeout: timeout,
	}
}

func (c *Client) ListUserSubscriptions(ctx context.Context, telegramUserID int64) ([]models.Subscription, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	response, err := c.client.ListUserSubscriptions(ctx, &productstorepb.ListUserSubscriptionsRequest{
		TelegramUserId: telegramUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("list user subscriptions grpc call: %w", err)
	}

	subscriptions := make([]models.Subscription, 0, len(response.Subscriptions))
	for _, sub := range response.Subscriptions {
		subscriptions = append(subscriptions, models.Subscription{
			SubscriptionID: sub.SubscriptionId,
			ProductID:      sub.ProductId,
			ProductSizeID:  sub.ProductSizeId,
			NmID:           sub.NmId,
			ProductName:    sub.ProductName,
			Brand:          sub.Brand,
			SizeName:       sub.SizeName,
			PriceMinor:     int(sub.PriceMinor),
			URL:            sub.Url,
		})
	}

	return subscriptions, nil
}
