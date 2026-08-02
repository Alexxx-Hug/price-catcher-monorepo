package entity

import "time"

type PriceChangeType string

const (
	PriceDropped PriceChangeType = "DROPPED"
	PriceRaised  PriceChangeType = "RAISED"
)

type ProductEvent struct {
	ProductID   string          `json:"product_id"`
	URL         string          `json:"url"`
	OldPrice    int             `json:"old_price"`
	NewPrice    int             `json:"new_price"`
	ChangeType  PriceChangeType `json:"change_price"`
	Delta       int             `json:"delta"`
	TriggeredAt time.Time       `json:"triggered_at"`
}
