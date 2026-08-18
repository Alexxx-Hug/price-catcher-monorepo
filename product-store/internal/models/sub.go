package entity

import (
	"time"

	apperrors "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/app_errors"
)

type Subscription struct {
	ID             int64
	TelegramUserID int64
	ProductSizeID  int64
	CreatedAt      time.Time
}

func NewSubscription(telegramUserID int64, productSizeID int64) (*Subscription, error) {
	if telegramUserID <= 0 {
		return nil, apperrors.ErrInvalidTelegramUserID
	}

	if productSizeID <= 0 {
		return nil, apperrors.ErrInvalidProductSizeID
	}

	return &Subscription{
		TelegramUserID: telegramUserID,
		ProductSizeID:  productSizeID,
		CreatedAt:      time.Now(),
	}, nil
}

type UserSubscriptionItem struct {
	SubscriptionID int64
	ProductID      int64
	ProductSizeID  int64
	NmID           int64
	ProductName    string
	Brand          string
	SizeName       string
	PriceMinor     int
	URL            string
}
