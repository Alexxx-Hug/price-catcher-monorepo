package providers

import (
	"fmt"

	produceradapter "github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/adapters/producer"
	"github.com/Alexxx-Hug/price-catcher-monorepo/product-store/internal/config"
)

type KafkaProvider struct {
	PriceCheckTaskProducer *produceradapter.KafkaPriceCheckTaskProducer
}

func NewKafkaProvider(cfg config.KafkaConfig) (*KafkaProvider, error) {
	if len(cfg.BrokerList()) == 0 {
		return nil, fmt.Errorf("kafka brokers are not configured")
	}

	if cfg.TaskCheckPricesTopic == "" {
		return nil, fmt.Errorf("task check prices topic is not configured")
	}

	return &KafkaProvider{
		PriceCheckTaskProducer: produceradapter.NewKafkaPriceCheckTaskProducer(
			cfg.BrokerList(),
			cfg.TaskCheckPricesTopic,
		),
	}, nil
}

func (p *KafkaProvider) Close() error {
	if p == nil || p.PriceCheckTaskProducer == nil {
		return nil
	}

	return p.PriceCheckTaskProducer.Close()
}
