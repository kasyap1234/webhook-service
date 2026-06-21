package queue

import (
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

type Worker struct {
	consumer *rabbitmq.Consumer
}

func NewWorker(conn *rabbitmq.Conn) (*Worker, error) {
	consumer, err := rabbitmq.NewConsumer(conn, "webhooks",
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsExchangeKind("direct"),
		rabbitmq.WithConsumerOptionsExchangeDurable,
		rabbitmq.WithConsumerOptionsLogging)
	if err != nil {
		return nil, err
	}

	return &Worker{consumer: consumer}, nil
}

func handler(d rabbitmq.Delivery) rabbitmq.Action {
	// Process delivery and return action
	return rabbitmq.ActionAck
}
