package producer

import (
	"context"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models/events"
	"go.uber.org/zap"
)

type MockKafkaUserActionProducer struct {
	Logger *zap.Logger
}

func NewMockKafkaUserActionProducer(logger *zap.Logger) *MockKafkaUserActionProducer {
	return &MockKafkaUserActionProducer{
		Logger: logger,
	}
}

func (p *MockKafkaUserActionProducer) SendUserAction(ctx context.Context, event events.UserActionEvent) error {
	p.Logger.Info("user action event produced", zap.Any("event", event))
	return nil
}
