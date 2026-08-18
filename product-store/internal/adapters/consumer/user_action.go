package consumer

import (
	"context"
	"encoding/json"
	"fmt"

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
}

func NewUserActionConsumer(
	brokers []string,
	topic string,
	groupID string,
	handler UserActionHandler,
	logger *zap.Logger,
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
	}
}

func (c *UserActionConsumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return fmt.Errorf("fetch user action message: %w", err)
		}

		var event eventdto.UserActionEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			c.logger.Error("failed to unmarshal user action event", zap.Error(err))

			if err := c.reader.CommitMessages(ctx, message); err != nil {
				return fmt.Errorf("commit user action message: %w", err)
			}

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
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			return fmt.Errorf("commit user action message: %w", err)
		}
	}
}

func (c *UserActionConsumer) Close() error {
	return c.reader.Close()
}
