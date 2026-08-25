package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/metrics"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/models/eventdto"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type UserActionHandler interface {
	ProcessUserAction(ctx context.Context, event eventdto.UserActionEvent) error
}

type UserActionConsumer struct {
	reader  *kafka.Reader
	topic   string
	handler UserActionHandler
	logger  *zap.Logger
	metrics *metrics.Metrics
}

func NewUserActionConsumer(
	brokers []string,
	topic string,
	groupID string,
	handler UserActionHandler,
	logger *zap.Logger,
	appMetrics *metrics.Metrics,
) *UserActionConsumer {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &UserActionConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Topic:   topic,
			GroupID: groupID,
			Brokers: brokers,
		}),
		topic:   topic,
		handler: handler,
		logger:  logger,
		metrics: appMetrics,
	}
}

func (c *UserActionConsumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			c.metrics.IncError(c.topic, "fetch")
			return fmt.Errorf("fetch user action message: %w", err)
		}

		startedAt := time.Now()
		c.logger.Info(
			"user action message fetched",
			zap.String("topic", c.topic),
			zap.Int64("offset", message.Offset),
			zap.String("key", string(message.Key)),
		)

		var event eventdto.UserActionEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			c.logger.Error("failed to unmarshal user action event", zap.Error(err))
			c.metrics.IncError(c.topic, "unmarshal")

			if err := c.reader.CommitMessages(ctx, message); err != nil {
				c.metrics.IncError(c.topic, "commit")
				c.metrics.ObserveConsumed(c.topic, "error", startedAt)
				return fmt.Errorf("commit user action message: %w", err)
			}

			c.metrics.ObserveConsumed(c.topic, "invalid", startedAt)
			continue
		}

		if err := c.handler.ProcessUserAction(ctx, event); err != nil {
			c.logger.Error(
				"failed to process user action event",
				zap.String("action_id", event.ActionID),
				zap.String("type", string(event.Type)),
				zap.Int64("telegram_user_id", event.TelegramUserID),
				zap.Error(err),
			)
			c.metrics.IncError(c.topic, "process")
			c.metrics.ObserveConsumed(c.topic, "error", startedAt)
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			c.metrics.IncError(c.topic, "commit")
			c.metrics.ObserveConsumed(c.topic, "error", startedAt)
			return fmt.Errorf("commit user action message: %w", err)
		}

		c.logger.Info(
			"user action message processed",
			zap.String("topic", c.topic),
			zap.Int64("offset", message.Offset),
			zap.String("key", string(message.Key)),
		)
		c.metrics.ObserveConsumed(c.topic, "success", startedAt)
	}
}

func (c *UserActionConsumer) Close() error {
	return c.reader.Close()
}
