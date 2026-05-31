package task

import (
	"errors"
	"sync"
	"time"
)

var ErrTaskNotFound = errors.New("task not found")

type Repo struct {
	mu   sync.RWMutex
	data map[int64]Task
}

func NewRepo() *Repo {
	return &Repo{
		data: map[int64]Task{
			1: {
				ID:          1,
				Title:       "Изучить Redis",
				Description: "Разобрать cache-aside",
				DueDate:     time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
			},
			2: {
				ID:          2,
				Title:       "Сделать ПЗ",
				Description: "Реализовать кэширование по id",
				DueDate:     time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC),
			},
		},
	}
}

func (r *Repo) GetByID(id int64) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.data[id]
	if !ok {
		return Task{}, ErrTaskNotFound
	}

	return t, nil
}

func (r *Repo) Patch(id int64, patch Patch) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.data[id]
	if !ok {
		return Task{}, ErrTaskNotFound
	}

	if patch.Title != nil {
		t.Title = *patch.Title
	}
	if patch.Description != nil {
		t.Description = *patch.Description
	}
	if patch.DueDate != nil {
		t.DueDate = *patch.DueDate
	}

	r.data[id] = t
	return t, nil
}

func (r *Repo) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[id]; !ok {
		return ErrTaskNotFound
	}

	delete(r.data, id)
	return nil
}
