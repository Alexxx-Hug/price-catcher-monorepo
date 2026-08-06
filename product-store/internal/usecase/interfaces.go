package usecase

import (
	"context"

	entity "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/models"
	eventdto "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/models/eventdto"
)

type ProductUseCaseInterface interface {
	UpsertProductWithSizes(ctx context.Context, input UpsertProductInput) (*entity.Product, error)
	GetProductByNmID(ctx context.Context, nmID int64) (*entity.Product, error)
	GetProductSizeByOptionID(ctx context.Context, optionID int64) (*entity.ProductSize, error)
	ListProductsForMonitoring(ctx context.Context, limit int) ([]entity.Product, error)
	PublishPriceCheckTasks(ctx context.Context, limit int) error
	ProcessCheckedProduct(ctx context.Context, event eventdto.ProductCheckedEvent) error
}
