package usecase

import (
	"context"
	"errors"
	"fmt"
	apperrors "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/app_errors"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/entity"
	repository "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/repository"

	"go.uber.org/zap"
)

type ProductUseCaseInterface interface {
	// tg-bot: сохраняет товар после парсинга WB, когда пользователь добавляет ссылку.
	// monitor: обновляет товар, размеры, цены и остатки после очередной проверки WB.
	UpsertProductWithSizes(ctx context.Context, request UpsertProductInput) (*entity.Product, error)

	// tg-bot: проверяет, есть ли товар уже в базе, когда пользователь прислал WB-ссылку.
	GetProductByNmID(ctx context.Context, nmID int64) (*entity.Product, error)

	// tg-bot: получает выбранный пользователем размер перед созданием подписки.
	GetProductSizeByOptionID(ctx context.Context, optionID int64) (*entity.ProductSize, error)

	// monitor: получает список товаров, которые нужно периодически проверять в WB.
	ListProductsForMonitoring(ctx context.Context) ([]entity.Product, error)
}

type productUseCase struct {
	repo   repository.ProductRepository
	logger *zap.Logger
}

func NewProductUseCase(repo repository.ProductRepository, logger *zap.Logger) *productUseCase {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &productUseCase{
		repo:   repo,
		logger: logger,
	}
}

type UpsertProductInput struct {
	NmID          int64
	Name          string
	Brand         string
	URL           string
	TotalQuantity int
	Sizes         []ProductSizeInput
}

type ProductSizeInput struct {
	OptionID   int64
	SizeName   string
	OrigName   string
	PriceMinor int
	Quantity   int
}

func (u *productUseCase) UpsertProductWithSizes(ctx context.Context, input UpsertProductInput) (*entity.Product, error) {
	productSizes := make([]entity.ProductSize, 0, len(input.Sizes))

	for _, inputSize := range input.Sizes {
		size, err := entity.NewProductSize(
			inputSize.OptionID,
			inputSize.SizeName,
			inputSize.OrigName,
			inputSize.PriceMinor,
			inputSize.Quantity,
		)

		if err != nil {
			u.logger.Error("failed to create product size", zap.Int64("option_id", inputSize.OptionID), zap.Error(err))
			return nil, fmt.Errorf("create product size option_id=%d: %w", inputSize.OptionID, err)
		}

		productSizes = append(productSizes, *size)
	}

	product, err := entity.NewProduct(input.NmID, input.Name, input.Brand, input.URL, input.TotalQuantity, productSizes)
	if err != nil {
		u.logger.Error("failed to create new product", zap.Int64("nm_id", input.NmID), zap.Error(err))
		return nil, fmt.Errorf("failed to create new product: %w", err)
	}

	savedProduct, err := u.repo.UpsertProductWithSizes(ctx, product)
	if err != nil {
		u.logger.Error("failed to upsert product", zap.Int64("nm_id", input.NmID), zap.Error(err))
		return nil, fmt.Errorf("upsert product with sizes: %w", err)
	}

	u.logger.Info("product upserted", zap.Int64("product_id", savedProduct.ID), zap.Int64("nm_id", savedProduct.NmID), zap.Int("sizes_count", len(savedProduct.Sizes)))
	return savedProduct, nil
}

func (u *productUseCase) GetProductByNmID(ctx context.Context, nmID int64) (*entity.Product, error) {
	if nmID <= 0 {
		u.logger.Warn("invalid nm_id", zap.Int64("nm_id", nmID))
		return nil, fmt.Errorf("%w: nm_id=%d", apperrors.ErrInvalidNmID, nmID)
	}

	product, err := u.repo.GetProductByNmID(ctx, nmID)
	if err != nil {
		if errors.Is(err, apperrors.ErrProductNotFound) {
			u.logger.Warn("product not found", zap.Int64("nm_id", nmID))
			return nil, fmt.Errorf("get product by nm_id=%d: %w", nmID, err)
		}

		u.logger.Error("failed to get product by nm_id",
			zap.Int64("nm_id", nmID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("get product by nm_id=%d: %w", nmID, err)
	}

	if product == nil {
		u.logger.Warn("product not found", zap.Int64("nm_id", nmID))
		return nil, fmt.Errorf("%w: nm_id=%d", apperrors.ErrProductNotFound, nmID)
	}

	return product, nil
}

func (u *productUseCase) GetProductSizeByOptionID(ctx context.Context, optionID int64) (*entity.ProductSize, error) {
	if optionID <= 0 {
		return nil, apperrors.ErrInvalidOptionID
	}

	productSize, err := u.repo.GetProductSizeByOptionID(ctx, optionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product size: %w", err)
	}

	if productSize == nil {
		return nil, apperrors.ErrProductSizeNotFound
	}

	return productSize, nil
}

func (u *productUseCase) ListProductsForMonitoring(ctx context.Context) ([]entity.Product, error) {
	products, err := u.repo.ListProductsForMonitoring(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of products: %w", err)
	}

	if len(products) == 0 {
		return nil, apperrors.ErrProductNotFound
	}

	return products, nil
}
