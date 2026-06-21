package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	commonjobs "pz14-task-queue/internal/jobs"
	taskjobs "pz14-task-queue/services/tasks/internal/jobs"
)

const processTaskPath = "/v1/jobs/process-task"

type JobEnqueuer interface {
	EnqueueProcessTask(ctx context.Context, taskID string) (commonjobs.TaskJob, error)
}

type Handler struct {
	enqueuer JobEnqueuer
}

func NewHandler(enqueuer JobEnqueuer) *Handler {
	return &Handler{enqueuer: enqueuer}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(processTaskPath, h.processTask)
	return mux
}

type processTaskRequest struct {
	TaskID string `json:"task_id"`
}

type processTaskResponse struct {
	Status    string `json:"status"`
	TaskID    string `json:"task_id"`
	MessageID string `json:"message_id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) processTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	defer r.Body.Close()

	var req processTaskRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	job, err := h.enqueuer.EnqueueProcessTask(r.Context(), req.TaskID)
	if err != nil {
		if errors.Is(err, taskjobs.ErrEmptyTaskID) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "task_id is required"})
			return
		}

		writeJSON(w, http.StatusBadGateway, errorResponse{Error: "publish failed"})
		return
	}

	writeJSON(w, http.StatusAccepted, processTaskResponse{
		Status:    "accepted",
		TaskID:    job.TaskID,
		MessageID: job.MessageID,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
