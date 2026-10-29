// Package media — HTTP-обработчики медиафайлов: скачивание (с Range), удаление и загрузка.
package media

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/media"
	"heritage-backend/internal/transport/http/httpx"
	"heritage-backend/internal/transport/http/middleware"
)

type Handler struct {
	svc media.Service
}

func NewHandler(svc media.Service) *Handler { return &Handler{svc: svc} }

// Routes подключается по адресу /api/media.
func (h *Handler) Routes(r chi.Router) {
	r.Method(http.MethodGet, "/{id}/file", httpx.Handler(h.download))
	r.With(middleware.RequireRole(auth.RoleModerator)).Method(http.MethodDelete, "/{id}", httpx.Handler(h.delete))
}

// Upload вешается на POST /api/objects/{id}/media (только модератор — права проверяет маршрутизатор).
// Форма multipart/form-data: kind, caption, yearTaken, sourceNote, source, projection, hAngle,
// author, license и file.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) error {
	objectID, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	in, cleanup, err := h.ParseUploadForm(w, r)
	if err != nil {
		return err
	}
	defer cleanup()

	in.ObjectID = objectID
	f, err := h.svc.Upload(r.Context(), in)
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, ToFileResponse(f))
	return nil
}

// ParseUploadForm читает форму загрузки в media.UploadInput (без ObjectID и SubmissionID —
// их заполняет вызывающий). Экспортирован для загрузки файлов к заявке.
// cleanup закрывает файл и удаляет временные файлы формы; вызывать после Upload.
func (h *Handler) ParseUploadForm(w http.ResponseWriter, r *http.Request) (media.UploadInput, func(), error) {
	r.Body = http.MaxBytesReader(w, r.Body, h.svc.MaxUploadBytes()+media.MB) // +1 МБ на служебные поля формы
	// В памяти держим до 8 МБ, остальное Go сам временно складывает на диск.
	if err := r.ParseMultipartForm(8 * media.MB); err != nil {
		return media.UploadInput{}, nil, httpx.BadRequest("не удалось прочитать форму: " + err.Error())
	}
	removeForm := func() { _ = r.MultipartForm.RemoveAll() }

	var hAngle *int
	if v := strings.TrimSpace(r.FormValue("hAngle")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			removeForm()
			return media.UploadInput{}, nil, httpx.BadRequest("hAngle должен быть целым числом градусов")
		}
		hAngle = &n
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		removeForm()
		return media.UploadInput{}, nil, httpx.BadRequest("нет поля file")
	}

	p, _ := middleware.PrincipalFrom(r.Context())
	in := media.UploadInput{
		Kind:         media.Kind(r.FormValue("kind")),
		Caption:      r.FormValue("caption"),
		YearTaken:    r.FormValue("yearTaken"),
		SourceNote:   r.FormValue("sourceNote"),
		Source:       media.Source(r.FormValue("source")),
		Projection:   media.Projection(r.FormValue("projection")),
		HAngle:       hAngle,
		Author:       r.FormValue("author"),
		License:      r.FormValue("license"),
		OriginalName: header.Filename,
		Size:         header.Size,
		Body:         file,
		UploadedBy:   p.UserID,
	}
	return in, func() { _ = file.Close(); removeForm() }, nil
}

// viewer — кто скачивает файл: гость, пользователь или модератор.
func viewer(r *http.Request) media.Viewer {
	p, ok := middleware.PrincipalFrom(r.Context())
	if !ok {
		return media.Viewer{}
	}
	return media.Viewer{UserID: p.UserID, Moderator: p.Role.AtLeast(auth.RoleModerator)}
}

// download: GET /api/media/{id}/file. Поддерживает Range: аудио и 3D-модели грузятся частями.
func (h *Handler) download(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.IDParam(r, "id")
	if err != nil {
		return err
	}
	f, body, err := h.svc.Open(r.Context(), id, viewer(r))
	if err != nil {
		return err
	}
	defer body.Close()

	w.Header().Set("Content-Type", f.ContentType)
	if f.IsPublic() {
		w.Header().Set("Cache-Control", "public, max-age=86400")
	} else {
		w.Header().Set("Cache-Control", "private, no-store") // файл заявки не кэшируют прокси
	}
	http.ServeContent(w, r, f.OriginalName, f.CreatedAt, body)
	return nil
}

// delete: DELETE /api/media/{id}
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
