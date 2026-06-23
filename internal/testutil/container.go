//go:build integration

package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp091 "github.com/rabbitmq/amqp091-go"
	gorabbitmq "github.com/wagslane/go-rabbitmq"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcrabbitmq "github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestDeps struct {
	Pool       *pgxpool.Pool
	RabbitConn *gorabbitmq.Conn
	AmqpConn   *amqp091.Connection
	DBURL      string
	RabbitURL  string
}

func Setup(t *testing.T) *TestDeps {
	t.Helper()
	ctx := context.Background()

	pgContainer, pgURL := startPostgres(t, ctx)
	rabbitContainer, rabbitURL := startRabbitMQ(t, ctx)

	pool, err := pgxpool.New(ctx, pgURL)
	if err != nil {
		t.Fatalf("failed to create pgxpool: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	runMigrations(t, pgURL)

	rabbitConn, err := gorabbitmq.NewConn(rabbitURL, gorabbitmq.WithConnectionOptionsLogging)
	if err != nil {
		t.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	t.Cleanup(func() { rabbitConn.Close() })

	amqpConn, err := amqp091.Dial(rabbitURL)
	if err != nil {
		t.Fatalf("failed to dial AMQP: %v", err)
	}
	t.Cleanup(func() { amqpConn.Close() })

	t.Cleanup(func() {
		_ = pgContainer.Terminate(ctx)
		_ = rabbitContainer.Terminate(ctx)
	})

	return &TestDeps{
		Pool:       pool,
		RabbitConn: rabbitConn,
		AmqpConn:   amqpConn,
		DBURL:      pgURL,
		RabbitURL:  rabbitURL,
	}
}

func startPostgres(t *testing.T, ctx context.Context) (testcontainers.Container, string) {
	t.Helper()

	dbName := "webhookdb"
	dbUser := "webhook"
	dbPass := "webhooksecret"

	container, err := tcpostgres.Run(ctx,
		"postgres:18-alpine",
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(dbUser),
		tcpostgres.WithPassword(dbPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
		testcontainers.WithReuseByName("webhook-test-postgres"),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, host, port.Port(), dbName)

	return container, dbURL
}

func startRabbitMQ(t *testing.T, ctx context.Context) (testcontainers.Container, string) {
	t.Helper()

	user := "webhook"
	pass := "rabbitsecret"

	container, err := tcrabbitmq.Run(ctx,
		"rabbitmq:4-management-alpine",
		tcrabbitmq.WithAdminUsername(user),
		tcrabbitmq.WithAdminPassword(pass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Server startup complete"),
		),
		testcontainers.WithReuseByName("webhook-test-rabbitmq"),
	)
	if err != nil {
		t.Fatalf("failed to start rabbitmq container: %v", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5672")
	rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port.Port())

	return container, rabbitURL
}

func runMigrations(t *testing.T, dbURL string) {
	t.Helper()

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		t.Fatalf("failed to open db for migrations: %v", err)
	}
	defer db.Close()

	goose.SetDialect("postgres")

	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "migrations")

	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}
}

func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}
