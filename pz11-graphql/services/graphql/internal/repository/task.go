package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

var ErrTaskNotFound = errors.New("task not found")

type Task struct {
	ID          string
	Title       string
	Description string
	Done        bool
}

type UpdateTask struct {
	Title       *string
	Description *string
	Done        *bool
}

type TaskRepository interface {
	List(ctx context.Context) ([]Task, error)
	Get(ctx context.Context, id string) (*Task, error)
	Create(ctx context.Context, title string, description string) (*Task, error)
	Update(ctx context.Context, id string, input UpdateTask) (*Task, error)
	Delete(ctx context.Context, id string) error
}

type MemoryTaskRepository struct {
	mu     sync.RWMutex
	tasks  map[string]*Task
	nextID int
}

func NewMemoryTaskRepository() *MemoryTaskRepository {
	return &MemoryTaskRepository{
		tasks:  make(map[string]*Task),
		nextID: 1,
	}
}

func (r *MemoryTaskRepository) List(ctx context.Context) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, *task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})

	return tasks, nil
}

func (r *MemoryTaskRepository) Get(ctx context.Context, id string) (*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return nil, ErrTaskNotFound
	}

	copyTask := *task
	return &copyTask, nil
}

func (r *MemoryTaskRepository) Create(ctx context.Context, title string, description string) (*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	task := &Task{
		ID:          fmt.Sprintf("t_%03d", r.nextID),
		Title:       title,
		Description: description,
		Done:        false,
	}
	r.tasks[task.ID] = task
	r.nextID++

	copyTask := *task
	return &copyTask, nil
}

func (r *MemoryTaskRepository) Update(ctx context.Context, id string, input UpdateTask) (*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return nil, ErrTaskNotFound
	}

	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Done != nil {
		task.Done = *input.Done
	}

	copyTask := *task
	return &copyTask, nil
}

func (r *MemoryTaskRepository) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return ErrTaskNotFound
	}

	delete(r.tasks, id)
	return nil
}
