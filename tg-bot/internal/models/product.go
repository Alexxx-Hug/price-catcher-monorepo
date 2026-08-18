package models

type Product struct {
	NmID          int64         `json:"nm_id"`
	Name          string        `json:"name"`
	Brand         string        `json:"brand"`
	URL           string        `json:"url"`
	TotalQuantity int           `json:"total_quantity"`
	Sizes         []ProductSize `json:"sizes"`
}

type ProductSize struct {
	OptionID   int64  `json:"option_id"`
	SizeName   string `json:"size_name"`
	OrigName   string `json:"orig_name"`
	PriceMinor int    `json:"price_minor"`
	Quantity   int    `json:"quantity"`
}

type UserState string

const (
	StateIdle                                UserState = "idle"
	StateWaitingProductURL                   UserState = "waiting_product_url"
	StateWaitingSizeChoiceForAddSubscription UserState = "waiting_size_choice_for_add_sub"
	StateWaitingSubscriptionChoiceForDelete  UserState = "waiting_subscription_choice_for_delete"
)

type PendingSubscription struct {
	Product Product
}
