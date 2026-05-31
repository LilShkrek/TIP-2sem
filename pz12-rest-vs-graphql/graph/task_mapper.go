package graph

import (
	"example.com/pz12-rest-vs-graphql/graph/model"
	"example.com/pz12-rest-vs-graphql/internal/task"
)

func toModelTask(item *task.Task) *model.Task {
	if item == nil {
		return nil
	}

	return &model.Task{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		Done:        item.Done,
	}
}
