package task

import (
	"context"
	"errors"
	"strings"
)

var ErrEmptyTitle = errors.New("task title must not be empty")

type CreateInput struct {
	Title       string
	Description string
}

type UpdateInput struct {
	Title       *string
	Description *string
	Done        *bool
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (*Task, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrEmptyTitle
	}

	return s.repo.Create(ctx, title, input.Description)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (*Task, error) {
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, ErrEmptyTitle
		}
		input.Title = &title
	}

	return s.repo.Update(ctx, id, UpdateTask{
		Title:       input.Title,
		Description: input.Description,
		Done:        input.Done,
	})
}
