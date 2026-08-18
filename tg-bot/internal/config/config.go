package config

import (
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Name    string `env:"APP_NAME" env-default:"tg-bot"`
	Version string `env:"APP_VERSION" env-default:"1.0.0"`
}

type TelegramConfig struct {
	Token string `env:"TELEGRAM_BOT_TOKEN" env-required:"true"`
}

type ProductStoreGRPCConfig struct{
	Address string `env:"PRODUCT_STORE_GRPC_ADDRESS" env-default:"localhost:50051"`
	Timeout time.Duration `env:"PRODUCT_STORE_GRPC_TIMEOUT" env-default:"5s"`
}

type KafkaConfig struct {
	Brokers          string `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	UserActionsTopic string `env:"KAFKA_TOPIC_USER_ACTIONS" env-default:"user-actions"`
}

func (c KafkaConfig) BrokerList() []string {
	parts := strings.Split(c.Brokers, ",")
	brokers := make([]string, 0, len(parts))

	for _, broker := range parts {
		broker = strings.TrimSpace(broker)
		if broker != "" {
			brokers = append(brokers, broker)
		}
	}

	return brokers
}

type Config struct {
	App      AppConfig
	Telegram TelegramConfig
	Kafka    KafkaConfig
	ProductStoreGRPCConfig ProductStoreGRPCConfig

}

func MustLoad() *Config {
	cfg := &Config{}

	if _, err := os.Stat(".env"); err == nil {
		if err := cleanenv.ReadConfig(".env", cfg); err != nil {
			panic("cannot read config from .env: " + err.Error())
		}

		return cfg
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return cfg
}
