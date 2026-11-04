// Package auth — HTTP-обработчики модуля пользователей: регистрация, вход,
// текущий пользователь и управление пользователями.
package auth

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/transport/http/httpx"
	"heritage-backend/internal/transport/http/middleware"
)

// Ограничения против подбора пароля и массовой регистрации. По IP — против одного источника,
// по email — против перебора пароля к одной учётной записи с разных адресов.
const (
	loginPerIP       = 10
	loginPerEmail    = 10
	registerPerIP    = 5
	loginWindow      = time.Minute
	loginEmailWindow = 15 * time.Minute
	registerWindow   = time.Hour
)

type Handler struct {
	svc          auth.Service
	loginByIP    *middleware.RateLimiter
	loginByEmail *middleware.RateLimiter
	registerByIP *middleware.RateLimiter
}

func NewHandler(svc auth.Service) *Handler {
	return &Handler{
		svc:          svc,
		loginByIP:    middleware.NewRateLimiter(loginPerIP, loginWindow),
		loginByEmail: middleware.NewRateLimiter(loginPerEmail, loginEmailWindow),
		registerByIP: middleware.NewRateLimiter(registerPerIP, registerWindow),
	}
}

// Routes подключается по адресу /api/auth.
func (h *Handler) Routes(r chi.Router) {
	r.With(middleware.LimitByIP(h.registerByIP)).Method(http.MethodPost, "/register", httpx.Handler(h.register))
	r.With(middleware.LimitByIP(h.loginByIP)).Method(http.MethodPost, "/login", httpx.Handler(h.login))
	r.With(middleware.RequireRole(auth.RoleUser)).Method(http.MethodGet, "/me", httpx.Handler(h.me))
}

// AdminRoutes подключается по адресу /api/admin/users.
func (h *Handler) AdminRoutes(r chi.Router) {
	r.With(middleware.RequireRole(auth.RoleAdmin)).Method(http.MethodGet, "/", httpx.Handler(h.list))
	r.With(middleware.RequireRole(auth.RoleAdmin)).Method(http.MethodPut, "/{id}/role", httpx.Handler(h.setRole))
	r.With(middleware.RequireRole(auth.RoleModerator)).Method(http.MethodPut, "/{id}/blocked", httpx.Handler(h.setBlocked))
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) error {
	var in registerRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	u, err := h.svc.Register(r.Context(), auth.Registration{Email: in.Email, Password: in.Password, Name: in.Name})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toUserResponse(u))
	return nil
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) error {
	var in loginRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	if ok, retry := h.loginByEmail.Allow(auth.NormalizeEmail(in.Email)); !ok {
		return httpx.TooManyRequests(w, "слишком много попыток входа, повторите позже", retry)
	}
	token, u, err := h.svc.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, loginResponse{Token: token, User: toUserResponse(u)})
	return nil
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) error {
	p, _ := middleware.PrincipalFrom(r.Context())
	u, err := h.svc.GetByID(r.Context(), p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toUserResponse(u))
	return nil
}

// list: GET /api/admin/users?limit=50&offset=0
func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	limit, offset := 50, 0
	if v, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && v > 0 && v <= 200 {
		limit = v
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && v >= 0 {
		offset = v
	}
	users, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		return err
	}
	out := make([]userResponse, len(users))
	for i, u := range users {
		out[i] = toUserResponse(u)
	}
	httpx.JSON(w, http.StatusOK, out)
	return nil
}

// setRole: PUT /api/admin/users/{id}/role {"role": "moderator"}
func (h *Handler) setRole(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	var in roleRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := middleware.PrincipalFrom(r.Context())
	u, err := h.svc.SetRole(r.Context(), p, id, auth.Role(in.Role))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toUserResponse(u))
	return nil
}

// setBlocked: PUT /api/admin/users/{id}/blocked {"blocked": true}
func (h *Handler) setBlocked(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	var in blockedRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := middleware.PrincipalFrom(r.Context())
	u, err := h.svc.SetBlocked(r.Context(), p, id, in.Blocked)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toUserResponse(u))
	return nil
}
