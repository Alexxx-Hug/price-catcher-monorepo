package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/metrics"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models/events"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type PriceChangedHandler interface {
	ProcessProductPriceChanged(ctx context.Context, event events.ProductPriceChangedEvent) error
}

type PriceChangedConsumer struct {
	reader  *kafka.Reader
	topic   string
	handler PriceChangedHandler
	logger  *zap.Logger
	metrics *metrics.Metrics
}

func NewPriceChangedConsumer(topic string, brokers []string, groupID string, handler PriceChangedHandler, logger *zap.Logger, appMetrics *metrics.Metrics) *PriceChangedConsumer {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &PriceChangedConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
		}),
		topic:   topic,
		handler: handler,
		logger:  logger,
		metrics: appMetrics,
	}
}

func (c *PriceChangedConsumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			c.metrics.IncError(c.topic, "fetch")
			return fmt.Errorf("fetch product price changed message: %w", err)
		}

		startedAt := time.Now()
		var event events.ProductPriceChangedEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			c.logger.Error("failed to unmarshal product price changed event", zap.Error(err))
			c.metrics.IncError(c.topic, "unmarshal")

			if err := c.reader.CommitMessages(ctx, message); err != nil {
				c.metrics.IncError(c.topic, "commit")
				c.metrics.ObserveConsumed(c.topic, "error", startedAt)
				return fmt.Errorf("commit invalid product price changed message: %w", err)
			}

			c.metrics.ObserveConsumed(c.topic, "invalid", startedAt)
			continue
		}

		if err := c.handler.ProcessProductPriceChanged(ctx, event); err != nil {
			c.logger.Error(
				"failed to process product price changed event",
				zap.String("event_id", event.EventID),
				zap.Int64("product_size_id", event.ProductSizeID),
				zap.Error(err),
			)
			c.metrics.IncError(c.topic, "process")
			c.metrics.ObserveConsumed(c.topic, "error", startedAt)
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			c.metrics.IncError(c.topic, "commit")
			c.metrics.ObserveConsumed(c.topic, "error", startedAt)
			return fmt.Errorf("commit product price changed message: %w", err)
		}

		c.metrics.ObserveConsumed(c.topic, "success", startedAt)
	}
}

func (c *PriceChangedConsumer) Close() error {
	return c.reader.Close()
}
