package providers

import (
	"fmt"

	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/adapters/producer"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/config"
	"github.com/Alexxx-Hug/price-catcher-monorepo/monitor/internal/metrics"
)

type KafkaProvider struct {
	ProductCheckedProducer *producer.ProductCheckedProducer
}

func NewKafkaProvider(cfg config.KafkaConfig, appMetrics *metrics.Metrics) (*KafkaProvider, error) {
	if len(cfg.BrokerList()) == 0 {
		return nil, fmt.Errorf("kafka brokers are not configured")
	}

	if cfg.TaskCheckPricesTopic == "" {
		return nil, fmt.Errorf("task check prices topic is not configured")
	}

	return &KafkaProvider{
		ProductCheckedProducer: producer.NewProductCheckedProducer(
			cfg.ProductCheckedTopic,
			cfg.BrokerList(),
			appMetrics,
		),
	}, nil
}

func (p *KafkaProvider) Close() error {
	if p == nil {
		return nil
	}

	if p.ProductCheckedProducer != nil {
		if err := p.ProductCheckedProducer.Close(); err != nil {
			return err
		}
	}

	return nil
}
