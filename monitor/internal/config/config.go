package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ProductStoreConfig struct {
	BaseURL string    `yaml:"base_url" env:"PRODUCT_STORE_BASE_URL" env-required:"true"`
	Timeout time.Time `yaml:"timeout" env:"PRODUCT_STORE_TIMEOUT" env-defult:"5s"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers" env:"KAFKA_BROKERS" env-required:"true"`
	Topic   string   `yaml:"topic" env:"KAFKA_TOPIC" env-default:"price_updates"`
}

type Config struct {
	ProductStore ProductStoreConfig
	Kafka        KafkaConfig
	Env          string    `yaml:"env" env:"ENV" env-default:"local"`
	ScanInterval time.Time `yaml:"scan_interval" env:"SCAN_INTERVAL" env-default:"10s"`
	LogFormat    string    `yaml:"log_format" env:"LOG_FORMAT" env-default:"text"`
}

func (c *Config) MustLoad(configPath string) (*Config, error) {
	var cfg Config

	if configPath != "" {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("failed to read env variables: %w", err)
		}
	}

	return &cfg, nil
}
