package producer

import (
	"context"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models/events"
	"go.uber.org/zap"
)

type MockKafkaUserActionProducer struct {
	logger *zap.Logger
}

func (p *MockKafkaUserActionProducer) SendUserAction(ctx context.Context, event events.UserActionEvent) error {
	p.logger.Info("user action event produced", zap.Any("event", event))
	return nil
}
