package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/pz12-rest-vs-graphql/internal/task"
)

func TestRESTTaskScenario(t *testing.T) {
	handler := newTestHandler()

	listResponse := request(handler, http.MethodGet, "/v1/tasks", nil)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("GET /v1/tasks status = %d", listResponse.Code)
	}

	var list []task.Task
	decodeBody(t, listResponse, &list)
	if len(list) != 2 {
		t.Fatalf("tasks length = %d, want 2", len(list))
	}

	detailResponse := request(handler, http.MethodGet, "/v1/tasks/t_001", nil)
	if detailResponse.Code != http.StatusOK {
		t.Fatalf("GET /v1/tasks/t_001 status = %d", detailResponse.Code)
	}

	var detail task.Task
	decodeBody(t, detailResponse, &detail)
	if detail.Description == "" {
		t.Fatal("detail response must include description")
	}

	createResponse := request(handler, http.MethodPost, "/v1/tasks", []byte(`{"title":"Новая задача","description":"REST"}`))
	if createResponse.Code != http.StatusCreated {
		t.Fatalf("POST /v1/tasks status = %d", createResponse.Code)
	}

	var created task.Task
	decodeBody(t, createResponse, &created)
	if created.ID != "t_003" {
		t.Fatalf("created ID = %q, want t_003", created.ID)
	}

	updateResponse := request(handler, http.MethodPatch, "/v1/tasks/t_003", []byte(`{"done":true}`))
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("PATCH /v1/tasks/t_003 status = %d", updateResponse.Code)
	}

	var updated task.Task
	decodeBody(t, updateResponse, &updated)
	if !updated.Done {
		t.Fatal("updated task must be done")
	}
}

func TestRESTNotFound(t *testing.T) {
	handler := newTestHandler()

	response := request(handler, http.MethodGet, "/v1/tasks/unknown", nil)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}

	var body map[string]string
	decodeBody(t, response, &body)
	if body["error"] != "task not found" {
		t.Fatalf("error = %q", body["error"])
	}
}

func newTestHandler() http.Handler {
	service := task.NewService(task.NewMemoryRepository(task.DefaultSeed()))
	mux := http.NewServeMux()
	NewHandler(service).Register(mux)
	return mux
}

func request(handler http.Handler, method string, target string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}

func decodeBody(t *testing.T, response *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(response.Body).Decode(dst); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}
