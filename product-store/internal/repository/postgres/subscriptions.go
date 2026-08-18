package postgres

import (
	"context"
	"errors"
	"fmt"

	apperrors "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/app_errors"
	entity "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	postgresUniqueViolation     = "23505"
	postgresForeignKeyViolation = "23503"
)

type subscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *subscriptionRepository {
	return &subscriptionRepository{
		db: db,
	}
}

func (r *subscriptionRepository) CreateSubscription(ctx context.Context, sub *entity.Subscription) (*entity.Subscription, error) {
	const createSubQuery = `
	INSERT INTO shop.subscriptions (telegram_user_id, product_size_id, created_at)
	VALUES ($1,$2,$3)
	RETURNING id
	`

	if err := r.db.QueryRow(ctx, createSubQuery, sub.TelegramUserID, sub.ProductSizeID, sub.CreatedAt).Scan(&sub.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case postgresUniqueViolation:
				return nil, fmt.Errorf("%w: telegram_user_id=%d product_size_id=%d", apperrors.ErrSubscriptionExists, sub.TelegramUserID, sub.ProductSizeID)
			case postgresForeignKeyViolation:
				return nil, fmt.Errorf("%w: product_size_id=%d", apperrors.ErrProductSizeNotFound, sub.ProductSizeID)
			}
		}

		return nil, fmt.Errorf("create subscription: %w", err)
	}

	return sub, nil
}

func (r *subscriptionRepository) ListUserSubscriptions(ctx context.Context, telegramUserID int64) ([]entity.Subscription, error) {
	const getUserSubscriptions = `
	SELECT id, telegram_user_id, product_size_id, created_at FROM shop.subscriptions
	WHERE telegram_user_id = $1
	`

	rows, err := r.db.Query(ctx, getUserSubscriptions, telegramUserID)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}

	defer rows.Close()

	subscriptions := make([]entity.Subscription, 0)

	for rows.Next() {
		var sub entity.Subscription
		if err := rows.Scan(
			&sub.ID,
			&sub.TelegramUserID,
			&sub.ProductSizeID,
			&sub.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}

		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user subscriptions telegram_user_id=%d: %w", telegramUserID, err)
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) DeleteSubscriptionAndCleanupProduct(ctx context.Context, telegramUserID int64, productSizeID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var productID int64

	err = tx.QueryRow(ctx, `
		SELECT product_id
		FROM shop.product_sizes
		WHERE id = $1
	`, productSizeID).Scan(&productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: product_size_id=%d", apperrors.ErrProductSizeNotFound, productSizeID)
		}

		return fmt.Errorf("get product id by product_size_id=%d: %w", productSizeID, err)
	}

	res, err := tx.Exec(ctx, `
		DELETE FROM shop.subscriptions
		WHERE telegram_user_id = $1
		  AND product_size_id = $2
	`, telegramUserID, productSizeID)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("%w: telegram_user_id=%d product_size_id=%d", apperrors.ErrSubscriptionNotFound, telegramUserID, productSizeID)
	}

	var hasSubscriptions bool

	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM shop.subscriptions s
			JOIN shop.product_sizes ps ON ps.id = s.product_size_id
			WHERE ps.product_id = $1
		)
	`, productID).Scan(&hasSubscriptions)
	if err != nil {
		return fmt.Errorf("check product subscriptions: %w", err)
	}

	if !hasSubscriptions {
		_, err = tx.Exec(ctx, `
			DELETE FROM shop.products
			WHERE id = $1
		`, productID)
		if err != nil {
			return fmt.Errorf("delete product without subscriptions: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *subscriptionRepository) ListSubscriptionsByProductSizeID(ctx context.Context, productSizeID int64) ([]entity.Subscription, error) {
	const getListSubQuery = `
		SELECT id, telegram_user_id, product_size_id, created_at FROM shop.subscriptions
		WHERE product_size_id = $1
	`

	rows, err := r.db.Query(ctx, getListSubQuery, productSizeID)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions by product size: %w", err)
	}

	defer rows.Close()

	subscriptions := make([]entity.Subscription, 0)

	for rows.Next() {
		var sub entity.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.TelegramUserID,
			&sub.ProductSizeID,
			&sub.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}

		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriptions product_size_id=%d: %w", productSizeID, err)
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) ListUserSubscriptionItems(ctx context.Context, telegramUserID int64) ([]entity.UserSubscriptionItem, error) {
	const query = `
	SELECT s.id,p.id,ps.id,p.nm_id,p.name,p.brand,ps.name,ps.price_minor,p.url
	FROM shop.subscriptions s
	JOIN shop.product_sizes ps ON ps.id = s.product_size_id
	JOIN shop.products p ON p.id = ps.product_id
	WHERE s.telegram_user_id = $1
	ORDER BY s.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, telegramUserID)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}

	defer rows.Close()

	listUserSubscriptionsWithProduct := make([]entity.UserSubscriptionItem, 0)

	for rows.Next() {
		var sub entity.UserSubscriptionItem
		if err := rows.Scan(
			&sub.SubscriptionID,
			&sub.ProductID,
			&sub.ProductSizeID,
			&sub.NmID,
			&sub.ProductName,
			&sub.Brand,
			&sub.SizeName,
			&sub.PriceMinor,
			&sub.URL,
		); err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}

		listUserSubscriptionsWithProduct = append(listUserSubscriptionsWithProduct, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user subscriptions telegram_user_id=%d: %w", telegramUserID, err)
	}

	return listUserSubscriptionsWithProduct, nil
}

func (r *subscriptionRepository) GetSubscriptionByIDAndTelegramUserID(ctx context.Context, subscriptionID, telegramUserID int64) (*entity.Subscription, error) {
	const getSubQuery = `
	SELECT id, telegram_user_id, product_size_id, created_at FROM shop.subscriptions
	WHERE id = $1 AND telegram_user_id = $2
	`

	var sub entity.Subscription
	if err := r.db.QueryRow(ctx, getSubQuery, subscriptionID, telegramUserID).Scan(
		&sub.ID,
		&sub.TelegramUserID,
		&sub.ProductSizeID,
		&sub.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: subscription_id=%d telegram_user_id=%d", apperrors.ErrSubscriptionNotFound, subscriptionID, telegramUserID)
		}

		return nil, fmt.Errorf("get subscription by id=%d telegram_user_id=%d: %w", subscriptionID, telegramUserID, err)
	}

	return &sub, nil
}
