package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	eventdto "github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/models/eventDTO"
	"github.com/segmentio/kafka-go"
)

type ProductCheckedProducer struct {
	writer *kafka.Writer
}

func NewProductCheckedProducer(topic string, brokers []string) *ProductCheckedProducer {
	return &ProductCheckedProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			RequiredAcks: kafka.RequireAll,
			Balancer:     &kafka.Hash{},
		},
	}
}

func (p *ProductCheckedProducer) SendProductChecked(ctx context.Context, event eventdto.ProductCheckedEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal product checked event: %w", err)
	}

	key := strconv.FormatInt(event.ProductSizeID, 10)
	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  event.CheckedAt,
	}); err != nil {
		return fmt.Errorf("write product checked event: %w", err)
	}

	return nil
}

func (p *ProductCheckedProducer) Close() error {
	return p.writer.Close()
}
