package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"example.com/pz4-monitoring/internal/student"
)

type Handler struct {
	repo *student.Repo
}

func NewHandler(repo *student.Repo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (h *Handler) GetStudentByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idText := strings.TrimPrefix(r.URL.Path, "/students/")
	if idText == "" || idText == r.URL.Path {
		http.Error(w, "student id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil {
		http.Error(w, "invalid student id", http.StatusBadRequest)
		return
	}

	st, err := h.repo.GetByID(id)
	if errors.Is(err, student.ErrStudentNotFound) {
		http.Error(w, "student not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(st)
}
