package dto

type UpsertProductRequest struct {
	NmID          int64  `json:"nm_id" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Brand         string `json:"brand" binding:"required"`
	URL           string `json:"url" binding:"required"`
	TotalQuantity int    `json:"total_quantity" binding:"required"`
	Sizes         []struct {
		OptionID   int64  `json:"option_id" binding:"required"`
		SizeName   string `json:"size_name" binding:"size_name"`
		OrigName   string `json:"orig_name" binding:"required"`
		PriceMinor int    `json:"price_minor" binding:"gt=0"`
		Quantity   int    `json:"quantity" binding:"required"`
	} `json:"sizes" binding:"required"`
}
