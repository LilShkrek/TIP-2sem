package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pz14-task-queue/internal/env"
	"pz14-task-queue/services/worker/internal/consumer"
	"pz14-task-queue/services/worker/internal/store"
)

func main() {
	rabbitURL := env.String("RABBIT_URL", "amqp://guest:guest@localhost:5672/")
	taskQueue := env.String("TASK_QUEUE", "task_jobs")
	dlqQueue := env.String("DLQ_QUEUE", "task_jobs_dlq")
	maxAttempts := env.Int("MAX_ATTEMPTS", 3)

	logger := log.New(os.Stdout, "worker ", log.LstdFlags)

	conn, ch, err := consumer.Dial(rabbitURL)
	if err != nil {
		logger.Fatalf("connect rabbitmq: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	if err := consumer.DeclareQueues(ch, taskQueue, dlqQueue); err != nil {
		logger.Fatalf("declare queues: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	publisher := consumer.NewAMQPPublisher(ch)
	processed := store.NewProcessedStore()
	handler := consumer.NewHandler(publisher, processed, taskQueue, dlqQueue, maxAttempts, 300*time.Millisecond, logger)

	logger.Printf("worker consuming queue=%s dlq=%s max_attempts=%d", taskQueue, dlqQueue, maxAttempts)
	if err := consumer.Consume(ctx, ch, taskQueue, handler, logger); err != nil && !errors.Is(err, context.Canceled) {
		logger.Fatalf("consume: %v", err)
	}
}
