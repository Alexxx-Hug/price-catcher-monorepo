package consumer

import (
	"context"
	"encoding/json"
	"fmt"

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
}

func NewProductCheckedConsumer(
	brokers []string,
	topic string,
	groupID string,
	handler ProductCheckedHandler,
	deadLetterProducer DeadLetterProducer,
	logger *zap.Logger,
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
	}
}

func (c *ProductCheckedConsumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return fmt.Errorf("fetch product checked message: %w", err)
		}

		var event eventdto.ProductCheckedEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			c.logger.Error("failed to unmarshal product checked event", zap.Error(err))

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
					return fmt.Errorf("send product checked message to dlq: %w", dlqErr)
				}
			}

			if err := c.reader.CommitMessages(ctx, message); err != nil {
				return fmt.Errorf("commit invalid product checked message: %w", err)
			}

			continue
		}

		if err := c.handler.ProcessCheckedProduct(ctx, event); err != nil {
			c.logger.Error(
				"failed to process product checked event",
				zap.String("task_id", event.TaskID),
				zap.Int64("product_size_id", event.ProductSizeID),
				zap.Error(err),
			)
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			return fmt.Errorf("commit product checked message: %w", err)
		}
	}
}

func (c *ProductCheckedConsumer) Close() error {
	return c.reader.Close()
}
