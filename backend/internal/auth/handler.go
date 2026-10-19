package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes подключается в server.go по адресу /api/auth.
func (h *Handler) Routes(r chi.Router) {
	r.Method(http.MethodPost, "/register", httpx.Handler(h.register))
	r.Method(http.MethodPost, "/login", httpx.Handler(h.login))
	r.With(RequireRole(RoleUser)).Method(http.MethodGet, "/me", httpx.Handler(h.me))
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) error {
	var in RegisterInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	u, err := h.svc.Register(r.Context(), in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, u)
	return nil
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	token, u, err := h.svc.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"token": token, "user": u})
	return nil
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) error {
	p, _ := FromContext(r.Context())
	u, err := h.svc.GetByID(r.Context(), p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, u)
	return nil
}
