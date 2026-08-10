package events

import (
	"encoding/json"
	"time"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models"
)

type UserActionType string

const UserActionAddSubscription UserActionType = "add_subscription"

type UserActionEvent struct {
	ActionID       string          `json:"action_id"`
	TelegramUserID int64           `json:"telegram_user_id"`
	Type           UserActionType  `json:"type"`
	Payload        json.RawMessage `json:"payload"`
	CreatedAt      time.Time       `json:"created_at"`
}

type AddSubscriptionPayload struct {
	Product          models.Product `json:"product"`
	SelectedOptionID int64          `json:"selected_option_id"`
}
