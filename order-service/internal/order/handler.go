package order

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	repo       Repository
	userClient UserClient
}

func NewHandler(repo Repository, userClient UserClient) *Handler {
	return &Handler{repo: repo, userClient: userClient}
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idText, full, ok := parseOrderPath(r.URL.Path)
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "invalid order path")
		return
	}

	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	order, found := h.repo.GetByID(id)
	if !found {
		writeJSONError(w, http.StatusNotFound, "order not found")
		return
	}

	if !full {
		writeJSON(w, order)
		return
	}

	user, err := h.userClient.GetUser(order.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			writeJSONError(w, http.StatusBadGateway, "user not found in user-service")
			return
		}
		writeJSONError(w, http.StatusBadGateway, "failed to get user from user-service")
		return
	}

	writeJSON(w, OrderWithUser{
		Order: order,
		User:  user,
	})
}

func parseOrderPath(path string) (id string, full bool, ok bool) {
	rest, ok := strings.CutPrefix(path, "/orders/")
	if !ok || rest == "" {
		return "", false, false
	}

	parts := strings.Split(rest, "/")
	switch len(parts) {
	case 1:
		return parts[0], false, parts[0] != ""
	case 2:
		return parts[0], parts[1] == "full", parts[0] != "" && parts[1] == "full"
	default:
		return "", false, false
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	if err := json.NewEncoder(w).Encode(value); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to encode response")
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
