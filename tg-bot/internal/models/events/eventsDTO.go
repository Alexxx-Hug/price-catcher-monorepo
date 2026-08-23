package events

import (
	"encoding/json"
	"time"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models"
)

type UserActionType string

const (
	UserActionAddSubscription    UserActionType = "add_subscription"
	UserActionDeleteSubscription UserActionType = "delete_subscription"
)

type UserActionEvent struct {
	ActionID       string          `json:"action_id"`
	TelegramUserID int64           `json:"telegram_user_id"`
	Type           UserActionType  `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	CreatedAt      time.Time       `json:"created_at"`
}

type AddSubscriptionPayload struct {
	Product     models.Product     `json:"product"`
	ProductSize models.ProductSize `json:"product_size"`
}

type DeleteSubscriptionPayload struct {
	SubscriptionID int64 `json:"subscription_id"`
}

type ProductPriceChangedEvent struct {
	EventID        string    `json:"event_id"`
	TelegramUserID int64     `json:"telegram_user_id"`
	ProductID      int64     `json:"product_id"`
	ProductName    string    `json:"product_name"`
	ProductSizeID  int64     `json:"product_size_id"`
	Brand          string    `json:"brand"`
	Size           string    `json:"size"`
	URL            string    `json:"url"`
	OldPriceMinor  int       `json:"old_price_minor"`
	NewPriceMinor  int       `json:"new_price_minor"`
	DeltaMinor     int       `json:"delta_minor"`
	ChangedAt      time.Time `json:"changed_at"`
}
