package repository

import (
	"context"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/entity"
)

type ProductRepository interface {
	// ProductUseCase: создает или обновляет товар и размеры в одной транзакции.
	UpsertProductWithSizes(ctx context.Context, product *entity.Product) (*entity.Product, error)

	// ProductUseCase: получает товар WB по nm_id вместе со всеми размерами.
	GetProductByNmID(ctx context.Context, nmID int64) (*entity.Product, error)

	// ProductUseCase: получает конкретный размер WB по option_id.
	GetProductSizeByOptionID(ctx context.Context, optionID int64) (*entity.ProductSize, error)

	// ProductUseCase: возвращает товары, у которых есть активные подписки и которые нужно мониторить.
	ListProductsForMonitoring(ctx context.Context) ([]entity.Product, error)
}

type SubscriptionRepository interface {
	// SubscriptionUseCase: создает связь user -> product_size.
	CreateSubscription(ctx context.Context, sub *entity.Subscription) (*entity.Subscription, error)

	// SubscriptionUseCase: возвращает все подписки конкретного Telegram-пользователя.
	ListUserSubscriptions(ctx context.Context, telegramUserID int64) ([]entity.Subscription, error)

	// SubscriptionUseCase: удаляет подписку пользователя с проверкой владельца через telegram_user_id.
	DeleteSubscriptionAndCleanupProduct(ctx context.Context, telegramUserID int64, productSizeID int64) error

	// SubscriptionUseCase: возвращает всех пользователей, подписанных на конкретный размер.
	ListSubscriptionsByProductSizeID(ctx context.Context, productSizeID int64) ([]entity.Subscription, error)
}
