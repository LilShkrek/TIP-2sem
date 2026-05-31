package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"example.com/pz9-redis-cache/internal/service"
	"example.com/pz9-redis-cache/internal/task"
)

type Handler struct {
	service *service.TaskService
}

func NewHandler(service *service.TaskService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) TaskByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseTaskID(r.URL.Path)
	if !ok {
		http.Error(w, "invalid task path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTaskByID(w, r, id)
	case http.MethodPatch:
		h.patchTask(w, r, id)
	case http.MethodDelete:
		h.deleteTask(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) getTaskByID(w http.ResponseWriter, r *http.Request, id int64) {
	t, err := h.service.GetTaskByID(r.Context(), id)
	if err != nil {
		writeTaskError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) patchTask(w http.ResponseWriter, r *http.Request, id int64) {
	defer r.Body.Close()

	var patch task.Patch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	t, err := h.service.PatchTask(r.Context(), id, patch)
	if err != nil {
		writeTaskError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.service.DeleteTask(r.Context(), id); err != nil {
		writeTaskError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseTaskID(path string) (int64, bool) {
	rawID, ok := strings.CutPrefix(path, "/v1/tasks/")
	if !ok || rawID == "" || strings.Contains(rawID, "/") {
		return 0, false
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}

	return id, true
}

func writeTaskError(w http.ResponseWriter, err error) {
	if errors.Is(err, task.ErrTaskNotFound) {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
