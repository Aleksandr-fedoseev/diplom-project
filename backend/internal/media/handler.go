package media

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

// Routes подключается по адресу /api/media.
func (h *Handler) Routes(r chi.Router) {
	r.Method(http.MethodGet, "/{id}/file", httpx.Handler(h.download))
	r.With(auth.RequireRole(auth.RoleModerator)).Method(http.MethodDelete, "/{id}", httpx.Handler(h.delete))
}

// Upload вешается на POST /api/objects/{id}/media.
// Форма multipart/form-data: поля kind, caption (необязательно) и file.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) error {
	objectID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return httpx.BadRequest("некорректный id объекта")
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
	f, err := h.svc.Upload(r.Context(), UploadInput{
		ObjectID:     objectID,
		Kind:         r.FormValue("kind"),
		Caption:      r.FormValue("caption"),
		YearTaken:    r.FormValue("yearTaken"),
		Source:       r.FormValue("source"),
		OriginalName: header.Filename,
		Size:         header.Size,
		Body:         file,
		UploadedBy:   p.UserID,
	})
	if err != nil {
		return err
	}
	httpx.JSON(w, http.StatusCreated, f)
	return nil
}

func (h *Handler) download(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return httpx.BadRequest("некорректный id")
	}
	f, obj, err := h.svc.Open(r.Context(), id, func(f File) bool {
		p, ok := auth.FromContext(r.Context())
		return ok && f.CanView(p.UserID, auth.AtLeast(p.Role, auth.RoleModerator))
	})
	if err != nil {
		return err
	}
	defer obj.Close()

	w.Header().Set("Content-Type", f.ContentType)
	if f.IsPublic() {
		w.Header().Set("Cache-Control", "public, max-age=86400")
	} else {
		w.Header().Set("Cache-Control", "private, no-store") // файл заявки не кэшируют прокси
	}
	http.ServeContent(w, r, f.OriginalName, f.CreatedAt, obj)
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
