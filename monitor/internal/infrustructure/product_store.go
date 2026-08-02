package infrustructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/entity"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/infrustructure/dto"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type ProductStoreClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewProductStoreClient(url string, timeout time.Duration) *ProductStoreClient {
	return &ProductStoreClient{
		baseURL: url,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *ProductStoreClient) FetchTrackedProducts(ctx context.Context) ([]entity.Product, error) {
	url := fmt.Sprintf("%s/api/v1/products", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("product_client: failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("product_client: request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product_client: unexpected status code: %d", resp.StatusCode)
	}

	var dtos []dto.ProductDTO
	if err := json.NewDecoder(resp.Body).Decode(&dtos); err != nil {
		return nil, fmt.Errorf("product client: failed to decode response body: %w", err)
	}

	products := make([]entity.Product, len(dtos))
	for i, dto := range dtos {
		products[i] = entity.Product{
			ID:        strconv.FormatInt(dto.ID, 10),
			URL:       dto.URL,
			Size:      dto.Size,
			Price:     dto.Price,
			UpdatedAt: dto.UpdatedAt,
		}
	}

	return products, nil
}

var wbArticle = regexp.MustCompile(`catalog/(\d+)/detail`)

func (c *ProductStoreClient) ParsePrice(ctx context.Context, url string) (int, error) {
	matches := wbArticle.FindStringSubmatch(url)
	if len(matches) < 2 {
		return 0, fmt.Errorf("ParsePrice: failed to extract article from url")
	}

	article := matches[1]

	wbAPIURL := fmt.Sprintf("https://card.wb.ru/cards/v2/detail?appType=1&curr=rub&dest=-1257786&srg=5&nm=%s", article)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wbAPIURL, nil)
	if err != nil {
		return 0, fmt.Errorf("ParsePrice: failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("ParsePrice: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("ParsePrice: unexpected status code: %d", resp.StatusCode)
	}

	var wbResp dto.WBResponse
	if err := json.NewDecoder(resp.Body).Decode(&wbResp); err != nil {
		return 0, fmt.Errorf("ParsePrice: failed to decode response body: %w", err)
	}

	if len(wbResp.Data.Products) == 0 {
		return 0, fmt.Errorf("ParsePrice: product not found")
	}

	return int(wbResp.Data.Products[0].Price.Product), nil
}
