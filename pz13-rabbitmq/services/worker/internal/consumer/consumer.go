package consumer

import (
	"context"
	"encoding/json"
	"log"

	"example.com/pz13-rabbitmq/internal/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Channel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	Qos(prefetchCount, prefetchSize int, global bool) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
}

func Run(ctx context.Context, ch Channel, queueName string, logger *log.Logger) error {
	if logger == nil {
		logger = log.Default()
	}

	if _, err := ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return err
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}

	deliveries, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	logger.Println("worker started, waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}
			HandleDelivery(delivery, logger)
		}
	}
}

func HandleDelivery(delivery amqp.Delivery, logger *log.Logger) {
	if logger == nil {
		logger = log.Default()
	}

	var event events.TaskEvent
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		logger.Printf("bad message: %v", err)
		if err := delivery.Nack(false, false); err != nil {
			logger.Printf("nack error: %v", err)
		}
		return
	}

	logger.Printf(
		"received event=%s task_id=%s ts=%s request_id=%s",
		event.Event,
		event.TaskID,
		event.TS,
		event.RequestID,
	)

	if err := delivery.Ack(false); err != nil {
		logger.Printf("ack error: %v", err)
	}
}
