// Package route — HTTP-обработчики модуля маршрутов: тематические маршруты,
// построение маршрута на карте и сохранённые маршруты пользователей.
package route

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/route"
	"heritage-backend/internal/transport/http/httpx"
	"heritage-backend/internal/transport/http/middleware"
)

type Handler struct {
	svc route.Service
}

func NewHandler(svc route.Service) *Handler { return &Handler{svc: svc} }

// Routes подключается по адресу /api.
// GET /routes/{slug} и DELETE /routes/{id} делят один сегмент адреса: chi хранит имена
// параметров отдельно для каждого метода, поэтому у них могут быть разные имена.
func (h *Handler) Routes(r chi.Router) {
	r.Method(http.MethodGet, "/routes", httpx.Handler(h.listThematic))
	r.Method(http.MethodGet, "/routes/{slug}", httpx.Handler(h.get))
	r.Method(http.MethodPost, "/routes/build", httpx.Handler(h.build))

	r.With(middleware.RequireRole(auth.RoleAdmin)).Method(http.MethodPost, "/routes", httpx.Handler(h.createThematic))

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(auth.RoleUser))
		r.Method(http.MethodGet, "/me/routes", httpx.Handler(h.listMine))
		r.Method(http.MethodPost, "/me/routes", httpx.Handler(h.saveMine))
		r.Method(http.MethodDelete, "/routes/{id}", httpx.Handler(h.delete))
	})
}

// actor — текущий пользователь в терминах модуля маршрутов; nil для гостя.
func actor(ctx context.Context) *route.Actor {
	p, ok := middleware.PrincipalFrom(ctx)
	if !ok {
		return nil
	}
	return &route.Actor{UserID: p.UserID, IsAdmin: p.Role == auth.RoleAdmin}
}

func (h *Handler) listThematic(w http.ResponseWriter, r *http.Request) error {
	list, err := h.svc.ListThematic(r.Context())
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toRouteList(list))
	return nil
}

// get: GET /api/routes/{slug}. Чужой сохранённый маршрут сервис не отдаёт (404).
func (h *Handler) get(w http.ResponseWriter, r *http.Request) error {
	rt, err := h.svc.Get(r.Context(), chi.URLParam(r, "slug"), actor(r.Context()))
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toRouteResponse(rt))
	return nil
}

// build: POST /api/routes/build {"objects": ["birzha", "azovo-donskoy-bank"]}
// Строит маршрут для показа на карте, ничего не сохраняет. Доступно и гостям.
func (h *Handler) build(w http.ResponseWriter, r *http.Request) error {
	var in buildRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	rt, err := h.svc.Build(r.Context(), in.Objects)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toRouteResponse(rt))
	return nil
}

// createThematic: POST /api/routes {"slug", "title", "description", "objects"} — только администратор.
func (h *Handler) createThematic(w http.ResponseWriter, r *http.Request) error {
	var in saveRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	rt, err := h.svc.Save(r.Context(), in.toInput(), nil)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toRouteResponse(rt))
	return nil
}

func (h *Handler) listMine(w http.ResponseWriter, r *http.Request) error {
	p, _ := middleware.PrincipalFrom(r.Context())
	list, err := h.svc.ListByOwner(r.Context(), p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toRouteList(list))
	return nil
}

// saveMine: POST /api/me/routes {"title", "description", "objects"}; slug сервис создаёт сам.
func (h *Handler) saveMine(w http.ResponseWriter, r *http.Request) error {
	var in saveRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := middleware.PrincipalFrom(r.Context())
	rt, err := h.svc.Save(r.Context(), in.toInput(), &p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toRouteResponse(rt))
	return nil
}

// delete: DELETE /api/routes/{id} — владелец или администратор.
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	if err := h.svc.Delete(r.Context(), id, *actor(r.Context())); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
