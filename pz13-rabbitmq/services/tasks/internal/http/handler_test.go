package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/pz13-rabbitmq/services/tasks/internal/service"
)

type fakeTaskService struct {
	input     service.CreateTaskInput
	requestID string
	task      service.Task
	err       error
}

func (s *fakeTaskService) CreateTask(_ context.Context, input service.CreateTaskInput, requestID string) (service.Task, error) {
	s.input = input
	s.requestID = requestID
	return s.task, s.err
}

func TestCreateTaskHandlerReturnsCreatedTask(t *testing.T) {
	task := service.Task{
		ID:          "t_001",
		Title:       "Rabbit",
		Description: "publish event",
		CreatedAt:   time.Date(2026, 3, 26, 10, 20, 30, 0, time.UTC),
	}
	fake := &fakeTaskService{task: task}
	handler := New(fake).Routes()

	req := httptest.NewRequest(http.MethodPost, "/v1/tasks", strings.NewReader(`{"title":"Rabbit","description":"publish event"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "req-001")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if fake.requestID != "req-001" {
		t.Fatalf("unexpected request id: %s", fake.requestID)
	}
	if fake.input.Title != "Rabbit" || fake.input.Description != "publish event" {
		t.Fatalf("unexpected input: %+v", fake.input)
	}

	var got service.Task
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != task.ID {
		t.Fatalf("unexpected task response: %+v", got)
	}
}

func TestCreateTaskHandlerReturnsBadRequestForInvalidJSON(t *testing.T) {
	handler := New(&fakeTaskService{}).Routes()
	req := httptest.NewRequest(http.MethodPost, "/v1/tasks", strings.NewReader(`{"title":`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestCreateTaskHandlerReturnsBadRequestForInvalidTask(t *testing.T) {
	handler := New(&fakeTaskService{err: service.ErrInvalidTask}).Routes()
	req := httptest.NewRequest(http.MethodPost, "/v1/tasks", strings.NewReader(`{"title":"","description":"x"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
