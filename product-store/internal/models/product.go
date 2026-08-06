package entity

import (
	"strings"
	"time"

	apperrors "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/app_errors"
)

type Product struct {
	ID            int64
	NmID          int64
	Name          string
	Brand         string
	URL           string
	TotalQuantity int
	UpdatedAt     time.Time
	Sizes         []ProductSize
}

type ProductSize struct {
	ID         int64
	ProductID  int64
	OptionID   int64
	Name       string
	OrigName   string
	PriceMinor int
	Quantity   int
	InStock    bool
	UpdatedAt  time.Time
}

type ProductSizeWithProduct struct {
	ProductID     int64
	ProductSizeID int64
	NmID          int64
	OptionID      int64
	ProductName   string
	Brand         string
	URL           string
	Size          string
	OldPriceMinor int
	Quantity      int
	InStock       bool
}

type ProductSizeCheckInput struct {
	ProductSizeID int64
	PriceMinor    int
	Quantity      int
	InStock       bool
	UpdatedAt     time.Time
}

func NewProduct(nmID int64, name string, brand string, url string, totalQuantity int, sizes []ProductSize) (*Product, error) {
	name = strings.TrimSpace(name)
	brand = strings.TrimSpace(brand)
	url = strings.TrimSpace(url)

	if nmID <= 0 {
		return nil, apperrors.ErrInvalidProductID
	}

	if name == "" {
		return nil, apperrors.ErrEmptyProductName
	}

	if url == "" {
		return nil, apperrors.ErrEmptyProductURL
	}

	if totalQuantity < 0 {
		return nil, apperrors.ErrInvalidQuantity
	}

	now := time.Now()

	for i := range sizes {
		sizes[i].ProductID = 0
		sizes[i].InStock = sizes[i].Quantity > 0

		if sizes[i].UpdatedAt.IsZero() {
			sizes[i].UpdatedAt = now
		}
	}

	return &Product{
		NmID:          nmID,
		Name:          name,
		Brand:         brand,
		URL:           url,
		TotalQuantity: totalQuantity,
		UpdatedAt:     now,
		Sizes:         sizes,
	}, nil
}

func NewProductSize(optionID int64, name string, origName string, priceMinor int, quantity int) (*ProductSize, error) {
	name = strings.TrimSpace(name)
	origName = strings.TrimSpace(origName)

	if optionID <= 0 {
		return nil, apperrors.ErrInvalidOptionID
	}

	if name == "" {
		return nil, apperrors.ErrEmptySizeName
	}

	if quantity < 0 {
		return nil, apperrors.ErrInvalidQuantity
	}

	if quantity > 0 && priceMinor <= 0 {
		return nil, apperrors.ErrInvalidPrice
	}

	return &ProductSize{
		OptionID:   optionID,
		Name:       name,
		OrigName:   origName,
		PriceMinor: priceMinor,
		Quantity:   quantity,
		InStock:    quantity > 0,
		UpdatedAt:  time.Now(),
	}, nil
}
