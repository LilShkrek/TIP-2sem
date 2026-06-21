package jobs

import (
	"context"
	"testing"

	commonjobs "pz14-task-queue/internal/jobs"
)

type mockPublisher struct {
	queue string
	job   commonjobs.TaskJob
	err   error
}

func (m *mockPublisher) PublishJob(_ context.Context, queue string, job commonjobs.TaskJob) error {
	m.queue = queue
	m.job = job
	return m.err
}

func TestEnqueueProcessTaskBuildsInitialJob(t *testing.T) {
	publisher := &mockPublisher{}
	service := NewService(publisher, "task_jobs", func() (string, error) {
		return "message-1", nil
	})

	job, err := service.EnqueueProcessTask(context.Background(), " t_001 ")
	if err != nil {
		t.Fatalf("EnqueueProcessTask returned error: %v", err)
	}

	if publisher.queue != "task_jobs" {
		t.Fatalf("queue = %q, want task_jobs", publisher.queue)
	}
	if job.Job != commonjobs.ProcessTaskJob {
		t.Fatalf("job name = %q, want %q", job.Job, commonjobs.ProcessTaskJob)
	}
	if job.TaskID != "t_001" {
		t.Fatalf("task_id = %q, want t_001", job.TaskID)
	}
	if job.Attempt != 1 {
		t.Fatalf("attempt = %d, want 1", job.Attempt)
	}
	if job.MessageID == "" {
		t.Fatal("message_id is empty")
	}
	if publisher.job != job {
		t.Fatalf("published job = %+v, want %+v", publisher.job, job)
	}
}

func TestEnqueueProcessTaskValidatesTaskID(t *testing.T) {
	service := NewService(&mockPublisher{}, "task_jobs", func() (string, error) {
		return "message-1", nil
	})

	if _, err := service.EnqueueProcessTask(context.Background(), " "); err != ErrEmptyTaskID {
		t.Fatalf("error = %v, want ErrEmptyTaskID", err)
	}
}

func TestNewUUIDReturnsUniqueIDs(t *testing.T) {
	first, err := NewUUID()
	if err != nil {
		t.Fatalf("NewUUID first: %v", err)
	}
	second, err := NewUUID()
	if err != nil {
		t.Fatalf("NewUUID second: %v", err)
	}

	if first == "" || second == "" {
		t.Fatal("uuid must not be empty")
	}
	if first == second {
		t.Fatalf("uuid values must be unique, got %q", first)
	}
	if len(first) != 36 || len(second) != 36 {
		t.Fatalf("uuid length = %d and %d, want 36", len(first), len(second))
	}
}
