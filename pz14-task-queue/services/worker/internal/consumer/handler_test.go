package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"testing"

	commonjobs "pz14-task-queue/internal/jobs"
	"pz14-task-queue/services/worker/internal/store"
)

type publishedJob struct {
	queue string
	job   commonjobs.TaskJob
}

type fakePublisher struct {
	published  []publishedJob
	errByQueue map[string]error
}

func (f *fakePublisher) PublishJob(_ context.Context, queue string, job commonjobs.TaskJob) error {
	if err := f.errByQueue[queue]; err != nil {
		return err
	}
	f.published = append(f.published, publishedJob{queue: queue, job: job})
	return nil
}

type fakeMessage struct {
	body         []byte
	ackCount     int
	nackCount    int
	nackRequeues []bool
}

func (m *fakeMessage) message() Message {
	return Message{
		Body: m.body,
		Ack: func() error {
			m.ackCount++
			return nil
		},
		Nack: func(requeue bool) error {
			m.nackCount++
			m.nackRequeues = append(m.nackRequeues, requeue)
			return nil
		},
	}
}

func TestHandleSuccess(t *testing.T) {
	pub := &fakePublisher{}
	processed := store.NewProcessedStore()
	handler := newTestHandler(pub, processed)

	called := false
	handler.SetProcessFunc(func(_ context.Context, job commonjobs.TaskJob) error {
		called = true
		if job.TaskID != "t_001" {
			t.Fatalf("task_id = %q, want t_001", job.TaskID)
		}
		return nil
	})

	msg := newJobMessage(t, commonjobs.TaskJob{Job: commonjobs.ProcessTaskJob, TaskID: "t_001", Attempt: 1, MessageID: "message-1"})
	if err := handler.Handle(context.Background(), msg.message()); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if !called {
		t.Fatal("process function was not called")
	}
	if msg.ackCount != 1 || msg.nackCount != 0 {
		t.Fatalf("ack=%d nack=%d, want ack=1 nack=0", msg.ackCount, msg.nackCount)
	}
	if !processed.Exists("message-1") {
		t.Fatal("message_id was not marked as processed")
	}
	if len(pub.published) != 0 {
		t.Fatalf("published %d messages, want 0", len(pub.published))
	}
}

func TestHandleRetry(t *testing.T) {
	pub := &fakePublisher{}
	handler := newTestHandler(pub, store.NewProcessedStore())
	handler.SetProcessFunc(func(context.Context, commonjobs.TaskJob) error {
		return errors.New("fail")
	})

	msg := newJobMessage(t, commonjobs.TaskJob{Job: commonjobs.ProcessTaskJob, TaskID: "t_fail", Attempt: 1, MessageID: "message-1"})
	if err := handler.Handle(context.Background(), msg.message()); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if msg.ackCount != 1 || msg.nackCount != 0 {
		t.Fatalf("ack=%d nack=%d, want ack=1 nack=0", msg.ackCount, msg.nackCount)
	}
	if len(pub.published) != 1 {
		t.Fatalf("published %d messages, want 1", len(pub.published))
	}
	if pub.published[0].queue != "task_jobs" {
		t.Fatalf("queue = %q, want task_jobs", pub.published[0].queue)
	}
	if pub.published[0].job.Attempt != 2 {
		t.Fatalf("retry attempt = %d, want 2", pub.published[0].job.Attempt)
	}
}

func TestHandleDLQAfterMaxAttempts(t *testing.T) {
	pub := &fakePublisher{}
	handler := newTestHandler(pub, store.NewProcessedStore())
	handler.SetProcessFunc(func(context.Context, commonjobs.TaskJob) error {
		return errors.New("fail")
	})

	msg := newJobMessage(t, commonjobs.TaskJob{Job: commonjobs.ProcessTaskJob, TaskID: "t_fail", Attempt: 3, MessageID: "message-1"})
	if err := handler.Handle(context.Background(), msg.message()); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if msg.ackCount != 1 || msg.nackCount != 0 {
		t.Fatalf("ack=%d nack=%d, want ack=1 nack=0", msg.ackCount, msg.nackCount)
	}
	if len(pub.published) != 1 {
		t.Fatalf("published %d messages, want 1", len(pub.published))
	}
	if pub.published[0].queue != "task_jobs_dlq" {
		t.Fatalf("queue = %q, want task_jobs_dlq", pub.published[0].queue)
	}
	if pub.published[0].job.Attempt != 4 {
		t.Fatalf("dlq attempt = %d, want 4", pub.published[0].job.Attempt)
	}
}

