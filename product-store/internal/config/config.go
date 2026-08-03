package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Name    string `env:"APP_NAME" env-default:"test-shop"`
	Version string `env:"APP_VERSION" env-default:"1.0.0"`
}

type HTTPConfig struct {
	Port string `env:"SERVER_PORT" env-default:"8080"`
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

type Config struct {
	App  AppConfig
	HTTP HTTPConfig
	DB   DBConfig
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
