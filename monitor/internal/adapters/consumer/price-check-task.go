package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/metrics"
	eventdto "github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/models/eventDTO"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type PriceCheckHandler interface {
	ProcessCheckPrice(ctx context.Context, event eventdto.TaskCheckPricesEvent) error
}

type PriceCheckConsumer struct {
	reader  *kafka.Reader
	topic   string
	handler PriceCheckHandler
	logger  *zap.Logger
	metrics *metrics.Metrics
}

func NewPriceCheckConsumer(brokers []string, topic string, groupID string, handler PriceCheckHandler, logger *zap.Logger, appMetrics *metrics.Metrics) *PriceCheckConsumer {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &PriceCheckConsumer{
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

func (c *PriceCheckConsumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			c.metrics.IncError(c.topic, "fetch")
			return fmt.Errorf("failed to fetch price check message: %w", err)
		}

		startedAt := time.Now()
		var event eventdto.TaskCheckPricesEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			c.logger.Error("failed to unmarshal price check event", zap.Error(err))
			c.metrics.IncError(c.topic, "unmarshal")

			if err := c.reader.CommitMessages(ctx, message); err != nil {
				c.metrics.IncError(c.topic, "commit")
				c.metrics.ObserveConsumed(c.topic, "error", startedAt)
				return fmt.Errorf("commit invalid price check message: %w", err)
			}

			c.metrics.ObserveConsumed(c.topic, "invalid", startedAt)
			continue
		}

		if err := c.handler.ProcessCheckPrice(ctx, event); err != nil {
			c.logger.Error(
				"failed to process price check event",
				zap.String("task_id", event.TaskID),
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
			return fmt.Errorf("commit price check message: %w", err)
		}

		c.metrics.ObserveConsumed(c.topic, "success", startedAt)
	}
}

func (c *PriceCheckConsumer) Close() error {
	return c.reader.Close()
}
