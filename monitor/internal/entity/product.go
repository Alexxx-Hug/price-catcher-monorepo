package entity

import "time"

type Product struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Size      string    `json:"size"`
	Price     int       `json:"base_price"`
	UpdatedAt time.Time `json:"updated_at"`
}
