package graph

import "example.com/pz12-rest-vs-graphql/internal/task"

type Resolver struct {
	TaskService *task.Service
}

func NewResolver(taskService *task.Service) *Resolver {
	return &Resolver{TaskService: taskService}
}
