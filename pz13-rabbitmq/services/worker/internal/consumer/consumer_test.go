package consumer

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type fakeAcknowledger struct {
	acked    bool
	nacked   bool
	multiple bool
	requeue  bool
}

func (a *fakeAcknowledger) Ack(tag uint64, multiple bool) error {
	a.acked = true
	a.multiple = multiple
	return nil
}

func (a *fakeAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	a.nacked = true
	a.multiple = multiple
	a.requeue = requeue
	return nil
}

func (a *fakeAcknowledger) Reject(tag uint64, requeue bool) error {
	return nil
}

func TestHandleDeliveryAcksValidMessage(t *testing.T) {
	ack := &fakeAcknowledger{}
	var logs bytes.Buffer
	delivery := amqp.Delivery{
		Acknowledger: ack,
		DeliveryTag:  1,
		Body:         []byte(`{"event":"task.created","task_id":"t_001","ts":"2026-03-26T10:20:30Z","request_id":"req-001"}`),
	}

	HandleDelivery(delivery, log.New(&logs, "", 0))

	if !ack.acked {
		t.Fatal("expected Ack(false)")
	}
	if ack.multiple {
		t.Fatal("expected multiple=false")
	}
	if ack.nacked {
		t.Fatal("did not expect Nack")
	}
	if !strings.Contains(logs.String(), "event=task.created") || !strings.Contains(logs.String(), "request_id=req-001") {
		t.Fatalf("unexpected log: %s", logs.String())
	}
}

func TestHandleDeliveryNacksInvalidJSONWithoutRequeue(t *testing.T) {
	ack := &fakeAcknowledger{}
	delivery := amqp.Delivery{
		Acknowledger: ack,
		DeliveryTag:  1,
		Body:         []byte(`{bad json`),
	}

	HandleDelivery(delivery, log.New(&bytes.Buffer{}, "", 0))

	if !ack.nacked {
		t.Fatal("expected Nack(false, false)")
	}
	if ack.multiple {
		t.Fatal("expected multiple=false")
	}
	if ack.requeue {
		t.Fatal("expected requeue=false")
	}
	if ack.acked {
		t.Fatal("did not expect Ack")
	}
}

type fakeConsumerChannel struct {
	queueName string
	durable   bool
	prefetch  int
	autoAck   bool
	msgs      chan amqp.Delivery
}

func (ch *fakeConsumerChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	ch.queueName = name
	ch.durable = durable
	return amqp.Queue{Name: name}, nil
}

func (ch *fakeConsumerChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	ch.prefetch = prefetchCount
	return nil
}

func (ch *fakeConsumerChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	ch.autoAck = autoAck
	close(ch.msgs)
	return ch.msgs, nil
}

func TestRunDeclaresQueueWithPrefetchAndManualAck(t *testing.T) {
	ch := &fakeConsumerChannel{msgs: make(chan amqp.Delivery)}

	if err := Run(context.Background(), ch, "task_events", log.New(&bytes.Buffer{}, "", 0)); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if ch.queueName != "task_events" || !ch.durable {
		t.Fatalf("expected durable task_events queue, got name=%s durable=%v", ch.queueName, ch.durable)
	}
	if ch.prefetch != 1 {
		t.Fatalf("expected prefetch=1, got %d", ch.prefetch)
	}
	if ch.autoAck {
		t.Fatal("expected autoAck=false")
	}
}
