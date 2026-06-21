package main

import (
	"context"
	"log"

	"example.com/pz13-rabbitmq/internal/env"
	"example.com/pz13-rabbitmq/services/worker/internal/consumer"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitURL := env.Get("RABBIT_URL", "amqp://guest:guest@localhost:5672/")
	queueName := env.Get("QUEUE_NAME", "task_events")

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("rabbit connect error: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel error: %v", err)
	}
	defer ch.Close()

	if err := consumer.Run(context.Background(), ch, queueName, log.Default()); err != nil {
		log.Fatalf("consumer error: %v", err)
	}
}
