package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Env     string `env:"APP_ENV"   envDefault:"local"`
	AppPort string `env:"APP_PORT" envDefault:":8080"`

	PostgresContainerHost string `env:"POSTGRES_CONTAINER_HOST"`
	PostgresContainerPort int    `env:"POSTGRES_CONTAINER_PORT"`
	PostgresUser          string `env:"POSTGRES_USER"`
	PostgresPassword      string `env:"POSTGRES_PASSWORD"`
	PostgresDB            string `env:"POSTGRES_DB"`
	RunMigrations         bool   `env:"RUN_MIGRATIONS" envDefault:"false"`

	WeatherApiKey string `env:"WEATHER_API_KEY"`
}

func NewConfigFromEnv() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config from env: %w", err)
	}
	return cfg, nil
}
