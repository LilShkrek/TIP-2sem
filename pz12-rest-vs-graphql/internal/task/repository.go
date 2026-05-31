package task

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

type UpdateTask struct {
	Title       *string
	Description *string
	Done        *bool
}

type Repository interface {
	List(ctx context.Context) ([]Task, error)
	Get(ctx context.Context, id string) (*Task, error)
	Create(ctx context.Context, title string, description string) (*Task, error)
	Update(ctx context.Context, id string, input UpdateTask) (*Task, error)
}

type MemoryRepository struct {
	mu     sync.RWMutex
	tasks  map[string]*Task
	nextID int
}

func NewMemoryRepository(seed []Task) *MemoryRepository {
	repo := &MemoryRepository{
		tasks:  make(map[string]*Task, len(seed)),
		nextID: 1,
	}

	for _, item := range seed {
		copyItem := item
		repo.tasks[item.ID] = &copyItem
	}
	repo.nextID = len(seed) + 1

	return repo
}

func DefaultSeed() []Task {
	return []Task{
		{
			ID:          "t_001",
			Title:       "Первая задача",
			Description: "Учебный пример для сравнения REST и GraphQL",
			Done:        false,
		},
		{
			ID:          "t_002",
			Title:       "Вторая задача",
			Description: "Проверка получения списка и деталей",
			Done:        true,
		},
	}
}

func (r *MemoryRepository) List(ctx context.Context) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]Task, 0, len(r.tasks))
	for _, item := range r.tasks {
		tasks = append(tasks, *item)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})

	return tasks, nil
}

func (r *MemoryRepository) Get(ctx context.Context, id string) (*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}

	copyItem := *item
	return &copyItem, nil
}

func (r *MemoryRepository) Create(ctx context.Context, title string, description string) (*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	item := &Task{
		ID:          fmt.Sprintf("t_%03d", r.nextID),
		Title:       title,
		Description: description,
		Done:        false,
	}
	r.tasks[item.ID] = item
	r.nextID++

	copyItem := *item
	return &copyItem, nil
}

func (r *MemoryRepository) Update(ctx context.Context, id string, input UpdateTask) (*Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}

	if input.Title != nil {
		item.Title = *input.Title
	}
	if input.Description != nil {
		item.Description = *input.Description
	}
	if input.Done != nil {
		item.Done = *input.Done
	}

	copyItem := *item
	return &copyItem, nil
}
