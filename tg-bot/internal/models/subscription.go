package models

type Subscription struct {
	SubscriptionID int64
	ProductID      int64
	ProductSizeID  int64
	NmID           int64
	ProductName    string
	Brand          string
	SizeName       string
	PriceMinor     int
	URL            string
}

type DeleteSubscriptionResult struct {
	Text          string
	Subscriptions []Subscription
}
