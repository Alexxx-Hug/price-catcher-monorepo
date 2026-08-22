package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/models"
)

const wildberriesCardDetailURL = "https://card.wb.ru/cards/v4/detail"

type WildberriesParser struct {
	client *http.Client
}

func NewWildberriesParser() *WildberriesParser {
	return &WildberriesParser{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *WildberriesParser) ParseProduct(ctx context.Context, rawURL string) (*models.Product, error) {
	nmID, err := extractNmID(rawURL)
	if err != nil {
		return nil, fmt.Errorf("extract nm_id: %w", err)
	}

	requestURL, err := buildCardDetailURL(nmID)
	if err != nil {
		return nil, fmt.Errorf("build card detail url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create wb request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://www.wildberries.ru/")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request wb card detail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wb card detail returned status %d", resp.StatusCode)
	}

	var cardResponse wbCardResponse
	if err := json.NewDecoder(resp.Body).Decode(&cardResponse); err != nil {
		return nil, fmt.Errorf("decode wb card detail response: %w", err)
	}

	product, err := cardResponse.product()
	if err != nil {
		return nil, err
	}

	return mapWBProduct(product, rawURL), nil
}

func buildCardDetailURL(nmID int64) (string, error) {
	parsedURL, err := url.Parse(wildberriesCardDetailURL)
	if err != nil {
		return "", err
	}

	query := parsedURL.Query()
	query.Set("appType", "1")
	query.Set("curr", "rub")
	query.Set("dest", "-1257786")
	query.Set("spp", "30")
	query.Set("nm", strconv.FormatInt(nmID, 10))

	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

func extractNmID(rawURL string) (int64, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return 0, fmt.Errorf("parse url: %w", err)
	}

	if parsedURL.Host == "" {
		return 0, fmt.Errorf("url host is empty")
	}

	if !strings.Contains(parsedURL.Host, "wildberries.") {
		return 0, fmt.Errorf("not a wildberries url")
	}

	parts := strings.Split(parsedURL.Path, "/")
	for i, part := range parts {
		if part == "catalog" && i+1 < len(parts) {
			nmID, err := strconv.ParseInt(parts[i+1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("parse nm_id from path: %w", err)
			}

			if nmID <= 0 {
				return 0, fmt.Errorf("nm_id must be positive")
			}

			return nmID, nil
		}
	}

	return 0, fmt.Errorf("catalog nm_id not found in url")
}

type wbCardResponse struct {
	Products []wbProduct `json:"products"`
	Data     struct {
		Products []wbProduct `json:"products"`
	} `json:"data"`
}

func (r wbCardResponse) product() (wbProduct, error) {
	if len(r.Products) > 0 {
		return r.Products[0], nil
	}

	if len(r.Data.Products) > 0 {
		return r.Data.Products[0], nil
	}

	return wbProduct{}, fmt.Errorf("wb product not found")
}

func mapWBProduct(product wbProduct, rawURL string) *models.Product {
	sizes := make([]models.ProductSize, 0, len(product.Sizes))
	totalQuantity := product.TotalQuantity

	if totalQuantity == 0 {
		for _, size := range product.Sizes {
			totalQuantity += sumStocks(size.Stocks)
		}
	}

	for _, size := range product.Sizes {
		quantity := sumStocks(size.Stocks)
		price := priceMinor(size.Price, product)

		sizes = append(sizes, models.ProductSize{
			OptionID:   size.OptionID,
			SizeName:   size.Name,
			OrigName:   size.OrigName,
			PriceMinor: int(price),
			Quantity:   quantity,
		})
	}

	return &models.Product{
		NmID:          product.ID,
		Name:          product.Name,
		Brand:         product.Brand,
		URL:           rawURL,
		TotalQuantity: totalQuantity,
		Sizes:         sizes,
	}
}

func sumStocks(stocks []wbStock) int {
	var quantity int
	for _, stock := range stocks {
		quantity += stock.Quantity
	}

	return quantity
}

func priceMinor(sizePrice wbPrice, product wbProduct) int64 {
	switch {
	case sizePrice.Total > 0:
		return sizePrice.Total
	case sizePrice.Product > 0:
		return sizePrice.Product
	case product.SalePriceU > 0:
		return product.SalePriceU
	default:
		return product.PriceU
	}
}
