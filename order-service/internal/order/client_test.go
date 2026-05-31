package order

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPUserClientGetUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/1" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1,"name":"Иван Иванов","email":"ivan@example.com"}`))
	}))
	defer server.Close()

	client := NewUserClient(server.URL)

	user, err := client.GetUser(1)
	if err != nil {
		t.Fatalf("GetUser returned error: %v", err)
	}
	if user.ID != 1 || user.Email != "ivan@example.com" {
		t.Fatalf("user = %+v", user)
	}
}

func TestHTTPUserClientGetUserNotFound(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	client := NewUserClient(server.URL)

	_, err := client.GetUser(99)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("error = %v, want ErrUserNotFound", err)
	}
}

func TestHTTPUserClientGetUserBadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewUserClient(server.URL)

	_, err := client.GetUser(1)
	if !errors.Is(err, ErrUserServiceRequest) {
		t.Fatalf("error = %v, want ErrUserServiceRequest", err)
	}
}
