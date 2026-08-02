package apperrors

import "errors"

// зачем делать отдельный пакет под константы по факту
// можно сделать пакет vars и там сделать errors.go и всё
var (
	ErrInvalidPrice     = errors.New("invalid product price")
	ErrInvalidProductID = errors.New("product id must be positive")
	ErrInvalidNmID      = errors.New("nm id must be positive")
	ErrInvalidOptionID  = errors.New("option id must be positive")
	ErrInvalidQuantity  = errors.New("quantity must be non-negative")

	ErrProductNotFound      = errors.New("product not found")
	ErrProductSizeNotFound  = errors.New("product size not found")
	ErrSubscriptionNotFound = errors.New("subscription not found")
	ErrSubscriptionExists   = errors.New("subscription already exists")

	ErrEmptyProductName  = errors.New("product name cannot be empty")
	ErrEmptyProductURL   = errors.New("product url cannot be empty")
	ErrEmptySizeName     = errors.New("size name cannot be empty")
	ErrEmptyProductSizes = errors.New("product sizes cannot be empty")

	ErrInvalidTelegramUserID = errors.New("telegram user id must be positive")
	ErrInvalidProductSizeID  = errors.New("product size id must be positive")
)
