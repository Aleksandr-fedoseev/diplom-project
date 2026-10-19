package routes

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes подключается по адресу /api.
func (h *Handler) Routes(r chi.Router) {
	r.Method(http.MethodGet, "/routes", httpx.Handler(h.listThematic))
	r.Method(http.MethodGet, "/routes/{slug}", httpx.Handler(h.get))
	r.Method(http.MethodPost, "/routes/build", httpx.Handler(h.build))

	r.With(auth.RequireRole(auth.RoleAdmin)).Method(http.MethodPost, "/routes", httpx.Handler(h.createThematic))

	r.Group(func(r chi.Router) {
		r.Use(auth.RequireRole(auth.RoleUser))
		r.Method(http.MethodGet, "/me/routes", httpx.Handler(h.listMine))
		r.Method(http.MethodPost, "/me/routes", httpx.Handler(h.saveMine))
		r.Method(http.MethodDelete, "/routes/{slug}", httpx.Handler(h.delete)) // здесь {slug} — это id
	})
}

func (h *Handler) listThematic(w http.ResponseWriter, r *http.Request) error {
	list, err := h.svc.ListThematic(r.Context())
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, nonNil(list))
	return nil
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) error {
	route, err := h.svc.Get(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		return err
	}
	// Чужой сохранённый маршрут не показываем.
	if route.OwnerID != nil {
		p, ok := auth.FromContext(r.Context())
		if !ok || (p.UserID != *route.OwnerID && p.Role != auth.RoleAdmin) {
			return httpx.NotFound("маршрут не найден")
		}
	}
	httpx.JSON(w, http.StatusOK, route)
	return nil
}

// build: POST /api/routes/build {"objects": ["birzha", "azovo-donskoy-bank"]}
// Строит маршрут для показа на карте, ничего не сохраняет. Доступно и гостям.
func (h *Handler) build(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Objects []string `json:"objects"`
	}
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	route, err := h.svc.Build(r.Context(), in.Objects)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, route)
	return nil
}

func (h *Handler) createThematic(w http.ResponseWriter, r *http.Request) error {
	var in SaveInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	route, err := h.svc.Save(r.Context(), in, nil)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, route)
	return nil
}

func (h *Handler) listMine(w http.ResponseWriter, r *http.Request) error {
	p, _ := auth.FromContext(r.Context())
	list, err := h.svc.ListByOwner(r.Context(), p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, nonNil(list))
	return nil
}

func (h *Handler) saveMine(w http.ResponseWriter, r *http.Request) error {
	var in SaveInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := auth.FromContext(r.Context())
	route, err := h.svc.Save(r.Context(), in, &p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, route)
	return nil
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "slug"), 10, 64)
	if err != nil {
		return httpx.BadRequest("некорректный id маршрута")
	}
	p, _ := auth.FromContext(r.Context())
	if err := h.svc.Delete(r.Context(), id, p.UserID, p.Role == auth.RoleAdmin); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// nonNil нужен, чтобы пустой список уходил клиенту как [], а не null.
func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
