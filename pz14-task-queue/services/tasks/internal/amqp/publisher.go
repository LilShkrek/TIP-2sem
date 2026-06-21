package amqpqueue

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"

	commonjobs "pz14-task-queue/internal/jobs"
)

type Publisher struct {
	ch *amqp.Channel
}

func NewPublisher(ch *amqp.Channel) *Publisher {
	return &Publisher{ch: ch}
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

func (p *Publisher) PublishJob(ctx context.Context, queue string, job commonjobs.TaskJob) error {
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
