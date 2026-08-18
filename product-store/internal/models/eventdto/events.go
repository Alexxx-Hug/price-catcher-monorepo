package eventdto

import (
	"encoding/json"
	"time"
)

type UserActionType string

const (
	UserActionAddSubcription     UserActionType = "add_subscription"
	UserActionDeleteSubscription UserActionType = "delete_subscription"
	UserActionListSubscriptions  UserActionType = "list_subscriptions"
)

// ивент для топика task-check-prices
type TaskCheckPricesEvent struct {
	TaskID        string    `json:"task_id"`
	ProductID     int64     `json:"product_id"`
	ProductSizeID int64     `json:"product_size_id"`
	NmID          int64     `json:"nm_id"`
	OptionID      int64     `json:"option_id"`
	URL           string    `json:"url"`
	PriceMinor    int       `json:"price_minor"`
	RequestedAt   time.Time `json:"requested_at"`
}

type ProductCheckedEvent struct {
	TaskID        string    `json:"task_id"`
	ProductID     int64     `json:"product_id"`
	ProductSizeID int64     `json:"product_size_id"`
	OptionID      int64     `json:"option_id"`
	PriceMinor    int       `json:"price_minor"`
	Quantity      int       `json:"quantity"`
	InStock       bool      `json:"in_stock"`
	CheckedAt     time.Time `json:"checked_at"`
	Error         *string   `json:"error,omitempty"`
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

type UserActionEvent struct {
	ActionID       string          `json:"action_id"`
	TelegramUserID int64           `json:"telegram_user_id"`
	Type           UserActionType  `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	CreatedAt      time.Time       `json:"created_at"`
}

type AddSubscriptionPayload struct {
	Product     ProductPayload     `json:"product"`
	ProductSize ProductSizePayload `json:"product_size"`
}

type ProductPayload struct {
	NmID          int64                `json:"nm_id"`
	Name          string               `json:"name"`
	Brand         string               `json:"brand"`
	URL           string               `json:"url"`
	TotalQuantity int                  `json:"total_quantity"`
	Sizes         []ProductSizePayload `json:"sizes"`
}

type ProductSizePayload struct {
	OptionID   int64  `json:"option_id"`
	SizeName   string `json:"size_name"`
	OrigName   string `json:"orig_name"`
	PriceMinor int    `json:"price_minor"`
	Quantity   int    `json:"quantity"`
}

