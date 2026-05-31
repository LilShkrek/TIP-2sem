package user

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerGetUser(t *testing.T) {
	handler := NewHandler(NewMemoryRepository())

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "existing user",
			method:     http.MethodGet,
			path:       "/users/1",
			wantStatus: http.StatusOK,
			wantBody:   `"name":"Иван Иванов"`,
		},
		{
			name:       "invalid id",
			method:     http.MethodGet,
			path:       "/users/abc",
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid user id",
		},
		{
			name:       "missing user",
			method:     http.MethodGet,
			path:       "/users/99",
			wantStatus: http.StatusNotFound,
			wantBody:   "user not found",
		},
		{
			name:       "unsupported method",
			method:     http.MethodPost,
			path:       "/users/1",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.GetUser(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("body = %q, want substring %q", rec.Body.String(), tt.wantBody)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Fatalf("content type = %q", got)
			}
		})
	}
}
