package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/metrics"
	eventdto "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/models/eventdto"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type ProductCheckedHandler interface {
	ProcessCheckedProduct(ctx context.Context, event eventdto.ProductCheckedEvent) error
}

type DeadLetterProducer interface {
	SendDeadLetter(
		ctx context.Context,
		sourceTopic string,
		key []byte,
		value []byte,
		partition int,
		offset int64,
		reason string,
	) error
}

type ProductCheckedConsumer struct {
	reader             *kafka.Reader
	topic              string
	handler            ProductCheckedHandler
	deadLetterProducer DeadLetterProducer
	logger             *zap.Logger
	metrics            *metrics.Metrics
}

func NewProductCheckedConsumer(
	brokers []string,
	topic string,
	groupID string,
	handler ProductCheckedHandler,
	deadLetterProducer DeadLetterProducer,
	logger *zap.Logger,
	appMetrics *metrics.Metrics,
) *ProductCheckedConsumer {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &ProductCheckedConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Topic:   topic,
			GroupID: groupID,
			Brokers: brokers,
		}),
		topic:              topic,
		handler:            handler,
		deadLetterProducer: deadLetterProducer,
		logger:             logger,
		metrics:            appMetrics,
	}
}

func (c *ProductCheckedConsumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			c.metrics.IncError(c.topic, "fetch")
			return fmt.Errorf("fetch product checked message: %w", err)
		}

		startedAt := time.Now()
		var event eventdto.ProductCheckedEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			c.logger.Error("failed to unmarshal product checked event", zap.Error(err))
			c.metrics.IncError(c.topic, "unmarshal")

			if c.deadLetterProducer != nil {
				if dlqErr := c.deadLetterProducer.SendDeadLetter(
					ctx,
					c.topic,
					message.Key,
					message.Value,
					message.Partition,
					message.Offset,
					err.Error(),
				); dlqErr != nil {
					c.metrics.ObserveConsumed(c.topic, "error", startedAt)
					return fmt.Errorf("send product checked message to dlq: %w", dlqErr)
				}
			}

			if err := c.reader.CommitMessages(ctx, message); err != nil {
				c.metrics.IncError(c.topic, "commit")
				c.metrics.ObserveConsumed(c.topic, "error", startedAt)
				return fmt.Errorf("commit invalid product checked message: %w", err)
			}

			c.metrics.ObserveConsumed(c.topic, "invalid", startedAt)
			continue
		}

		if err := c.handler.ProcessCheckedProduct(ctx, event); err != nil {
			c.logger.Error(
				"failed to process product checked event",
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
			return fmt.Errorf("commit product checked message: %w", err)
		}

		c.metrics.ObserveConsumed(c.topic, "success", startedAt)
	}
}

func (c *ProductCheckedConsumer) Close() error {
	return c.reader.Close()
}
