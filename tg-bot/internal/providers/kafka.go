package providers

import (
	"fmt"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/adapters/producer"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/config"
)

type KafkaProvider struct {
	UserActionProducer *producer.UserActionProducer
}

func NewKafkaProvider(cfg config.KafkaConfig) (*KafkaProvider, error) {
	if len(cfg.BrokerList()) == 0 {
		return nil, fmt.Errorf("kafka brokers are not configured")
	}

	if cfg.UserActionsTopic == "" {
		return nil, fmt.Errorf("user actions topic is not configured")
	}

	return &KafkaProvider{
		UserActionProducer: producer.NewUserActionProducer(
			cfg.UserActionsTopic,
			cfg.BrokerList(),
		),
	}, nil
}

func (p *KafkaProvider) Close() error {
	if p == nil {
		return nil
	}

	if p.UserActionProducer != nil {
		err := p.UserActionProducer.Close()
		return err
	}

	return nil
}
