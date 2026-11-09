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
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(auth.RoleUser))
		r.Method(http.MethodGet, "/me", httpx.Handler(h.me))
		r.Method(http.MethodPut, "/me", httpx.Handler(h.updateMe))
		r.Method(http.MethodPut, "/me/password", httpx.Handler(h.changePassword))
		r.Method(http.MethodPut, "/me/avatar", httpx.Handler(h.setAvatar))
		r.Method(http.MethodDelete, "/me/avatar", httpx.Handler(h.removeAvatar))
	})
}

// UserRoutes подключается по адресу /api/users: открытые сведения о пользователях.
func (h *Handler) UserRoutes(r chi.Router) {
	r.Method(http.MethodGet, "/{id}/avatar", httpx.Handler(h.avatar))
}

// AdminRoutes подключается по адресу /api/admin/users.
func (h *Handler) AdminRoutes(r chi.Router) {
	// Управление пользователями — прецедент администратора. Модератор блокирует автора
	// только вместе с отказом по заявке (POST /moderation/submissions/{id}/reject).
	r.Use(middleware.RequireRole(auth.RoleAdmin))
	r.Method(http.MethodGet, "/", httpx.Handler(h.list))
	r.Method(http.MethodPut, "/{id}/role", httpx.Handler(h.setRole))
	r.Method(http.MethodPut, "/{id}/blocked", httpx.Handler(h.setBlocked))
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

// updateMe: PUT /api/auth/me {"name": "..."} — своё имя.
func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request) error {
	var in profileRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := middleware.PrincipalFrom(r.Context())
	u, err := h.svc.UpdateName(r.Context(), p.UserID, in.Name)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toUserResponse(u))
	return nil
}

// changePassword: PUT /api/auth/me/password {"currentPassword", "newPassword"}.
// Подбор текущего пароля ограничен так же, как вход.
func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) error {
	var in passwordRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := middleware.PrincipalFrom(r.Context())
	if ok, retry := h.loginByEmail.Allow("password:" + strconv.FormatInt(p.UserID, 10)); !ok {
		return httpx.TooManyRequests(w, "слишком много попыток, повторите позже", retry)
	}
	if err := h.svc.ChangePassword(r.Context(), p.UserID, in.CurrentPassword, in.NewPassword); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// setAvatar: PUT /api/auth/me/avatar, форма multipart с полем file.
func (h *Handler) setAvatar(w http.ResponseWriter, r *http.Request) error {
	r.Body = http.MaxBytesReader(w, r.Body, auth.MaxAvatarBytes+64<<10) // +64 КБ на служебные поля формы
	file, header, err := r.FormFile("file")
	if err != nil {
		return httpx.BadRequest("нужна картинка в поле file, не больше 2 МБ")
	}
	defer file.Close()
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	p, _ := middleware.PrincipalFrom(r.Context())
	u, err := h.svc.SetAvatar(r.Context(), p.UserID, file, header.Size)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toUserResponse(u))
	return nil
}

func (h *Handler) removeAvatar(w http.ResponseWriter, r *http.Request) error {
	p, _ := middleware.PrincipalFrom(r.Context())
	u, err := h.svc.RemoveAvatar(r.Context(), p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toUserResponse(u))
	return nil
}

// avatar: GET /api/users/{id}/avatar — картинка аватара. Адрес меняется вместе с файлом
// (параметр v), поэтому браузеру можно долго хранить её в кэше.
func (h *Handler) avatar(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	body, contentType, err := h.svc.OpenAvatar(r.Context(), id)
	if err != nil {
		return err
	}
	defer body.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=604800")
	http.ServeContent(w, r, "", time.Time{}, body)
	return nil
}
