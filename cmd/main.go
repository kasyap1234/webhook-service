package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/kasyap1234/webhook-service/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application := app.NewApp()
	defer application.Close()

	go application.Run()

	<-ctx.Done()
	log.Println("shutting down gracefully...")
}
