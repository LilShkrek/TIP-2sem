package amqppublisher

import (
	"context"
	"encoding/json"
	"testing"

	"example.com/pz13-rabbitmq/internal/events"
	amqp "github.com/rabbitmq/amqp091-go"
)

type fakeChannel struct {
	queueName    string
	durable      bool
	publishedKey string
	published    amqp.Publishing
}

func (ch *fakeChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	ch.queueName = name
	ch.durable = durable
	return amqp.Queue{Name: name}, nil
}

func (ch *fakeChannel) PublishWithContext(_ context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	ch.publishedKey = key
	ch.published = msg
	return nil
}

func TestPublishTaskCreatedDeclaresDurableQueueAndPublishesPersistentJSON(t *testing.T) {
	channel := &fakeChannel{}
	publisher := New(channel, "task_events")
	event := events.TaskEvent{
		Event:     events.TaskCreatedEvent,
		TaskID:    "t_001",
		TS:        "2026-03-26T10:20:30Z",
		RequestID: "req-001",
		Producer:  events.ProducerTasks,
		Version:   events.EventVersion,
	}

	if err := publisher.PublishTaskCreated(context.Background(), event); err != nil {
		t.Fatalf("PublishTaskCreated returned error: %v", err)
	}

	if channel.queueName != "task_events" || !channel.durable {
		t.Fatalf("expected durable task_events queue, got name=%s durable=%v", channel.queueName, channel.durable)
	}
	if channel.publishedKey != "task_events" {
		t.Fatalf("unexpected routing key: %s", channel.publishedKey)
	}
	if channel.published.DeliveryMode != amqp.Persistent {
		t.Fatalf("expected persistent delivery mode, got %d", channel.published.DeliveryMode)
	}
	if channel.published.ContentType != "application/json" {
		t.Fatalf("unexpected content type: %s", channel.published.ContentType)
	}

	var got events.TaskEvent
	if err := json.Unmarshal(channel.published.Body, &got); err != nil {
		t.Fatalf("published body is not json: %v", err)
	}
	if got != event {
		t.Fatalf("unexpected event: %+v", got)
	}
}
