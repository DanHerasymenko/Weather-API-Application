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

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks that all required configuration values are present
func (cfg *Config) Validate() error {
	if cfg.WeatherApiKey == "" {
		return fmt.Errorf("WEATHER_API_KEY is required")
	}
	if cfg.BaseURL == "" {
		return fmt.Errorf("APP_BASE_URL is required")
	}
	if cfg.EmailClientFrom == "" {
		return fmt.Errorf("SMTP_FROM is required")
	}
	if cfg.EmailClientPassword == "" {
		return fmt.Errorf("SMTP_PASSWORD is required")
	}
	if cfg.EmailClientHost == "" {
		return fmt.Errorf("SMTP_HOST is required")
	}
	if cfg.EmailClientPort == "" {
		return fmt.Errorf("SMTP_PORT is required")
	}
	if cfg.PostgresContainerHost == "" {
		return fmt.Errorf("POSTGRES_CONTAINER_HOST is required")
	}
	if cfg.PostgresUser == "" {
		return fmt.Errorf("POSTGRES_USER is required")
	}
	if cfg.PostgresPassword == "" {
		return fmt.Errorf("POSTGRES_PASSWORD is required")
	}
	if cfg.PostgresDB == "" {
		return fmt.Errorf("POSTGRES_DB is required")
	}
	return nil
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
