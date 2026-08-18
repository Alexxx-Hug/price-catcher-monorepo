package usecase

import (
	"context"
	"fmt"

	apperrors "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/app_errors"
	entity "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/models"
)

type SubscriptionDataBaseRepository interface {
	// tg-bot: создает подписку пользователя на конкретный размер товара.
	CreateSubscription(ctx context.Context, sub *entity.Subscription) (*entity.Subscription, error)

	// tg-bot: показывает пользователю список товаров, которые он отслеживает.
	ListUserSubscriptions(ctx context.Context, telegramUserID int64) ([]entity.Subscription, error)

	// tg-bot: удаляет подписку пользователя и товар из БД, если его никто не отслеживает
	DeleteSubscriptionAndCleanupProduct(ctx context.Context, telegramUserID int64, productSizeID int64) error

	// monitor: получает подписки на размер товара, если цена изменилась или стала ниже целевой.
	ListSubscriptionsByProductSizeID(ctx context.Context, productSizeID int64) ([]entity.Subscription, error)

	ListUserSubscriptionItems(ctx context.Context, telegramUserID int64) ([]entity.UserSubscriptionItem, error)

	GetSubscriptionByIDAndTelegramUserID(ctx context.Context, subscriptionID, telegramUserID int64) (*entity.Subscription, error)
}

type SubscriptionUseCase struct {
	repo SubscriptionDataBaseRepository
}

func NewSubcriptionUseCase(repo SubscriptionDataBaseRepository) *SubscriptionUseCase {
	return &SubscriptionUseCase{
		repo: repo,
	}
}

func (h *SubscriptionUseCase) CreateSubscription(ctx context.Context, telegramUserID int64, productSizeID int64) (*entity.Subscription, error) {
	subscription, err := entity.NewSubscription(telegramUserID, productSizeID)
	if err != nil {
		return nil, fmt.Errorf("failed to create new subscription for telegram user id %v with product size %v: %w", telegramUserID, productSizeID, err)
	}

	id, err := h.repo.CreateSubscription(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to save subscription: %w", err)
	}

	return id, err
}

func (h *SubscriptionUseCase) ListUserSubscriptions(ctx context.Context, telegramUserID int64) ([]entity.Subscription, error) {
	subscriptions, err := h.repo.ListUserSubscriptions(ctx, telegramUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	if subscriptions == nil {
		return nil, apperrors.ErrSubscriptionNotFound
	}

	return subscriptions, nil
}

func (h *SubscriptionUseCase) DeleteSubscriptionAndCleanupProduct(ctx context.Context, telegramUserID int64, productSizeID int64) error {
	err := h.repo.DeleteSubscriptionAndCleanupProduct(ctx, telegramUserID, productSizeID)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

func (h *SubscriptionUseCase) ListSubscriptionsByProductSizeID(ctx context.Context, productSizeID int64) ([]entity.Subscription, error) {
	subscriptions, err := h.repo.ListSubscriptionsByProductSizeID(ctx, productSizeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription list: %w", err)
	}

	if subscriptions == nil {
		return nil, apperrors.ErrSubscriptionNotFound
	}

	return subscriptions, nil
}

func (h *SubscriptionUseCase) ListUserSubscriptionItems(ctx context.Context, telegramUserID int64) ([]entity.UserSubscriptionItem, error) {
	items, err := h.repo.ListUserSubscriptionItems(ctx, telegramUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscription items: %w", err)
	}

	if items == nil {
		return nil, apperrors.ErrSubscriptionNotFound
	}

	return items, nil
}

func (h *SubscriptionUseCase) GetSubscriptionByIDAndTelegramUserID(ctx context.Context, subscriptionID, telegramUserID int64) (*entity.Subscription, error) {
	if subscriptionID <= 0 {
		return nil, fmt.Errorf("invalid subscription_id=%d", subscriptionID)
	}

	if telegramUserID <= 0 {
		return nil, fmt.Errorf("invalid telegram_user_id=%d", telegramUserID)
	}

	subscription, err := h.repo.GetSubscriptionByIDAndTelegramUserID(ctx, subscriptionID, telegramUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return subscription, nil
}
