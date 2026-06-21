package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"example.com/pz13-rabbitmq/services/tasks/internal/service"
)

type TaskService interface {
	CreateTask(ctx context.Context, input service.CreateTaskInput, requestID string) (service.Task, error)
}

type Handler struct {
	service TaskService
}

func New(service TaskService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/tasks", h.createTask)

	return mux
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var input service.CreateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	task, err := h.service.CreateTask(r.Context(), input, r.Header.Get("X-Request-ID"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidTask) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
