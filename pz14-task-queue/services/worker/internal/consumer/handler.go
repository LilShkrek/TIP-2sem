package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	commonjobs "pz14-task-queue/internal/jobs"
)

type Publisher interface {
	PublishJob(ctx context.Context, queue string, job commonjobs.TaskJob) error
}

type ProcessedStore interface {
	Exists(id string) bool
	MarkDone(id string)
}

type ProcessFunc func(ctx context.Context, job commonjobs.TaskJob) error

type Message struct {
	Body []byte
	Ack  func() error
	Nack func(requeue bool) error
}

type Handler struct {
	publisher   Publisher
	processed   ProcessedStore
	taskQueue   string
	dlqQueue    string
	maxAttempts int
	delay       time.Duration
	process     ProcessFunc
	logger      *log.Logger
}

func NewHandler(publisher Publisher, processed ProcessedStore, taskQueue, dlqQueue string, maxAttempts int, delay time.Duration, logger *log.Logger) *Handler {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	if logger == nil {
		logger = log.Default()
	}

	return &Handler{
		publisher:   publisher,
		processed:   processed,
		taskQueue:   taskQueue,
		dlqQueue:    dlqQueue,
		maxAttempts: maxAttempts,
		delay:       delay,
		process:     defaultProcessTask,
		logger:      logger,
	}
}

func (h *Handler) SetProcessFunc(process ProcessFunc) {
	if process != nil {
		h.process = process
	}
}

func (h *Handler) Handle(ctx context.Context, msg Message) error {
	var job commonjobs.TaskJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		h.logger.Printf("invalid json: %v; reject without requeue", err)
		return nack(msg, false)
	}

	if err := job.ValidateMessageID(); err != nil {
		h.logger.Printf("invalid message: %v task_id=%s attempt=%d; send to DLQ", err, job.TaskID, job.Attempt)
		return h.publishDLQAndAck(ctx, msg, job)
	}

	if h.processed.Exists(job.MessageID) {
		h.logger.Printf("duplicate message_id=%s task_id=%s attempt=%d; ack without processing", job.MessageID, job.TaskID, job.Attempt)
		return ack(msg)
	}

	if err := h.wait(ctx); err != nil {
		return err
	}

	if err := h.process(ctx, job); err != nil {
		h.logger.Printf("processing failed message_id=%s task_id=%s attempt=%d error=%v", job.MessageID, job.TaskID, job.Attempt, err)
		return h.retryOrDLQ(ctx, msg, job)
	}

	h.processed.MarkDone(job.MessageID)
	h.logger.Printf("processed successfully message_id=%s task_id=%s attempt=%d", job.MessageID, job.TaskID, job.Attempt)
	return ack(msg)
}

func (h *Handler) wait(ctx context.Context) error {
	if h.delay <= 0 {
		return nil
	}

	timer := time.NewTimer(h.delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (h *Handler) retryOrDLQ(ctx context.Context, msg Message, job commonjobs.TaskJob) error {
	next := job
	next.Attempt++

	if next.Attempt <= h.maxAttempts {
		if err := h.publisher.PublishJob(ctx, h.taskQueue, next); err != nil {
			h.logger.Printf("retry publish failed message_id=%s task_id=%s next_attempt=%d error=%v", next.MessageID, next.TaskID, next.Attempt, err)
			return nack(msg, true)
		}

		h.logger.Printf("republished retry message_id=%s task_id=%s next_attempt=%d", next.MessageID, next.TaskID, next.Attempt)
		return ack(msg)
	}

	if err := h.publisher.PublishJob(ctx, h.dlqQueue, next); err != nil {
		h.logger.Printf("dlq publish failed message_id=%s task_id=%s exceeded_attempt=%d error=%v", next.MessageID, next.TaskID, next.Attempt, err)
		return nack(msg, true)
	}

	h.logger.Printf("sent to DLQ message_id=%s task_id=%s exceeded_attempt=%d", next.MessageID, next.TaskID, next.Attempt)
	return ack(msg)
}

func (h *Handler) publishDLQAndAck(ctx context.Context, msg Message, job commonjobs.TaskJob) error {
	if err := h.publisher.PublishJob(ctx, h.dlqQueue, job); err != nil {
		h.logger.Printf("dlq publish failed task_id=%s attempt=%d error=%v", job.TaskID, job.Attempt, err)
		return nack(msg, true)
	}

	h.logger.Printf("sent invalid message to DLQ task_id=%s attempt=%d", job.TaskID, job.Attempt)
	return ack(msg)
}

func defaultProcessTask(_ context.Context, job commonjobs.TaskJob) error {
	if job.TaskID == "t_fail" {
		return errors.New("simulated processing error")
	}
	return nil
}

func ack(msg Message) error {
	if msg.Ack == nil {
		return nil
	}
	if err := msg.Ack(); err != nil {
		return fmt.Errorf("ack: %w", err)
	}
	return nil
}

func nack(msg Message, requeue bool) error {
	if msg.Nack == nil {
		return nil
	}
	if err := msg.Nack(requeue); err != nil {
		return fmt.Errorf("nack: %w", err)
	}
	return nil
}
