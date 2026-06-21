package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	commonjobs "pz14-task-queue/internal/jobs"
	taskjobs "pz14-task-queue/services/tasks/internal/jobs"
)

type fakeEnqueuer struct {
	job commonjobs.TaskJob
	err error
}

func (f fakeEnqueuer) EnqueueProcessTask(_ context.Context, taskID string) (commonjobs.TaskJob, error) {
	if strings.TrimSpace(taskID) == "" {
		return commonjobs.TaskJob{}, taskjobs.ErrEmptyTaskID
	}
	if f.err != nil {
		return commonjobs.TaskJob{}, f.err
	}
	return f.job, nil
}

func TestProcessTaskAccepted(t *testing.T) {
	handler := NewHandler(fakeEnqueuer{
		job: commonjobs.TaskJob{
			TaskID:    "t_001",
			MessageID: "message-1",
		},
	})

	req := httptest.NewRequest(http.MethodPost, processTaskPath, strings.NewReader(`{"task_id":"t_001"}`))
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}

	var resp processTaskResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != "accepted" || resp.TaskID != "t_001" || resp.MessageID != "message-1" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestProcessTaskValidatesJSON(t *testing.T) {
	handler := NewHandler(fakeEnqueuer{})
	req := httptest.NewRequest(http.MethodPost, processTaskPath, strings.NewReader(`{bad json}`))
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestProcessTaskValidatesTaskID(t *testing.T) {
	handler := NewHandler(fakeEnqueuer{})
	req := httptest.NewRequest(http.MethodPost, processTaskPath, strings.NewReader(`{"task_id":" "}`))
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestProcessTaskPublishError(t *testing.T) {
	handler := NewHandler(fakeEnqueuer{err: errors.New("publish failed")})
	req := httptest.NewRequest(http.MethodPost, processTaskPath, strings.NewReader(`{"task_id":"t_001"}`))
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
}
