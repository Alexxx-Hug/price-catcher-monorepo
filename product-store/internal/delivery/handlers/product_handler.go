package handlers

import (
	"errors"
	"net/http"
	"strconv"

	apperrors "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/app_errors"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/delivery/dto"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/usecase"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductHandler struct {
	usecase usecase.ProductUseCaseInterface
	logger  *zap.Logger
}

func NewProductHandler(productUseCase usecase.ProductUseCaseInterface, logger *zap.Logger) *ProductHandler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &ProductHandler{
		usecase: productUseCase,
		logger:  logger,
	}
}

func (h *ProductHandler) UpsertProductWithSizes(c *gin.Context) {
	var req dto.UpsertProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body", "INVALID_REQUEST_BODY")
		return
	}

	input := usecase.UpsertProductInput{
		NmID:          req.NmID,
		Name:          req.Name,
		Brand:         req.Brand,
		URL:           req.URL,
		TotalQuantity: req.TotalQuantity,
		Sizes:         make([]usecase.ProductSizeInput, 0, len(req.Sizes)),
	}

	for _, size := range req.Sizes {
		input.Sizes = append(input.Sizes, usecase.ProductSizeInput{
			OptionID:   size.OptionID,
			SizeName:   size.SizeName,
			OrigName:   size.OrigName,
			PriceMinor: size.PriceMinor,
			Quantity:   size.Quantity,
		})
	}

	product, err := h.usecase.UpsertProductWithSizes(c.Request.Context(), input)
	if err != nil {
		h.handleProductError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) GetProductByNmID(c *gin.Context) {
	nmID, ok := parsePositiveInt64Param(c, "nm_id")
	if !ok {
		writeError(c, http.StatusBadRequest, "invalid nm_id", "INVALID_NM_ID")
		return
	}

	product, err := h.usecase.GetProductByNmID(c.Request.Context(), nmID)
	if err != nil {
		h.handleProductError(c, err)
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) GetProductSizeByOptionID(c *gin.Context) {
	optionID, ok := parsePositiveInt64Param(c, "option_id")
	if !ok {
		writeError(c, http.StatusBadRequest, "invalid option_id", "INVALID_OPTION_ID")
		return
	}

	size, err := h.usecase.GetProductSizeByOptionID(c.Request.Context(), optionID)
	if err != nil {
		h.handleProductError(c, err)
		return
	}

	c.JSON(http.StatusOK, size)
}

func (h *ProductHandler) ListProductsForMonitoring(c *gin.Context) {
	products, err := h.usecase.ListProductsForMonitoring(c.Request.Context())
	if err != nil {
		h.handleProductError(c, err)
		return
	}

	response := make([]gin.H, 0)
	for _, product := range products {
		for _, size := range product.Sizes {
			response = append(response, gin.H{
				"id":         size.ID,
				"url":        product.URL,
				"size":       size.Name,
				"price":      size.PriceMinor,
				"updated_at": size.UpdatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProductHandler) handleProductError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrInvalidNmID), errors.Is(err, apperrors.ErrInvalidProductID):
		writeError(c, http.StatusBadRequest, "invalid product id", "INVALID_PRODUCT_ID")
	case errors.Is(err, apperrors.ErrInvalidOptionID):
		writeError(c, http.StatusBadRequest, "invalid option id", "INVALID_OPTION_ID")
	case errors.Is(err, apperrors.ErrEmptyProductName):
		writeError(c, http.StatusBadRequest, "product name cannot be empty", "EMPTY_PRODUCT_NAME")
	case errors.Is(err, apperrors.ErrEmptyProductURL):
		writeError(c, http.StatusBadRequest, "product url cannot be empty", "EMPTY_PRODUCT_URL")
	case errors.Is(err, apperrors.ErrEmptySizeName):
		writeError(c, http.StatusBadRequest, "size name cannot be empty", "EMPTY_SIZE_NAME")
	case errors.Is(err, apperrors.ErrInvalidQuantity):
		writeError(c, http.StatusBadRequest, "quantity must be non-negative", "INVALID_QUANTITY")
	case errors.Is(err, apperrors.ErrInvalidPrice):
		writeError(c, http.StatusBadRequest, "product price must be positive", "INVALID_PRICE")
	case errors.Is(err, apperrors.ErrProductNotFound):
		writeError(c, http.StatusNotFound, "product not found", "PRODUCT_NOT_FOUND")
	case errors.Is(err, apperrors.ErrProductSizeNotFound):
		writeError(c, http.StatusNotFound, "product size not found", "PRODUCT_SIZE_NOT_FOUND")
	default:
		h.logger.Error("product request failed", zap.Error(err))
		writeError(c, http.StatusInternalServerError, "internal server error", "INTERNAL_ERROR")
	}
}

func parsePositiveInt64Param(c *gin.Context, name string) (int64, bool) {
	value, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, false
	}

	return value, true
}

func writeError(c *gin.Context, status int, message string, code string) {
	c.JSON(status, dto.ErrorResponse{
		Message: message,
		Code:    code,
	})
}
