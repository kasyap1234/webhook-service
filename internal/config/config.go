package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port        int    `env:"PORT" envDefault:"8080"`
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://localhost:5432/webhooks?sslmode=disable"`
	RabbitMQURL string `env:"RABBITMQ_URL"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("unable to parse environment variables : %w", err)
	}
	return cfg, nil
}
