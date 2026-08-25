package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/metrics"
	eventdto "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/models/eventdto"
	"github.com/segmentio/kafka-go"
)

type KafkaPriceChangedProducer struct {
	writer  *kafka.Writer
	topic   string
	metrics *metrics.Metrics
}

func NewKafkaPriceChangedProducer(brokers []string, topic string, appMetrics *metrics.Metrics) *KafkaPriceChangedProducer {
	return &KafkaPriceChangedProducer{
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

func (p *KafkaPriceChangedProducer) SendProductPriceChanged(ctx context.Context, event eventdto.ProductPriceChangedEvent) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal product price changed event: %w", err)
	}

	key := strconv.FormatInt(event.TelegramUserID, 10)
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  event.ChangedAt,
	})
	p.metrics.IncProduced(p.topic, err)
	if err != nil {
		return fmt.Errorf("write product price changed event: %w", err)
	}

	return nil
}

func (p *KafkaPriceChangedProducer) Close() error {
	return p.writer.Close()
}
