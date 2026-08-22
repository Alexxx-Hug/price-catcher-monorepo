package config

import (
	"os"
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

type Config struct {
	App  AppConfig
	GRPC GRPCConfig
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
