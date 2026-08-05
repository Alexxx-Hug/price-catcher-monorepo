package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	eventdto "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/models/eventdto"
	"github.com/segmentio/kafka-go"
)

type KafkaPriceCheckTaskProducer struct {
	writer *kafka.Writer
}

func NewKafkaPriceCheckTaskProducer(brokers []string, topic string) *KafkaPriceCheckTaskProducer {
	return &KafkaPriceCheckTaskProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *KafkaPriceCheckTaskProducer) SendPriceCheckTask(ctx context.Context, event eventdto.TaskCheckPricesEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal task check prices event: %w", err)
	}

	key := strconv.FormatInt(event.ProductSizeID, 10)

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  event.RequestedAt,
	})

	if err != nil {
		return fmt.Errorf("write check task prices event: %w", err)
	}

	return nil
}

func (p *KafkaPriceCheckTaskProducer) Close() error {
	return p.writer.Close()
}
