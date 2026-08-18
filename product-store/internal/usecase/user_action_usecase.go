package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/models/eventdto"
	"go.uber.org/zap"
)

type UserActionUseCase struct {
	productUseCase      ProductUseCaseInterface
	subscriptionUseCase SubscriptionUseCaseInterface
	logger              *zap.Logger
}

func NewUserActionUseCase(
	productUseCase ProductUseCaseInterface,
	subscriptionUseCase SubscriptionUseCaseInterface,
	logger *zap.Logger,
) *UserActionUseCase {
	return &UserActionUseCase{
		productUseCase:      productUseCase,
		subscriptionUseCase: subscriptionUseCase,
		logger:              logger,
	}
}

func (u *UserActionUseCase) ProcessUserAction(ctx context.Context, event eventdto.UserActionEvent) error {
	switch event.Type {
	case eventdto.UserActionAddSubcription:
		return u.processAddSubscriptionWithProduct(ctx, event)

	case eventdto.UserActionDeleteSubscription:
		return fmt.Errorf("delete subscription action is not implemented")

	case eventdto.UserActionListSubscriptions:
		return fmt.Errorf("list subscriptions action is not implemented")

	default:
		return fmt.Errorf("unknown user action type: %s", event.Type)
	}
}

func (u *UserActionUseCase) processAddSubscriptionWithProduct(ctx context.Context, event eventdto.UserActionEvent) error {
	var payload eventdto.AddSubscriptionPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal user action payload: %w", err)
	}

	sizes := make([]ProductSizeInput, 0, len(payload.Product.Sizes))

	for _, size := range payload.Product.Sizes {
		sizes = append(sizes, ProductSizeInput{
			OptionID:   size.OptionID,
			SizeName:   size.SizeName,
			OrigName:   size.OrigName,
			PriceMinor: size.PriceMinor,
			Quantity:   size.Quantity,
		})
	}

	var upsertProductInput UpsertProductInput = UpsertProductInput{
		NmID:          payload.Product.NmID,
		Name:          payload.Product.Name,
		Brand:         payload.Product.Brand,
		URL:           payload.Product.URL,
		TotalQuantity: payload.Product.TotalQuantity,
		Sizes:         sizes,
	}

	savedProduct, err := u.productUseCase.UpsertProductWithSizes(ctx, upsertProductInput)
	if err != nil {
		return fmt.Errorf("failed to upsert product with sizes: %w", err)
	}

	var selectedSizeID int64

	for _, size := range savedProduct.Sizes {
		if size.OptionID == payload.ProductSize.OptionID {
			selectedSizeID = size.ID
			break
		}
	}

	if selectedSizeID == 0 {
		return fmt.Errorf("saved product size with option_id=%d not found", payload.ProductSize.OptionID)
	}

	if _, err := u.subscriptionUseCase.CreateSubscription(ctx, event.TelegramUserID, selectedSizeID); err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	return nil
}
