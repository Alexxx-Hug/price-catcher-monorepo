package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaDeadLetterProducer struct {
	writer *kafka.Writer
}

type deadLetterMessage struct {
	SourceTopic string    `json:"source_topic"`
	Partition   int       `json:"partition"`
	Offset      int64     `json:"offset"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Reason      string    `json:"reason"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func NewKafkaDeadLetterProducer(brokers []string, topic string) *KafkaDeadLetterProducer {
	return &KafkaDeadLetterProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *KafkaDeadLetterProducer) SendDeadLetter(
	ctx context.Context,
	sourceTopic string,
	key []byte,
	value []byte,
	partition int,
	offset int64,
	reason string,
) error {
	message := deadLetterMessage{
		SourceTopic: sourceTopic,
		Partition:   partition,
		Offset:      offset,
		Key:         string(key),
		Value:       string(value),
		Reason:      reason,
		OccurredAt:  time.Now(),
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal dead letter message: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(offset, 10)),
		Value: payload,
		Time:  message.OccurredAt,
	})
	if err != nil {
		return fmt.Errorf("write dead letter message: %w", err)
	}

	return nil
}

func (p *KafkaDeadLetterProducer) Close() error {
	return p.writer.Close()
}
