package postgres

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/app_errors"
	entity "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type productRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *productRepository {
	return &productRepository{
		db: db,
	}
}

func (r *productRepository) UpsertProductWithSizes(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	const productQuery = `
		INSERT INTO shop.products (
			nm_id, name, brand, url, total_quantity, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (nm_id) DO UPDATE SET
			name = EXCLUDED.name,
			brand = EXCLUDED.brand,
			url = EXCLUDED.url,
			total_quantity = EXCLUDED.total_quantity,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	err = tx.QueryRow(
		ctx,
		productQuery,
		product.NmID,
		product.Name,
		product.Brand,
		product.URL,
		product.TotalQuantity,
		product.UpdatedAt).Scan(&product.ID)

	if err != nil {
		return nil, fmt.Errorf("upsert product: %w", err)
	}

	const sizeQuery = `
		INSERT INTO shop.product_sizes (
			product_id, option_id, name, orig_name, price_minor, quantity, in_stock, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (product_id, option_id) DO UPDATE SET
			name = EXCLUDED.name,
			orig_name = EXCLUDED.orig_name,
			price_minor = EXCLUDED.price_minor,
			quantity = EXCLUDED.quantity,
			in_stock = EXCLUDED.in_stock,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	for i := range product.Sizes {
		product.Sizes[i].ProductID = product.ID

		err = tx.QueryRow(
			ctx,
			sizeQuery,
			product.Sizes[i].ProductID,
			product.Sizes[i].OptionID,
			product.Sizes[i].Name,
			product.Sizes[i].OrigName,
			product.Sizes[i].PriceMinor,
			product.Sizes[i].Quantity,
			product.Sizes[i].InStock,
			product.Sizes[i].UpdatedAt,
		).Scan(&product.Sizes[i].ID)
		if err != nil {
			return nil, fmt.Errorf("upsert product size option_id=%d: %w", product.Sizes[i].OptionID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return product, nil
}

func (r *productRepository) GetProductByNmID(ctx context.Context, nmID int64) (*entity.Product, error) {
	const getProductQuery = `
		SELECT id, nm_id, name, brand, url, total_quantity, updated_at FROM shop.products
		WHERE nm_id = $1
	`

	var product entity.Product
	err := r.db.QueryRow(ctx, getProductQuery, nmID).Scan(
		&product.ID,
		&product.NmID,
		&product.Name,
		&product.Brand,
		&product.URL,
		&product.TotalQuantity,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: nm_id=%d", apperrors.ErrProductNotFound, nmID)
		}

		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	const getProductSizesQuery = `
		SELECT id, product_id, option_id, name, orig_name, price_minor, quantity, in_stock, updated_at
		FROM shop.product_sizes
		WHERE product_id = $1
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, getProductSizesQuery, product.ID)
	if err != nil {
		return nil, fmt.Errorf("query product sizes product_id=%d: %w", product.ID, err)
	}

	defer rows.Close()

	product.Sizes = make([]entity.ProductSize, 0)

	for rows.Next() {
		var size entity.ProductSize

		err := rows.Scan(
			&size.ID,
			&size.ProductID,
			&size.OptionID,
			&size.Name,
			&size.OrigName,
			&size.PriceMinor,
			&size.Quantity,
			&size.InStock,
			&size.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan product size product_id=%d: %w", product.ID, err)
		}

		product.Sizes = append(product.Sizes, size)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product sizes product_id=%d: %w", product.ID, err)
	}

	return &product, nil
}

func (r *productRepository) GetProductSizeByOptionID(ctx context.Context, optionID int64) (*entity.ProductSize, error) {
	const getProductSizeQuery = `
		SELECT id, product_id, option_id, name, orig_name, price_minor, quantity, in_stock, updated_at FROM shop.product_sizes
		WHERE option_id = $1
	`

	var size entity.ProductSize
	err := r.db.QueryRow(ctx, getProductSizeQuery, optionID).Scan(
		&size.ID,
		&size.ProductID,
		&size.OptionID,
		&size.Name,
		&size.OrigName,
		&size.PriceMinor,
		&size.Quantity,
		&size.InStock,
		&size.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: option_id=%d", apperrors.ErrProductSizeNotFound, optionID)
		}

		return nil, fmt.Errorf("get product size by option_id=%d: %w", optionID, err)
	}

	return &size, nil

}

func (r *productRepository) ListProductsForMonitoring(ctx context.Context, limit int) ([]entity.Product, error) {
	if limit <= 0 {
		limit = 5
	}

	const query = `
		WITH sizes_to_check AS (
			SELECT
				ps.id
			FROM shop.product_sizes ps
			JOIN shop.subscriptions s ON s.product_size_id = ps.id
			GROUP BY ps.id
			ORDER BY MIN(ps.updated_at) ASC
			LIMIT $1
		)
		SELECT
			p.id,
			p.nm_id,
			p.name,
			p.brand,
			p.url,
			p.total_quantity,
			p.updated_at,
			ps.id,
			ps.product_id,
			ps.option_id,
			ps.name,
			ps.orig_name,
			ps.price_minor,
			ps.quantity,
			ps.in_stock,
			ps.updated_at
		FROM sizes_to_check stc
		JOIN shop.product_sizes ps ON ps.id = stc.id
		JOIN shop.products p ON p.id = ps.product_id
		ORDER BY ps.updated_at ASC, ps.id
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query products for monitoring: %w", err)
	}
	defer rows.Close()

	productsByID := make(map[int64]*entity.Product)
	productOrder := make([]int64, 0)

	for rows.Next() {
		var product entity.Product
		var size entity.ProductSize

		err := rows.Scan(
			&product.ID,
			&product.NmID,
			&product.Name,
			&product.Brand,
			&product.URL,
			&product.TotalQuantity,
			&product.UpdatedAt,
			&size.ID,
			&size.ProductID,
			&size.OptionID,
			&size.Name,
			&size.OrigName,
			&size.PriceMinor,
			&size.Quantity,
			&size.InStock,
			&size.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan product for monitoring: %w", err)
		}

		existingProduct, ok := productsByID[product.ID]
		if !ok {
			product.Sizes = make([]entity.ProductSize, 0)
			productsByID[product.ID] = &product
			productOrder = append(productOrder, product.ID)
			existingProduct = &product
		}

		existingProduct.Sizes = append(existingProduct.Sizes, size)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate products for monitoring: %w", err)
	}

	products := make([]entity.Product, 0, len(productOrder))
	for _, productID := range productOrder {
		products = append(products, *productsByID[productID])
	}

	return products, nil
}
