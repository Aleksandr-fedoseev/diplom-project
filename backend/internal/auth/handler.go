package auth

import (
	"net/http"
	"strconv"

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

// AdminRoutes подключается в server.go по адресу /api/admin/users.
func (h *Handler) AdminRoutes(r chi.Router) {
	r.With(RequireRole(RoleAdmin)).Method(http.MethodGet, "/", httpx.Handler(h.list))
	r.With(RequireRole(RoleAdmin)).Method(http.MethodPut, "/{id}/role", httpx.Handler(h.setRole))
	r.With(RequireRole(RoleModerator)).Method(http.MethodPut, "/{id}/blocked", httpx.Handler(h.setBlocked))
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

// list: GET /api/admin/users?limit=50&offset=0
func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	limit, offset := int32(50), int32(0)
	if v, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32); err == nil && v > 0 && v <= 200 {
		limit = int32(v)
	}
	if v, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32); err == nil && v >= 0 {
		offset = int32(v)
	}
	users, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, users)
	return nil
}

// setRole: PUT /api/admin/users/{id}/role {"role": "moderator"}
func (h *Handler) setRole(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return httpx.BadRequest("некорректный id")
	}
	var in struct {
		Role string `json:"role"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := FromContext(r.Context())
	u, err := h.svc.SetRole(r.Context(), p.UserID, id, in.Role)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, u)
	return nil
}

// setBlocked: PUT /api/admin/users/{id}/blocked {"blocked": true}
func (h *Handler) setBlocked(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return httpx.BadRequest("некорректный id")
	}
	var in struct {
		Blocked bool `json:"blocked"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := FromContext(r.Context())
	u, err := h.svc.SetBlocked(r.Context(), p.UserID, p.Role, id, in.Blocked)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, u)
	return nil
}
