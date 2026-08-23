package eventdto

import "time"

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
