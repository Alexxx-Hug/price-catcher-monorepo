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

type KafkaPriceCheckTaskProducer struct {
	writer  *kafka.Writer
	topic   string
	metrics *metrics.Metrics
}

func NewKafkaPriceCheckTaskProducer(brokers []string, topic string, appMetrics *metrics.Metrics) *KafkaPriceCheckTaskProducer {
	return &KafkaPriceCheckTaskProducer{
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
	p.metrics.IncProduced(p.topic, err)

	if err != nil {
		return fmt.Errorf("write check task prices event: %w", err)
	}

	return nil
}

func (p *KafkaPriceCheckTaskProducer) Close() error {
	return p.writer.Close()
}
