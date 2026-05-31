package order

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeUserClient struct {
	user UserDTO
	err  error
}

func (c fakeUserClient) GetUser(id int64) (UserDTO, error) {
	return c.user, c.err
}

func TestHandlerGetOrder(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		client     UserClient
		wantStatus int
		wantBody   string
	}{
		{
			name:       "existing order",
			method:     http.MethodGet,
			path:       "/orders/101",
			client:     fakeUserClient{},
			wantStatus: http.StatusOK,
			wantBody:   `"item":"Ноутбук"`,
		},
		{
			name:   "full order",
			method: http.MethodGet,
			path:   "/orders/101/full",
			client: fakeUserClient{user: UserDTO{
				ID:    1,
				Name:  "Иван Иванов",
				Email: "ivan@example.com",
			}},
			wantStatus: http.StatusOK,
			wantBody:   `"user":{"id":1,"name":"Иван Иванов","email":"ivan@example.com"}`,
		},
		{
			name:       "invalid id",
			method:     http.MethodGet,
			path:       "/orders/abc",
			client:     fakeUserClient{},
			wantStatus: http.StatusBadRequest,
			wantBody:   "invalid order id",
		},
		{
			name:       "missing order",
			method:     http.MethodGet,
			path:       "/orders/999",
			client:     fakeUserClient{},
			wantStatus: http.StatusNotFound,
			wantBody:   "order not found",
		},
		{
			name:       "unsupported method",
			method:     http.MethodPost,
			path:       "/orders/101",
			client:     fakeUserClient{},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "user service error",
			method:     http.MethodGet,
			path:       "/orders/101/full",
			client:     fakeUserClient{err: errors.New("service unavailable")},
			wantStatus: http.StatusBadGateway,
			wantBody:   "failed to get user from user-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(NewMemoryRepository(), tt.client)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.GetOrder(rec, req)

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
