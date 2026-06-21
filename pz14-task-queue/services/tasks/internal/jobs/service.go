package jobs

import (
	"context"
	"errors"
	"strings"

	commonjobs "pz14-task-queue/internal/jobs"
)

var ErrEmptyTaskID = errors.New("task_id is required")

type Publisher interface {
	PublishJob(ctx context.Context, queue string, job commonjobs.TaskJob) error
}

type MessageIDGenerator func() (string, error)

type Service struct {
	publisher    Publisher
	queue        string
	newMessageID MessageIDGenerator
}

func NewService(publisher Publisher, queue string, newMessageID MessageIDGenerator) *Service {
	if newMessageID == nil {
		newMessageID = NewUUID
	}

	return &Service{
		publisher:    publisher,
		queue:        queue,
		newMessageID: newMessageID,
	}
}

func (s *Service) EnqueueProcessTask(ctx context.Context, taskID string) (commonjobs.TaskJob, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return commonjobs.TaskJob{}, ErrEmptyTaskID
	}

	messageID, err := s.newMessageID()
	if err != nil {
		return commonjobs.TaskJob{}, err
	}

	job := commonjobs.TaskJob{
		Job:       commonjobs.ProcessTaskJob,
		TaskID:    taskID,
		Attempt:   1,
		MessageID: messageID,
	}

	if err := s.publisher.PublishJob(ctx, s.queue, job); err != nil {
		return commonjobs.TaskJob{}, err
	}

	return job, nil
}
