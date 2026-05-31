package service

import (
	"context"
	"errors"
	"testing"

	"example.com/pz11-graphql/services/graphql/internal/repository"
)

func TestTaskServiceCreateUpdateDelete(t *testing.T) {
	ctx := context.Background()
	svc := NewTaskService(repository.NewMemoryTaskRepository())

	created, err := svc.CreateTask(ctx, CreateTaskInput{
		Title:       "Learn GraphQL",
		Description: "Build gqlgen API",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if created.ID != "t_001" || created.Done {
		t.Fatalf("unexpected created task: %+v", created)
	}

	done := true
	updated, err := svc.UpdateTask(ctx, created.ID, UpdateTaskInput{Done: &done})
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if !updated.Done {
		t.Fatalf("expected task to be done: %+v", updated)
	}

	if err := svc.DeleteTask(ctx, created.ID); err != nil {
		t.Fatalf("delete task: %v", err)
	}

	_, err = svc.GetTask(ctx, created.ID)
	if !errors.Is(err, repository.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskServiceRejectsEmptyTitle(t *testing.T) {
	svc := NewTaskService(repository.NewMemoryTaskRepository())

	_, err := svc.CreateTask(context.Background(), CreateTaskInput{Title: "   "})
	if !errors.Is(err, ErrEmptyTaskTitle) {
		t.Fatalf("expected ErrEmptyTaskTitle, got %v", err)
	}
}
