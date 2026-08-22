package parser

type wbProduct struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Brand         string   `json:"brand"`
	TotalQuantity int      `json:"totalQuantity"`
	PriceU        int64    `json:"priceU"`
	SalePriceU    int64    `json:"salePriceU"`
	Sizes         []wbSize `json:"sizes"`
}

type wbSize struct {
	OptionID int64     `json:"optionId"`
	Name     string    `json:"name"`
	OrigName string    `json:"origName"`
	Price    wbPrice   `json:"price"`
	Stocks   []wbStock `json:"stocks"`
}

type wbPrice struct {
	Basic   int64 `json:"basic"`
	Product int64 `json:"product"`
	Total   int64 `json:"total"`
}

type wbStock struct {
	Quantity int `json:"qty"`
}
