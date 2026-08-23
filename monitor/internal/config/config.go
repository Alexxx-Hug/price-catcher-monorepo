package config

import (
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Name    string `env:"APP_NAME" env-default:"monitor"`
	Version string `env:"APP_VERSION" env-default:"1.0.0"`
}

type GRPCConfig struct {
	Port    string        `env:"GRPC_PORT" env-default:"50052"`
	Timeout time.Duration `env:"GRPC_TIMEOUT" env-default:"5s"`
}

type KafkaConfig struct {
	Brokers              string `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	GroupID              string `env:"KAFKA_GROUP_ID" env-default:"monitor"`
	TaskCheckPricesTopic string `env:"KAFKA_TOPIC_TASK_CHECK_PRICES" env-default:"task-check-prices"`
	ProductCheckedTopic  string `env:"KAFKA_TOPIC_PRODUCT_CHECKED" env-default:"product-checked"`
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
	App   AppConfig
	GRPC  GRPCConfig
	Kafka KafkaConfig
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
