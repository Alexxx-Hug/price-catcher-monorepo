package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/metrics"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models/events"
	"github.com/segmentio/kafka-go"
)

type UserActionProducer struct {
	writer  *kafka.Writer
	topic   string
	metrics *metrics.Metrics
}

func NewUserActionProducer(topic string, brokers []string, appMetrics *metrics.Metrics) *UserActionProducer {
	return &UserActionProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
		},
		topic:   topic,
		metrics: appMetrics,
	}
}

func (p *UserActionProducer) SendUserAction(ctx context.Context, event events.UserActionEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	key := strconv.FormatInt(event.TelegramUserID, 10)

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  event.CreatedAt,
	})
	p.metrics.IncProduced(p.topic, err)
	if err != nil {
		return fmt.Errorf("failed to send event: %w", err)
	}

	return nil
}

func (p *UserActionProducer) Close() error {
	return p.writer.Close()
}
