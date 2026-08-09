package models

type Product struct {
	NmID          int64
	Name          string
	Brand         string
	URL           string
	TotalQuantity int
	Sizes         []ProductSize
}

type ProductSize struct {
	OptionID   int64
	SizeName   string
	OrigName   string
	PriceMinor int
	Quantity   int
}

type UserState string

const (
	StateIdle              UserState = "idle"
	StateWaitingProductURL UserState = "waiting_product_url"
	StateWaitingSizeChoice UserState = "waiting_size_choice"
)

type PendingSubscription struct {
	Product Product
}
