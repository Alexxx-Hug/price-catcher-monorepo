package providers

import (
	"fmt"

	producer "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/adapters/producer"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/config"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/metrics"
)

type KafkaProvider struct {
	PriceCheckTaskProducer *producer.KafkaPriceCheckTaskProducer
	PriceChangedProducer   *producer.KafkaPriceChangedProducer
	DeadLetterProducer     *producer.KafkaDeadLetterProducer
}

func NewKafkaProvider(cfg config.KafkaConfig, appMetrics *metrics.Metrics) (*KafkaProvider, error) {
	if len(cfg.BrokerList()) == 0 {
		return nil, fmt.Errorf("kafka brokers are not configured")
	}

	if cfg.TaskCheckPricesTopic == "" {
		return nil, fmt.Errorf("task check prices topic is not configured")
	}

	if cfg.ProductCheckedDLQTopic == "" {
		return nil, fmt.Errorf("product checked dlq topic is not configured")
	}

	if cfg.ProductPriceChangedTopic == "" {
		return nil, fmt.Errorf("product price changed topic is not configured")
	}

	return &KafkaProvider{
		PriceCheckTaskProducer: producer.NewKafkaPriceCheckTaskProducer(
			cfg.BrokerList(),
			cfg.TaskCheckPricesTopic,
			appMetrics,
		),
		DeadLetterProducer: producer.NewKafkaDeadLetterProducer(
			cfg.BrokerList(),
			cfg.ProductCheckedDLQTopic,
			appMetrics,
		),
		PriceChangedProducer: producer.NewKafkaPriceChangedProducer(
			cfg.BrokerList(),
			cfg.ProductPriceChangedTopic,
			appMetrics,
		),
	}, nil
}

func (p *KafkaProvider) Close() error {
	if p == nil {
		return nil
	}

	if p.PriceCheckTaskProducer != nil {
		if err := p.PriceCheckTaskProducer.Close(); err != nil {
			return err
		}
	}

	if p.DeadLetterProducer != nil {
		if err := p.DeadLetterProducer.Close(); err != nil {
			return err
		}
	}

	if p.PriceChangedProducer != nil {
		if err := p.PriceChangedProducer.Close(); err != nil {
			return err
		}
	}

	return nil
}
