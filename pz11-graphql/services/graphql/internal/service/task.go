package service

import (
	"context"
	"errors"
	"strings"

	"example.com/pz11-graphql/services/graphql/internal/repository"
)

var ErrEmptyTaskTitle = errors.New("task title must not be empty")

type Task struct {
	ID          string
	Title       string
	Description string
	Done        bool
}

type CreateTaskInput struct {
	Title       string
	Description string
}

type UpdateTaskInput struct {
	Title       *string
	Description *string
	Done        *bool
}

type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) ListTasks(ctx context.Context) ([]Task, error) {
	tasks, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, toServiceTask(task))
	}

	return result, nil
}

func (s *TaskService) GetTask(ctx context.Context, id string) (*Task, error) {
	task, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	result := toServiceTask(*task)
	return &result, nil
}

func (s *TaskService) CreateTask(ctx context.Context, input CreateTaskInput) (*Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrEmptyTaskTitle
	}

	task, err := s.repo.Create(ctx, title, input.Description)
	if err != nil {
		return nil, err
	}

	result := toServiceTask(*task)
	return &result, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, id string, input UpdateTaskInput) (*Task, error) {
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, ErrEmptyTaskTitle
		}
		input.Title = &title
	}

	task, err := s.repo.Update(ctx, id, repository.UpdateTask{
		Title:       input.Title,
		Description: input.Description,
		Done:        input.Done,
	})
	if err != nil {
		return nil, err
	}

	result := toServiceTask(*task)
	return &result, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func toServiceTask(task repository.Task) Task {
	return Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Done:        task.Done,
	}
}
