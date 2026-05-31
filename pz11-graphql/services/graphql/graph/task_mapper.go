package graph

import (
	"example.com/pz11-graphql/services/graphql/graph/model"
	"example.com/pz11-graphql/services/graphql/internal/service"
)

func toModelTask(task *service.Task) *model.Task {
	var description *string
	if task.Description != "" {
		description = &task.Description
	}

	return &model.Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: description,
		Done:        task.Done,
	}
}
