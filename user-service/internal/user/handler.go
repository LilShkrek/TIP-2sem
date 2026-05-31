package user

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idText, ok := strings.CutPrefix(r.URL.Path, "/users/")
	if !ok || idText == "" || strings.Contains(idText, "/") {
		writeJSONError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, found := h.repo.GetByID(id)
	if !found {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	if err := json.NewEncoder(w).Encode(user); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to encode response")
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
