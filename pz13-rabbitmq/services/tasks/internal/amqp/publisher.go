package amqppublisher

import (
	"context"
	"encoding/json"

	"example.com/pz13-rabbitmq/internal/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Channel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

type Publisher struct {
	channel   Channel
	queueName string
}

func New(channel Channel, queueName string) *Publisher {
	return &Publisher{
		channel:   channel,
		queueName: queueName,
	}
}

func (p *Publisher) PublishTaskCreated(ctx context.Context, event events.TaskEvent) error {
	if _, err := p.channel.QueueDeclare(p.queueName, true, false, false, false, nil); err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(ctx, "", p.queueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

type Connection interface {
	Channel() (*amqp.Channel, error)
	Close() error
}

type DialFunc func(url string) (Connection, error)

type DialingPublisher struct {
	rabbitURL string
	queueName string
	dial      DialFunc
}

func NewDialing(rabbitURL, queueName string) *DialingPublisher {
	return &DialingPublisher{
		rabbitURL: rabbitURL,
		queueName: queueName,
		dial: func(url string) (Connection, error) {
			return amqp.Dial(url)
		},
	}
}

func (p *DialingPublisher) PublishTaskCreated(ctx context.Context, event events.TaskEvent) error {
	conn, err := p.dial(p.rabbitURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return New(ch, p.queueName).PublishTaskCreated(ctx, event)
}
