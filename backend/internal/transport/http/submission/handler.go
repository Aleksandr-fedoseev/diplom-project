// Package submission — HTTP-обработчики модуля заявок: подача, файлы, просмотр
// и модерация (очередь, одобрение, отказ).
package submission

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/submission"
	"heritage-backend/internal/transport/http/httpx"
	"heritage-backend/internal/transport/http/middleware"
)

const mb = 1 << 20

type Handler struct {
	svc            submission.Service
	maxUploadBytes int64 // предел размера файла (MAX_UPLOAD_MB), к нему добавляется 1 МБ на поля формы
}

func NewHandler(svc submission.Service, maxUploadBytes int64) *Handler {
	return &Handler{svc: svc, maxUploadBytes: maxUploadBytes}
}

// Routes подключается по адресу /api.
func (h *Handler) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(auth.RoleUser))
		r.Method(http.MethodPost, "/submissions", httpx.Handler(h.create))
		r.Method(http.MethodGet, "/submissions/{id}", httpx.Handler(h.get))
		r.Method(http.MethodPost, "/submissions/{id}/media", httpx.Handler(h.uploadMedia))
		r.Method(http.MethodGet, "/me/submissions", httpx.Handler(h.listMine))
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireRole(auth.RoleModerator))
		r.Method(http.MethodGet, "/moderation/submissions", httpx.Handler(h.queue))
		r.Method(http.MethodPost, "/moderation/submissions/{id}/approve", httpx.Handler(h.approve))
		r.Method(http.MethodPost, "/moderation/submissions/{id}/reject", httpx.Handler(h.reject))
	})
}

// moderator — текущий пользователь как модератор: роль нужна модулю пользователей при блокировке автора.
func moderator(p auth.Principal) submission.Moderator {
	return submission.Moderator{UserID: p.UserID, Role: string(p.Role)}
}

// create: POST /api/submissions
// {"kind": "new_object", "title": "...", "address": "...", "coordinates": {"lat": 46.35, "lng": 48.03},
// "period": "1890-е", "source": "ГААО, ф. 1", "description": "..."}
// Для дополнения: {"kind": "addition", "objectSlug": "birzha", "source": "...", "description": "..."}.
func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var in createRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := middleware.PrincipalFrom(r.Context())
	sub, err := h.svc.Submit(r.Context(), p.UserID, in.toDomain())
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toSubmissionResponse(sub))
	return nil
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	p, _ := middleware.PrincipalFrom(r.Context())
	sub, err := h.svc.Get(r.Context(), id, submission.Viewer{UserID: p.UserID, IsModerator: p.Role.AtLeast(auth.RoleModerator)})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toSubmissionResponse(sub))
	return nil
}

// uploadMedia: POST /api/submissions/{id}/media
// Форма multipart/form-data: поля kind, caption, yearTaken, sourceNote (необязательные, кроме kind) и file.
// Вместо sourceNote принимается и прежнее имя поля source.
func (h *Handler) uploadMedia(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes+mb) // +1 МБ на служебные поля формы
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

	sourceNote := r.FormValue("sourceNote")
	if sourceNote == "" {
		sourceNote = r.FormValue("source")
	}
	p, _ := middleware.PrincipalFrom(r.Context())
	f, err := h.svc.AddMedia(r.Context(), id, p.UserID, submission.MediaInput{
		Kind:         r.FormValue("kind"),
		Caption:      r.FormValue("caption"),
		YearTaken:    r.FormValue("yearTaken"),
		SourceNote:   sourceNote,
		OriginalName: header.Filename,
		Size:         header.Size,
		Body:         file,
	})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, toFileResponse(f))
	return nil
}

func (h *Handler) listMine(w http.ResponseWriter, r *http.Request) error {
	p, _ := middleware.PrincipalFrom(r.Context())
	list, err := h.svc.ListMine(r.Context(), p.UserID)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toListResponse(list))
	return nil
}

// queue: GET /api/moderation/submissions?status=pending&limit=50&offset=0
// Без status показываются ожидающие решения заявки, старые первыми.
func (h *Handler) queue(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	status := submission.Status(q.Get("status"))
	if status == "" {
		status = submission.StatusPending
	}
	limit, offset := 50, 0
	if v, err := strconv.Atoi(q.Get("limit")); err == nil && v > 0 {
		limit = v
	}
	if v, err := strconv.Atoi(q.Get("offset")); err == nil && v >= 0 {
		offset = v
	}
	list, err := h.svc.Queue(r.Context(), status, limit, offset)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toListResponse(list))
	return nil
}

// approve: POST /api/moderation/submissions/{id}/approve
// Для новой заявки тело — поля объекта каталога: {"slug": "dom-ivanova", "category": "manor",
// "protectionStatus": "identified", ...}; незаполненные берутся из заявки. Для дополнения тело не нужно.
func (h *Handler) approve(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	var in approveRequest
	if r.ContentLength != 0 { // пустое тело допустимо — для дополнения
		if err := httpx.DecodeJSON(r, &in); err != nil {
			return err
		}
	}
	p, _ := middleware.PrincipalFrom(r.Context())
	sub, err := h.svc.Approve(r.Context(), id, moderator(p), in.toDomain())
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toSubmissionResponse(sub))
	return nil
}

// reject: POST /api/moderation/submissions/{id}/reject {"reason": "...", "blockAuthor": false}
func (h *Handler) reject(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	var in rejectRequest
	if err := httpx.DecodeJSON(r, &in); err != nil {
		return err
	}
	p, _ := middleware.PrincipalFrom(r.Context())
	sub, err := h.svc.Reject(r.Context(), id, moderator(p), submission.RejectInput{Reason: in.Reason, BlockAuthor: in.BlockAuthor})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusOK, toSubmissionResponse(sub))
	return nil
}
