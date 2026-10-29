package catalog

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/httpx"
	"heritage-backend/internal/media"
)

type Handler struct {
	svc   *Service
	media *media.Service
	// upload — обработчик загрузки из пакета media; подключается сюда,
	// чтобы адрес был «вложенным»: POST /api/objects/{id}/media.
	upload http.Handler
}

func NewHandler(svc *Service, mediaSvc *media.Service, mediaHandler *media.Handler) *Handler {
	return &Handler{svc: svc, media: mediaSvc, upload: httpx.Handler(mediaHandler.Upload)}
}

// Routes подключается по адресу /api.
func (h *Handler) Routes(r chi.Router) {
	r.Method(http.MethodGet, "/categories", httpx.Handler(h.listCategories))
	r.Method(http.MethodGet, "/objects", httpx.Handler(h.list))
	r.Method(http.MethodGet, "/objects/{id}", httpx.Handler(h.get)) // здесь {id} — это slug

	r.Group(func(r chi.Router) {
		r.Use(auth.RequireRole(auth.RoleModerator))
		r.Method(http.MethodPost, "/objects", httpx.Handler(h.create))
		r.Method(http.MethodPut, "/objects/{id}", httpx.Handler(h.update))
		r.Method(http.MethodPost, "/objects/{id}/media", h.upload)
	})
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireRole(auth.RoleAdmin))
		r.Method(http.MethodDelete, "/objects/{id}", httpx.Handler(h.delete))
		r.Method(http.MethodPost, "/categories", httpx.Handler(h.createCategory))
		r.Method(http.MethodPut, "/categories/{id}", httpx.Handler(h.updateCategory))
		r.Method(http.MethodDelete, "/categories/{id}", httpx.Handler(h.deleteCategory))
	})
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) error {
	cats, err := h.svc.ListCategories(r.Context())
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, cats)
	return nil
}

// list: GET /api/objects?category=modern&status=lost&q=банк
//
//	&lat=46.35&lng=48.03&radius=500   — объекты в радиусе (метры)
//	&bbox=48.02,46.34,48.05,46.36     — объекты в видимой области карты
func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	f := Filter{
		Category: q.Get("category"),
		Status:   q.Get("status"),
		Query:    strings.TrimSpace(q.Get("q")),
	}

	// Неопубликованные объекты видят только модераторы.
	if p, ok := auth.FromContext(r.Context()); ok && p.Role != auth.RoleUser && q.Get("all") == "true" {
		f.IncludeUnpublished = true
	}

	if q.Has("lat") || q.Has("lng") {
		lat, err1 := strconv.ParseFloat(q.Get("lat"), 64)
		lng, err2 := strconv.ParseFloat(q.Get("lng"), 64)
		if err1 != nil || err2 != nil {
			return httpx.BadRequest("lat и lng должны быть числами")
		}
		f.Near = &Coordinates{Lat: lat, Lng: lng}
		f.RadiusM = 1000
		if v := q.Get("radius"); v != "" {
			radius, err := strconv.ParseFloat(v, 64)
			if err != nil || radius <= 0 || radius > 50000 {
				return httpx.BadRequest("radius — число метров от 1 до 50000")
			}
			f.RadiusM = radius
		}
	}

	if v := q.Get("bbox"); v != "" {
		parts := strings.Split(v, ",")
		if len(parts) != 4 {
			return httpx.BadRequest("bbox: minLng,minLat,maxLng,maxLat")
		}
		var b [4]float64
		for i, p := range parts {
			n, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
			if err != nil {
				return httpx.BadRequest("bbox: значения должны быть числами")
			}
			b[i] = n
		}
		f.BBox = &b
	}

	objects, err := h.svc.List(r.Context(), f)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, objects)
	return nil
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) error {
	o, err := h.svc.GetBySlug(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		return err
	}
	if !o.IsPublished {
		if p, ok := auth.FromContext(r.Context()); !ok || p.Role == auth.RoleUser {
			return httpx.NotFound("объект не найден")
		}
	}
	files, err := h.media.ListByObject(r.Context(), o.ID)
	if err != nil {
		return err
	}
	if files == nil {
		files = []media.File{}
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"object": o, "media": files})
	return nil
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var in ObjectInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	o, err := h.svc.Create(r.Context(), in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, o)
	return nil
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return httpx.BadRequest("некорректный id")
	}
	var in ObjectInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	o, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, o)
	return nil
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return httpx.BadRequest("некорректный id")
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) error {
	var in CategoryInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	c, err := h.svc.CreateCategory(r.Context(), in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, c)
	return nil
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		return httpx.BadRequest("некорректный id")
	}
	var in CategoryInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	c, err := h.svc.UpdateCategory(r.Context(), int32(id), in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, c)
	return nil
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		return httpx.BadRequest("некорректный id")
	}
	if err := h.svc.DeleteCategory(r.Context(), int32(id)); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
