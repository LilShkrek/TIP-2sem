package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"example.com/pz13-rabbitmq/internal/events"
)

var ErrInvalidTask = errors.New("title and description are required")

type Publisher interface {
	PublishTaskCreated(ctx context.Context, event events.TaskEvent) error
}

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateTaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Service struct {
	mu        sync.Mutex
	nextID    int
	tasks     []Task
	publisher Publisher
	logger    *log.Logger
	now       func() time.Time
}

func New(publisher Publisher, logger *log.Logger) *Service {
	if logger == nil {
		logger = log.Default()
	}

	return &Service{
		publisher: publisher,
		logger:    logger,
		now:       time.Now,
	}
}

func (s *Service) CreateTask(ctx context.Context, input CreateTaskInput, requestID string) (Task, error) {
	title := strings.TrimSpace(input.Title)
	description := strings.TrimSpace(input.Description)
	if title == "" || description == "" {
		return Task{}, ErrInvalidTask
	}

	task := s.create(title, description)
	s.publishCreated(ctx, task, requestID)

	return task, nil
}

func (s *Service) create(title, description string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := s.now().UTC()
	task := Task{
		ID:          fmt.Sprintf("t_%03d", s.nextID),
		Title:       title,
		Description: description,
		CreatedAt:   now,
	}

	s.tasks = append(s.tasks, task)

	return task
}

func (s *Service) publishCreated(ctx context.Context, task Task, requestID string) {
	if s.publisher == nil {
		return
	}

	event := events.TaskEvent{
		Event:     events.TaskCreatedEvent,
		TaskID:    task.ID,
		TS:        task.CreatedAt.Format(time.RFC3339),
		RequestID: strings.TrimSpace(requestID),
		Producer:  events.ProducerTasks,
		Version:   events.EventVersion,
	}

	if err := s.publisher.PublishTaskCreated(ctx, event); err != nil {
		s.logger.Printf("task event publish error: %v", err)
	}
}
