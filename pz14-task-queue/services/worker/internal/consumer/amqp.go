package consumer

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	commonjobs "pz14-task-queue/internal/jobs"
)

type AMQPPublisher struct {
	ch *amqp.Channel
}

func NewAMQPPublisher(ch *amqp.Channel) *AMQPPublisher {
	return &AMQPPublisher{ch: ch}
}

func Dial(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}

func DeclareQueues(ch *amqp.Channel, taskQueue, dlqQueue string) error {
	for _, queue := range []string{taskQueue, dlqQueue} {
		if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
			return err
		}
	}
	return nil
}

func (p *AMQPPublisher) PublishJob(ctx context.Context, queue string, job commonjobs.TaskJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx, "", queue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    job.MessageID,
		Body:         body,
	})
}

func Consume(ctx context.Context, ch *amqp.Channel, queue string, handler *Handler, logger *log.Logger) error {
	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}

	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return nil
			}

			msg := Message{
				Body: delivery.Body,
				Ack: func() error {
					return delivery.Ack(false)
				},
				Nack: func(requeue bool) error {
					return delivery.Nack(false, requeue)
				},
			}

			if err := handler.Handle(ctx, msg); err != nil && logger != nil {
				logger.Printf("handle message: %v", err)
			}
		}
	}
}
