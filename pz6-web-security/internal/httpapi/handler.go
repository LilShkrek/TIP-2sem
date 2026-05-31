package httpapi

import (
	"html/template"
	"net/http"
	"strings"

	"example.com/pz6-web-security/internal/auth"
	"example.com/pz6-web-security/internal/store"
)

type Handler struct {
	store       *store.Store
	profileTmpl *template.Template
	helloTmpl   *template.Template
}

func NewHandler(s *store.Store) (*Handler, error) {
	profileTmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		return nil, err
	}

	helloTmpl, err := template.ParseFiles("templates/hello.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		store:       s,
		profileTmpl: profileTmpl,
		helloTmpl:   helloTmpl,
	}, nil
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID, err := auth.RandomToken(16)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	csrfToken, err := auth.RandomToken(16)
	if err != nil {
		http.Error(w, "failed to create csrf token", http.StatusInternalServerError)
		return
	}

	h.store.Save(&store.UserProfile{
		SessionID: sessionID,
		Name:      "Студент",
		CSRFToken: csrfToken,
	})

	auth.SetSessionCookie(w, sessionID)
	http.Redirect(w, r, "/profile", http.StatusFound)
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	sessionID, err := auth.ReadSessionCookie(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profile, ok := h.store.Get(sessionID)
	if !ok {
		http.Error(w, "session not found", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		data := struct {
			Name      string
			CSRFToken string
		}{
			Name:      profile.Name,
			CSRFToken: profile.CSRFToken,
		}

		if err := h.profileTmpl.Execute(w, data); err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		tokenFromForm := r.FormValue("csrf_token")
		if tokenFromForm == "" || !auth.EqualToken(tokenFromForm, profile.CSRFToken) {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		if !h.store.UpdateName(sessionID, name) {
			http.Error(w, "session not found", http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, "/hello", http.StatusFound)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) Hello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID, err := auth.ReadSessionCookie(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profile, ok := h.store.Get(sessionID)
	if !ok {
		http.Error(w, "session not found", http.StatusUnauthorized)
		return
	}

	data := struct {
		Name string
	}{
		Name: profile.Name,
	}

	if err := h.helloTmpl.Execute(w, data); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
}
