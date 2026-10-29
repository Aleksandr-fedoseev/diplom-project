package submissions

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/catalog"
	"heritage-backend/internal/httpx"
)

const mb = 1 << 20

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes подключается по адресу /api.
func (h *Handler) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireRole(auth.RoleUser))
		r.Method(http.MethodPost, "/submissions", httpx.Handler(h.create))
		r.Method(http.MethodGet, "/submissions/{id}", httpx.Handler(h.get))
		r.Method(http.MethodPost, "/submissions/{id}/media", httpx.Handler(h.uploadMedia))
		r.Method(http.MethodGet, "/me/submissions", httpx.Handler(h.listMine))
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.RequireRole(auth.RoleModerator))
		r.Method(http.MethodGet, "/moderation/submissions", httpx.Handler(h.queue))
		r.Method(http.MethodPost, "/moderation/submissions/{id}/approve", httpx.Handler(h.approve))
		r.Method(http.MethodPost, "/moderation/submissions/{id}/reject", httpx.Handler(h.reject))
	})
}

// create: POST /api/submissions
// {"kind": "new_object", "title": "...", "address": "...", "coordinates": {"lat": 46.35, "lng": 48.03},
//
//	"period": "1890-е", "source": "ГААО, ф. 1", "description": "..."}
//
// Для дополнения: {"kind": "addition", "objectSlug": "birzha", "source": "...", "description": "..."}.
func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var in CreateInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := auth.FromContext(r.Context())
	sub, err := h.svc.Create(r.Context(), p.UserID, in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, sub)
	return nil
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) error {
	id, err := submissionID(r)
	if err != nil {
		return err
	}
	p, _ := auth.FromContext(r.Context())
	sub, err := h.svc.Get(r.Context(), id, Viewer{UserID: p.UserID, IsModerator: auth.AtLeast(p.Role, auth.RoleModerator)})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, sub)
	return nil
}

// uploadMedia: POST /api/submissions/{id}/media
// Форма multipart/form-data: поля kind, caption, yearTaken, source (необязательные, кроме kind) и file.
func (h *Handler) uploadMedia(w http.ResponseWriter, r *http.Request) error {
	id, err := submissionID(r)
	if err != nil {
		return err
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.svc.MaxUploadBytes()+mb) // +1 МБ на служебные поля формы
	// В памяти держим до 8 МБ, остальное Go сам временно складывает на диск.
	if err := r.ParseMultipartForm(8 * mb); err != nil {
		return httpx.BadRequest("не удалось прочитать форму: " + err.Error())
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	file, header, err := r.FormFile("file")
	if err != nil {
		return httpx.BadRequest("нет поля file")
	}
	defer file.Close()

	p, _ := auth.FromContext(r.Context())
	f, err := h.svc.UploadMedia(r.Context(), id, p.UserID, UploadInput{
		Kind:         r.FormValue("kind"),
		Caption:      r.FormValue("caption"),
		YearTaken:    r.FormValue("yearTaken"),
		Source:       r.FormValue("source"),
		OriginalName: header.Filename,
		Size:         header.Size,
		Body:         file,
	})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, f)
	return nil
}

func (h *Handler) listMine(w http.ResponseWriter, r *http.Request) error {
	p, _ := auth.FromContext(r.Context())
	list, err := h.svc.ListByAuthor(r.Context(), p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, nonNil(list))
	return nil
}

// queue: GET /api/moderation/submissions?status=pending&limit=50&offset=0
// Без status показываются ожидающие решения заявки, старые первыми.
func (h *Handler) queue(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	status := q.Get("status")
	if status == "" {
		status = StatusPending
	}
	limit, offset := int32(50), int32(0)
	if v, err := strconv.ParseInt(q.Get("limit"), 10, 32); err == nil && v > 0 {
		limit = int32(v)
	}
	if v, err := strconv.ParseInt(q.Get("offset"), 10, 32); err == nil && v >= 0 {
		offset = int32(v)
	}
	list, err := h.svc.ListByStatus(r.Context(), status, limit, offset)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, nonNil(list))
	return nil
}

// approve: POST /api/moderation/submissions/{id}/approve
// Для новой заявки тело — поля объекта каталога: {"slug": "dom-ivanova", "category": "manor",
// "protectionStatus": "identified", ...}; незаполненные берутся из заявки. Для дополнения тело не нужно.
func (h *Handler) approve(w http.ResponseWriter, r *http.Request) error {
	id, err := submissionID(r)
	if err != nil {
		return err
	}
	var in catalog.ObjectInput
	if r.ContentLength != 0 { // пустое тело допустимо — для дополнения
		if err := httpx.DecodeJSON(r, &in); err != nil {
			return err
		}
	}
	p, _ := auth.FromContext(r.Context())
	sub, err := h.svc.Approve(r.Context(), id, Moderator{UserID: p.UserID, Role: p.Role}, in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, sub)
	return nil
}

// reject: POST /api/moderation/submissions/{id}/reject {"reason": "...", "blockAuthor": false}
func (h *Handler) reject(w http.ResponseWriter, r *http.Request) error {
	id, err := submissionID(r)
	if err != nil {
		return err
	}
	var in RejectInput
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := auth.FromContext(r.Context())
	sub, err := h.svc.Reject(r.Context(), id, Moderator{UserID: p.UserID, Role: p.Role}, in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, sub)
	return nil
}

func submissionID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, httpx.BadRequest("некорректный id заявки")
	}
	return id, nil
}

// nonNil нужен, чтобы пустой список уходил клиенту как [], а не null.
func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
