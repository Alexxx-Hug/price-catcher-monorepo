package consumer

import (
	"context"
	"encoding/json"
	"fmt"

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
}

func NewPriceCheckConsumer(brokers []string, topic string, groupID string, handler PriceCheckHandler, logger *zap.Logger) *PriceCheckConsumer {
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
	}
}

func (c *PriceCheckConsumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch price check message: %w", err)
		}

		var event eventdto.TaskCheckPricesEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			c.logger.Error("failed to unmarshal price check event", zap.Error(err))

			if err := c.reader.CommitMessages(ctx, message); err != nil {
				return fmt.Errorf("commit invalid price check message: %w", err)
			}

			continue
		}

		if err := c.handler.ProcessCheckPrice(ctx, event); err != nil {
			c.logger.Error(
				"failed to process price check event",
				zap.String("task_id", event.TaskID),
				zap.Int64("product_size_id", event.ProductSizeID),
				zap.Error(err),
			)
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			return fmt.Errorf("commit price check message: %w", err)
		}
	}
}

func (c *PriceCheckConsumer) Close() error {
	return c.reader.Close()
}
