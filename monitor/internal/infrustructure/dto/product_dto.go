package dto

import "time"

type ProductDTO struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	Size      string    `json:"size"`
	Price     int       `json:"price"`
	UpdatedAt time.Time `json:"updated_at"`
}
