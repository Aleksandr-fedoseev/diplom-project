// Package catalog — HTTP-обработчики каталога: категории, список объектов с фильтрами
// и поиском на карте, карточка объекта, редактирование и загрузка медиафайлов к объекту.
package catalog

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/catalog"
	"heritage-backend/internal/domain/media"
	"heritage-backend/internal/transport/http/httpx"
	mediahttp "heritage-backend/internal/transport/http/media"
	"heritage-backend/internal/transport/http/middleware"
)

type Handler struct {
	svc   catalog.Service
	media media.Service
	// upload — обработчик загрузки из пакета media; подключается здесь,
	// чтобы адрес был «вложенным»: POST /api/objects/{id}/media.
	upload http.Handler
}

func NewHandler(svc catalog.Service, mediaSvc media.Service, mediaHandler *mediahttp.Handler) *Handler {
	return &Handler{svc: svc, media: mediaSvc, upload: httpx.Handler(mediaHandler.Upload)}
}

// Routes подключается по адресу /api.
func (h *Handler) Routes(r chi.Router) {
	r.Method(http.MethodGet, "/categories", httpx.Handler(h.listCategories))
	r.Method(http.MethodGet, "/objects", httpx.Handler(h.list))
	r.Method(http.MethodGet, "/objects/{id}", httpx.Handler(h.get)) // здесь {id} — это slug

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(auth.RoleModerator))
		r.Method(http.MethodPost, "/objects", httpx.Handler(h.create))
		r.Method(http.MethodPut, "/objects/{id}", httpx.Handler(h.update))
		r.Method(http.MethodPost, "/objects/{id}/media", h.upload)
	})
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(auth.RoleAdmin))
		r.Method(http.MethodDelete, "/objects/{id}", httpx.Handler(h.delete))
		r.Method(http.MethodPost, "/categories", httpx.Handler(h.createCategory))
		r.Method(http.MethodPut, "/categories/{id}", httpx.Handler(h.updateCategory))
		r.Method(http.MethodDelete, "/categories/{id}", httpx.Handler(h.deleteCategory))
	})
}

// viewer — кто смотрит каталог: от этого зависит, видны ли неопубликованные объекты.
func viewer(r *http.Request) catalog.Viewer {
	p, ok := middleware.PrincipalFrom(r.Context())
	return catalog.Viewer{Moderator: ok && p.Role.AtLeast(auth.RoleModerator)}
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) error {
	cats, err := h.svc.ListCategories(r.Context())
	if err != nil {
		return err
	}
	out := make([]categoryResponse, len(cats))
	for i, c := range cats {
		out[i] = toCategoryResponse(c)
	}
	httpx.JSON(w, http.StatusOK, out)
	return nil
}

// list: GET /api/objects?category=modern&status=lost&q=банк&all=true
//
//	&lat=46.35&lng=48.03&radius=500   — объекты в радиусе (метры), ближайшие первыми
//	&bbox=48.02,46.34,48.05,46.36     — объекты в видимой области карты
//
// all=true — показать и неопубликованные (выполняется только для модератора).
func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	f, err := parseFilter(r)
	if err != nil {
		return err
	}
	objects, err := h.svc.List(r.Context(), f, viewer(r))
	if err != nil {
		return err
	}
	out := make([]objectResponse, len(objects))
	for i, o := range objects {
		out[i] = toObjectResponse(o)
	}
	httpx.JSON(w, http.StatusOK, out)
	return nil
}

// parseFilter разбирает параметры списка. Здесь только формат; что из этого разрешено, решает сервис.
func parseFilter(r *http.Request) (catalog.Filter, error) {
	q := r.URL.Query()
	f := catalog.Filter{
		Category:           q.Get("category"),
		Status:             q.Get("status"),
		Query:              strings.TrimSpace(q.Get("q")),
		IncludeUnpublished: q.Get("all") == "true",
	}

	if q.Has("lat") || q.Has("lng") {
		lat, err1 := strconv.ParseFloat(q.Get("lat"), 64)
		lng, err2 := strconv.ParseFloat(q.Get("lng"), 64)
		if err1 != nil || err2 != nil {
			return f, httpx.BadRequest("lat и lng должны быть числами")
		}
		f.Near = &catalog.Coordinates{Lat: lat, Lng: lng}
		f.RadiusM = 1000
		if v := q.Get("radius"); v != "" {
			radius, err := strconv.ParseFloat(v, 64)
			if err != nil || radius <= 0 || radius > 50000 {
				return f, httpx.BadRequest("radius — число метров от 1 до 50000")
			}
			f.RadiusM = radius
		}
	}

	if v := q.Get("bbox"); v != "" {
		parts := strings.Split(v, ",")
		if len(parts) != 4 {
			return f, httpx.BadRequest("bbox: minLng,minLat,maxLng,maxLat")
		}
		var b [4]float64
		for i, p := range parts {
			n, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
			if err != nil {
				return f, httpx.BadRequest("bbox: значения должны быть числами")
			}
			b[i] = n
		}
		f.BBox = &b
	}
	return f, nil
}

// get: GET /api/objects/{slug} — карточка объекта с медиафайлами.
func (h *Handler) get(w http.ResponseWriter, r *http.Request) error {
	o, err := h.svc.GetVisible(r.Context(), chi.URLParam(r, "id"), viewer(r))
	if err != nil {
		return err
	}
	files, err := h.media.ListByObject(r.Context(), o.ID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, cardResponse{Object: toObjectResponse(o), Media: mediahttp.ToFileResponses(files)})
	return nil
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var in objectRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	o, err := h.svc.Create(r.Context(), in.toInput())
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toObjectResponse(o))
	return nil
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	var in objectRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	o, err := h.svc.Update(r.Context(), id, in.toInput())
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toObjectResponse(o))
	return nil
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// categoryID читает {id} категории: в БД это INT, поэтому проверяем диапазон int32.
func categoryID(r *http.Request) (int32, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil || id <= 0 {
		return 0, httpx.BadRequest("некорректный id")
	}
	return int32(id), nil
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) error {
	var in categoryRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	c, err := h.svc.CreateCategory(r.Context(), catalog.CategoryInput{Slug: in.Slug, Name: in.Name})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toCategoryResponse(c))
	return nil
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) error {
	id, err := categoryID(r)
	if err != nil {
		return err
	}
	var in categoryRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	c, err := h.svc.UpdateCategory(r.Context(), id, catalog.CategoryInput{Slug: in.Slug, Name: in.Name})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toCategoryResponse(c))
	return nil
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) error {
	id, err := categoryID(r)
	if err != nil {
		return err
	}
	if err := h.svc.DeleteCategory(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
