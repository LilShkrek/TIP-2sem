package task

import (
	"context"
	"errors"
	"testing"
)

func TestServiceCreateAndUpdateTask(t *testing.T) {
	service := NewService(NewMemoryRepository(DefaultSeed()))

	created, err := service.Create(context.Background(), CreateInput{
		Title:       "  Сравнить REST и GraphQL  ",
		Description: "Практическая работа N12",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if created.ID != "t_003" {
		t.Fatalf("created ID = %q, want %q", created.ID, "t_003")
	}
	if created.Title != "Сравнить REST и GraphQL" {
		t.Fatalf("created title = %q", created.Title)
	}
	if created.Done {
		t.Fatal("new task must not be done")
	}

	done := true
	updated, err := service.Update(context.Background(), created.ID, UpdateInput{Done: &done})
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if !updated.Done {
		t.Fatal("updated task must be done")
	}
}

func TestServiceRejectsEmptyTitle(t *testing.T) {
	service := NewService(NewMemoryRepository(nil))

	_, err := service.Create(context.Background(), CreateInput{Title: "   "})
	if !errors.Is(err, ErrEmptyTitle) {
		t.Fatalf("error = %v, want %v", err, ErrEmptyTitle)
	}
}

func TestServiceReturnsNotFound(t *testing.T) {
	service := NewService(NewMemoryRepository(DefaultSeed()))

	_, err := service.Get(context.Background(), "unknown")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrNotFound)
	}
}
