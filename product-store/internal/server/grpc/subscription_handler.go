package grpc

import (
	"context"

	"github.com/Alexxx-Hug/price-catcher-monorepo/gen/go/productstore"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SubscriptionHandler struct {
	productstore.UnimplementedSubscriptionServiceServer
	subscriptionUseCase usecase.SubscriptionUseCaseInterface
}

func NewSubscriptionHandler(subscriptionUseCase usecase.SubscriptionUseCaseInterface) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionUseCase: subscriptionUseCase,
	}
}

func (h *SubscriptionHandler) ListUserSubscriptions(ctx context.Context, req *productstore.ListUserSubscriptionsRequest) (*productstore.ListUserSubscriptionsResponse, error) {
	items, err := h.subscriptionUseCase.ListUserSubscriptionItems(ctx, req.TelegramUserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list user subscriptions: %v", err)
	}

	response := productstore.ListUserSubscriptionsResponse{
		Subscriptions: make([]*productstore.Subscription, 0, len(items)),
	}

	for _, item := range items {
		response.Subscriptions = append(response.Subscriptions, &productstore.Subscription{
			SubscriptionId: item.SubscriptionID,
			ProductId:      item.ProductID,
			ProductSizeId:  item.ProductSizeID,
			NmId:           item.NmID,
			ProductName:    item.ProductName,
			Brand:          item.Brand,
			SizeName:       item.SizeName,
			PriceMinor:     int64(item.PriceMinor),
			Url:            item.URL,
		})
	}

	return &response, nil
}
