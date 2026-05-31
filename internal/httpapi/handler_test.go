package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	handler := NewHandler(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected ok status in response, got %q", rec.Body.String())
	}
}

func TestHealthRejectsNonGet(t *testing.T) {
	handler := NewHandler(nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestGetStudentByIDRequiresID(t *testing.T) {
	handler := NewHandler(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/students", nil)
	rec := httptest.NewRecorder()

	handler.GetStudentByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetStudentByIDRejectsInvalidID(t *testing.T) {
	handler := NewHandler(nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/students?id=1%20OR%201=1", nil)
	rec := httptest.NewRecorder()

	handler.GetStudentByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
