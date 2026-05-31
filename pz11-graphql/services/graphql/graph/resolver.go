package graph

import "example.com/pz11-graphql/services/graphql/internal/service"

type Resolver struct {
	TaskService *service.TaskService
}

func NewResolver(taskService *service.TaskService) *Resolver {
	return &Resolver{TaskService: taskService}
}
