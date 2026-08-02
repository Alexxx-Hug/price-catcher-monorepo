package handlers

import (
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/usecase"

	"go.uber.org/zap"
)

type SubscriptionHandler struct {
	subHandler usecase.SubscriptionUseCaseInterface
	logger     *zap.Logger
}

func NewSubscriptionHandler(subHandler usecase.SubscriptionUseCaseInterface, logger *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		subHandler: subHandler,
		logger:     logger,
	}
}
