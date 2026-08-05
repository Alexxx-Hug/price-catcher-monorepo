package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Name    string `env:"APP_NAME" env-default:"test-shop"`
	Version string `env:"APP_VERSION" env-default:"1.0.0"`
}

type HTTPConfig struct {
	Port                   string        `env:"SERVER_PORT" env-default:"8080"`
	ReadinessCheckInterval time.Duration `env:"READINESS_CHECK_INTERVAL" env-default:"10s"`
	ReadinessCheckTimeout  time.Duration `env:"READINESS_CHECK_TIMEOUT" env-default:"2s"`
}

type DBConfig struct {
	Port         string `env:"DB_PORT" env-default:"5432"`
	Host         string `env:"DB_HOST" env-required:"true"`
	User         string `env:"DB_USER" env-required:"true"`
	Password     string `env:"DB_PASSWORD" env-required:"true"`
	Name         string `env:"DB_NAME" env-required:"true"`
	SSLMode      string `env:"DB_SSLMODE" env-default:"disable"`
	PoolMaxConns int    `env:"DB_MAX_POOL_CONNS" env-default:"30"`
	PoolMinConns int    `env:"DB_MIN_POOL_CONNS" env-default:"5"`
}

type KafkaConfig struct {
	Brokers                  string `env:"KAFKA_BROKERS" env-default:"localhost:9092"`
	TaskCheckPricesTopic     string `env:"KAFKA_TOPIC_TASK_CHECK_PRICES" env-default:"task-check-prices"`
	ProductCheckedTopic      string `env:"KAFKA_TOPIC_PRODUCT_CHECKED" env-default:"product-checked"`
	ProductPriceChangedTopic string `env:"KAFKA_TOPIC_PRODUCT_PRICE_CHANGED" env-default:"product-price-changed"`
	UserActionsTopic         string `env:"KAFKA_TOPIC_USER_ACTIONS" env-default:"user-actions"`
	GroupID                  string `env:"KAFKA_GROUP_ID" env-default:"product-store"`
}

type MonitoringConfig struct {
	PriceCheckInterval   time.Duration `env:"PRICE_CHECK_INTERVAL" env-default:"5m"`
	PriceCheckBatchLimit int           `env:"PRICE_CHECK_BATCH_LIMIT" env-default:"5"`
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
	App        AppConfig
	HTTP       HTTPConfig
	Kafka      KafkaConfig
	Monitoring MonitoringConfig
	DB         DBConfig
}

func (c *DBConfig) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&pool_max_conns=%d&pool_min_conns=%d",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode, c.PoolMaxConns, c.PoolMinConns,
	)
}

// githab.com/joho/godotenv
// _ = godotenv.Load()
// val, err := os.GetEnv(key)
// написать Лене и спросить как сделана работа с конфигом на шансах/ии-прокси
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
