package config

import (
	"fmt"
	"strconv"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port        int    `env:"PORT" envDefault:"8080"`
	RabbitMQURL string `env:"RABBITMQ_URL"`
}

func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("unable to parse environment variables : %w", err)
	}
	return cfg, nil
}
