package service

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"example.com/pz13-rabbitmq/internal/events"
)

type fakePublisher struct {
	event events.TaskEvent
	err   error
	calls int
}

func (p *fakePublisher) PublishTaskCreated(_ context.Context, event events.TaskEvent) error {
	p.calls++
	p.event = event
	return p.err
}

func TestCreateTaskPublishesTaskCreatedEvent(t *testing.T) {
	publisher := &fakePublisher{}
	svc := New(publisher, log.New(&bytes.Buffer{}, "", 0))
	svc.now = func() time.Time {
		return time.Date(2026, 3, 26, 10, 20, 30, 0, time.UTC)
	}

	task, err := svc.CreateTask(context.Background(), CreateTaskInput{
		Title:       " Rabbit ",
		Description: " publish event ",
	}, "req-001")
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}

	if task.ID != "t_001" || task.Title != "Rabbit" || task.Description != "publish event" {
		t.Fatalf("unexpected task: %+v", task)
	}

	if publisher.calls != 1 {
		t.Fatalf("expected one publish call, got %d", publisher.calls)
	}

	event := publisher.event
	if event.Event != events.TaskCreatedEvent {
		t.Fatalf("unexpected event name: %s", event.Event)
	}
	if event.TaskID != task.ID {
		t.Fatalf("unexpected task id: %s", event.TaskID)
	}
	if event.TS != "2026-03-26T10:20:30Z" {
		t.Fatalf("unexpected timestamp: %s", event.TS)
	}
	if event.RequestID != "req-001" || event.Producer != events.ProducerTasks || event.Version != events.EventVersion {
		t.Fatalf("unexpected event metadata: %+v", event)
	}
}

func TestCreateTaskUsesBestEffortPublishing(t *testing.T) {
	publisher := &fakePublisher{err: errors.New("rabbit unavailable")}
	var logs bytes.Buffer
	svc := New(publisher, log.New(&logs, "", 0))

	task, err := svc.CreateTask(context.Background(), CreateTaskInput{
		Title:       "Rabbit",
		Description: "publish event",
	}, "")
	if err != nil {
		t.Fatalf("CreateTask returned error: %v", err)
	}
	if task.ID == "" {
		t.Fatal("expected created task")
	}
	if !strings.Contains(logs.String(), "task event publish error") {
		t.Fatalf("expected publish error in log, got %q", logs.String())
	}
}

func TestCreateTaskValidatesInput(t *testing.T) {
	svc := New(&fakePublisher{}, log.New(&bytes.Buffer{}, "", 0))

	_, err := svc.CreateTask(context.Background(), CreateTaskInput{
		Title:       "",
		Description: "publish event",
	}, "")
	if !errors.Is(err, ErrInvalidTask) {
		t.Fatalf("expected ErrInvalidTask, got %v", err)
	}
}
