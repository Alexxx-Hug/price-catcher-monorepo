package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	eventdto "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/models/eventdto"
	"github.com/segmentio/kafka-go"
)

type KafkaPriceChangedProducer struct {
	writer *kafka.Writer
}

func NewKafkaPriceChangedProducer(brokers []string, topic string) *KafkaPriceChangedProducer {
	return &KafkaPriceChangedProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *KafkaPriceChangedProducer) SendProductPriceChanged(ctx context.Context, event eventdto.ProductPriceChangedEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal product price changed event: %w", err)
	}

	key := strconv.FormatInt(event.TelegramUserID, 10)
	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  event.ChangedAt,
	}); err != nil {
		return fmt.Errorf("write product price changed event: %w", err)
	}

	return nil
}

func (p *KafkaPriceChangedProducer) Close() error {
	return p.writer.Close()
}
