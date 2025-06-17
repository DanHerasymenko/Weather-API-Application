package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Env            string `env:"APP_ENV"   envDefault:"local"`
	AppPort        string `env:"APP_PORT" envDefault:":8080"`
	BaseURL        string `env:"APP_BASE_URL"`
	DailyStartHour int    `env:"DAILY_START_HOUR" envDefault:"8"`

	PostgresContainerHost string `env:"POSTGRES_CONTAINER_HOST"`
	PostgresContainerPort int    `env:"POSTGRES_CONTAINER_PORT"`
	PostgresUser          string `env:"POSTGRES_USER"`
	PostgresPassword      string `env:"POSTGRES_PASSWORD"`
	PostgresDB            string `env:"POSTGRES_DB"`

	WeatherApiKey string `env:"WEATHER_API_KEY"`

	EmailClientFrom     string `env:"SMTP_FROM"`
	EmailClientPassword string `env:"SMTP_PASSWORD"`
	EmailClientHost     string `env:"SMTP_HOST"`
	EmailClientPort     string `env:"SMTP_PORT"`
}

// NewConfigFromEnv creates a new Config instance and populates it with values from environment variables.
func NewConfigFromEnv() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config from env: %w", err)
	}
	return cfg, nil
}

func (cfg *Config) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresContainerHost,
		cfg.PostgresContainerPort,
		cfg.PostgresDB,
	)
}
