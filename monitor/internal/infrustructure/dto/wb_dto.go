package dto

type WBResponse struct {
	Data struct {
		Products []WBProduct `json:"products"`
	} `json:"data"`
}

type WBProduct struct {
	ID    int64   `json:"id"`
	Price WBPrice `json:"price"`
}

type WBPrice struct {
	Product int64 `json:"product"`
}
