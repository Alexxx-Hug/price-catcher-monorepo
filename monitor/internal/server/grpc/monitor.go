package grpc

import (
	"context"

	"github.com/Alexxx-Hug/price-catcher-monorepo/gen/go/monitor"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MonitorHandler struct {
	monitor.UnimplementedMonitorServiceServer
	monitorService service.MonitorServiceInterface
}

func NewMonitorHandler(monitorService service.MonitorServiceInterface) *MonitorHandler {
	return &MonitorHandler{
		monitorService: monitorService,
	}
}

func (h *MonitorHandler) ParseProduct(ctx context.Context, req *monitor.ParseProductRequest) (*monitor.ParseProductResponse, error) {
	url := req.GetUrl()
	if url == "" {
		return nil, status.Error(codes.InvalidArgument, "url is empty")
	}

	product, err := h.monitorService.ParseProduct(ctx, url)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "parse product: %v", err)
	}

	if product == nil {
		return nil, status.Error(codes.Internal, "parsed product is nil")
	}

	var sizes []*monitor.ProductSize
	for _, size := range product.Sizes {
		sizes = append(sizes, &monitor.ProductSize{
			OptionId:   size.OptionID,
			SizeName:   size.SizeName,
			OrigName:   size.OrigName,
			PriceMinor: int64(size.PriceMinor),
			Quantity:   int32(size.Quantity),
		})
	}

	response := monitor.ParseProductResponse{
		NmId:          product.NmID,
		Name:          product.Name,
		Brand:         product.Brand,
		Url:           product.URL,
		TotalQuantity: int32(product.TotalQuantity),
		Sizes:         sizes,
	}

	return &response, nil
}
