package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("RABBITMQ_URL")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}
	if cfg.DatabaseURL == "" {
		t.Error("expected non-empty default DatabaseURL")
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb?sslmode=disable")
	os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("RABBITMQ_URL")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://user:pass@localhost:5432/testdb?sslmode=disable" {
		t.Errorf("unexpected DatabaseURL: %s", cfg.DatabaseURL)
	}
	if cfg.RabbitMQURL != "amqp://guest:guest@localhost:5672/" {
		t.Errorf("unexpected RabbitMQURL: %s", cfg.RabbitMQURL)
	}
}

func TestLoadConfig_EmptyRabbitMQ(t *testing.T) {
	os.Unsetenv("RABBITMQ_URL")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RabbitMQURL != "" {
		t.Errorf("expected empty RabbitMQURL, got %s", cfg.RabbitMQURL)
	}
}