func TestHandleDuplicateSkipsProcessing(t *testing.T) {
	pub := &fakePublisher{}
	processed := store.NewProcessedStore()
	processed.MarkDone("message-1")
	handler := newTestHandler(pub, processed)

	handler.SetProcessFunc(func(context.Context, commonjobs.TaskJob) error {
		t.Fatal("process function must not be called for duplicate")
		return nil
	})

	msg := newJobMessage(t, commonjobs.TaskJob{Job: commonjobs.ProcessTaskJob, TaskID: "t_001", Attempt: 1, MessageID: "message-1"})
	if err := handler.Handle(context.Background(), msg.message()); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if msg.ackCount != 1 || msg.nackCount != 0 {
		t.Fatalf("ack=%d nack=%d, want ack=1 nack=0", msg.ackCount, msg.nackCount)
	}
	if len(pub.published) != 0 {
		t.Fatalf("published %d messages, want 0", len(pub.published))
	}
}

func TestHandleRetryPublishErrorNacksWithRequeue(t *testing.T) {
	pub := &fakePublisher{errByQueue: map[string]error{"task_jobs": errors.New("rabbit down")}}
	handler := newTestHandler(pub, store.NewProcessedStore())
	handler.SetProcessFunc(func(context.Context, commonjobs.TaskJob) error {
		return errors.New("fail")
	})

	msg := newJobMessage(t, commonjobs.TaskJob{Job: commonjobs.ProcessTaskJob, TaskID: "t_fail", Attempt: 1, MessageID: "message-1"})
	if err := handler.Handle(context.Background(), msg.message()); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if msg.ackCount != 0 || msg.nackCount != 1 {
		t.Fatalf("ack=%d nack=%d, want ack=0 nack=1", msg.ackCount, msg.nackCount)
	}
	if len(msg.nackRequeues) != 1 || !msg.nackRequeues[0] {
		t.Fatalf("nack requeues = %+v, want [true]", msg.nackRequeues)
	}
}

func TestHandleDLQPublishErrorNacksWithRequeue(t *testing.T) {
	pub := &fakePublisher{errByQueue: map[string]error{"task_jobs_dlq": errors.New("rabbit down")}}
	handler := newTestHandler(pub, store.NewProcessedStore())
	handler.SetProcessFunc(func(context.Context, commonjobs.TaskJob) error {
		return errors.New("fail")
	})

	msg := newJobMessage(t, commonjobs.TaskJob{Job: commonjobs.ProcessTaskJob, TaskID: "t_fail", Attempt: 3, MessageID: "message-1"})
	if err := handler.Handle(context.Background(), msg.message()); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if msg.ackCount != 0 || msg.nackCount != 1 {
		t.Fatalf("ack=%d nack=%d, want ack=0 nack=1", msg.ackCount, msg.nackCount)
	}
	if len(msg.nackRequeues) != 1 || !msg.nackRequeues[0] {
		t.Fatalf("nack requeues = %+v, want [true]", msg.nackRequeues)
	}
}

func TestHandleInvalidJSONRejectsWithoutRequeue(t *testing.T) {
	handler := newTestHandler(&fakePublisher{}, store.NewProcessedStore())
	msg := &fakeMessage{body: []byte(`{bad json}`)}

	if err := handler.Handle(context.Background(), msg.message()); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if msg.ackCount != 0 || msg.nackCount != 1 {
		t.Fatalf("ack=%d nack=%d, want ack=0 nack=1", msg.ackCount, msg.nackCount)
	}
	if len(msg.nackRequeues) != 1 || msg.nackRequeues[0] {
		t.Fatalf("nack requeues = %+v, want [false]", msg.nackRequeues)
	}
}

func newTestHandler(pub *fakePublisher, processed *store.ProcessedStore) *Handler {
	if pub.errByQueue == nil {
		pub.errByQueue = make(map[string]error)
	}
	return NewHandler(pub, processed, "task_jobs", "task_jobs_dlq", 3, 0, log.New(io.Discard, "", 0))
}

func newJobMessage(t *testing.T, job commonjobs.TaskJob) *fakeMessage {
	t.Helper()

	body, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("marshal job: %v", err)
	}

	return &fakeMessage{body: body}
}
