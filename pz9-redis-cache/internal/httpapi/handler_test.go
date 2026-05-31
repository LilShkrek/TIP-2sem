package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/pz9-redis-cache/internal/config"
	"example.com/pz9-redis-cache/internal/service"
	"example.com/pz9-redis-cache/internal/task"
)

func TestTaskByIDGetPatchDelete(t *testing.T) {
	handler := newTestHandler()

	getReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/1", nil)
	getResp := httptest.NewRecorder()
	handler.TaskByID(getResp, getReq)

	if getResp.Code != http.StatusOK {
		t.Fatalf("GET status: got %d, want %d", getResp.Code, http.StatusOK)
	}

	var got task.Task
	if err := json.NewDecoder(getResp.Body).Decode(&got); err != nil {
		t.Fatalf("decode GET response: %v", err)
	}
	if got.ID != 1 || got.Title == "" {
		t.Fatalf("unexpected task after GET: %+v", got)
	}

	patchBody := `{"title":"Обновлённая задача","due_date":"2026-01-22T00:00:00Z"}`
	patchReq := httptest.NewRequest(http.MethodPatch, "/v1/tasks/1", strings.NewReader(patchBody))
	patchResp := httptest.NewRecorder()
	handler.TaskByID(patchResp, patchReq)

	if patchResp.Code != http.StatusOK {
		t.Fatalf("PATCH status: got %d, want %d", patchResp.Code, http.StatusOK)
	}

	var patched task.Task
	if err := json.NewDecoder(patchResp.Body).Decode(&patched); err != nil {
		t.Fatalf("decode PATCH response: %v", err)
	}
	if patched.Title != "Обновлённая задача" {
		t.Fatalf("title after PATCH: got %q", patched.Title)
	}
	wantDate := time.Date(2026, 1, 22, 0, 0, 0, 0, time.UTC)
	if !patched.DueDate.Equal(wantDate) {
		t.Fatalf("due date after PATCH: got %s, want %s", patched.DueDate, wantDate)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/tasks/1", nil)
	deleteResp := httptest.NewRecorder()
	handler.TaskByID(deleteResp, deleteReq)

	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("DELETE status: got %d, want %d", deleteResp.Code, http.StatusNoContent)
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/1", nil)
	missingResp := httptest.NewRecorder()
	handler.TaskByID(missingResp, missingReq)

	if missingResp.Code != http.StatusNotFound {
		t.Fatalf("GET deleted task status: got %d, want %d", missingResp.Code, http.StatusNotFound)
	}
}

func TestTaskByIDInvalidPath(t *testing.T) {
	handler := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/bad", nil)
	resp := httptest.NewRecorder()
	handler.TaskByID(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", resp.Code, http.StatusBadRequest)
	}
}

func newTestHandler() *Handler {
	repo := task.NewRepo()
	taskService := service.NewTaskService(repo, nil, config.New())
	return NewHandler(taskService)
}
