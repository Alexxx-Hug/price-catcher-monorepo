package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models/events"
	"github.com/segmentio/kafka-go"
)

type UserActionProducer struct {
	writer *kafka.Writer
}

func NewUserActionProducer(topic string, brokers []string) *UserActionProducer {
	return &UserActionProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *UserActionProducer) SendUserAction(ctx context.Context, event events.UserActionEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	key := strconv.FormatInt(event.TelegramUserID, 10)

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  event.CreatedAt,
	}); err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	return nil
}

func (p *UserActionProducer) Close() error {
	return p.writer.Close()
}
