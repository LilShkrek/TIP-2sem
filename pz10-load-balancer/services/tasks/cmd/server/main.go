package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func main() {
	instanceID := getenvDefault("INSTANCE_ID", "tasks-unknown")
	port := getenvDefault("APP_PORT", "8082")

	tasks := []Task{
		{ID: 1, Title: "Изучить NGINX Load Balancer"},
		{ID: 2, Title: "Проверить горизонтальное масштабирование"},
		{ID: 3, Title: "Убедиться в работе X-Instance-ID"},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":   "ok",
			"instance": instanceID,
		})
	})

	mux.HandleFunc("GET /v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, tasks)
	})

	mux.HandleFunc("GET /whoami", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"instance": instanceID,
		})
	})

	handler := instanceHeaderMiddleware(instanceID, loggingMiddleware(instanceID, mux))

	addr := ":" + port
	log.Printf("tasks service started on %s instance=%s", addr, instanceID)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func getenvDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response error: %v", err)
	}
}

func instanceHeaderMiddleware(instanceID string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header()["X-Instance-ID"] = []string{instanceID}
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(instanceID string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)
		log.Printf("method=%s path=%s status=%d instance=%s", r.Method, r.URL.Path, rw.statusCode, instanceID)
	})
}
