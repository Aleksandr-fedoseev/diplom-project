package media

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/media"
	"heritage-backend/internal/transport/http/middleware"
)

// fakeService — сервис медиафайлов в памяти; правило видимости — настоящее, из домена.
type fakeService struct {
	files   map[int64]media.File
	content string
	deleted []int64
}

func (s *fakeService) Upload(context.Context, media.UploadInput) (media.File, error) {
	return media.File{}, nil
}
func (s *fakeService) ListByObject(context.Context, int64) ([]media.File, error)     { return nil, nil }
func (s *fakeService) ListBySubmission(context.Context, int64) ([]media.File, error) { return nil, nil }
func (s *fakeService) AttachToObject(context.Context, int64, int64) error            { return nil }
func (s *fakeService) MaxUploadBytes() int64                                         { return 10 * media.MB }

func (s *fakeService) Open(_ context.Context, id int64, v media.Viewer) (media.File, io.ReadSeekCloser, error) {
	f, ok := s.files[id]
	if !ok || !f.CanView(v) {
		return media.File{}, nil, media.ErrFileNotFound
	}
	return f, nopCloser{bytes.NewReader([]byte(s.content))}, nil
}

func (s *fakeService) Delete(_ context.Context, id int64) error {
	s.deleted = append(s.deleted, id)
	return nil
}

type nopCloser struct{ io.ReadSeeker }

func (nopCloser) Close() error { return nil }

func newServer() (*fakeService, http.Handler) {
	objectID, submissionID, author := int64(1), int64(2), int64(7)
	svc := &fakeService{content: "0123456789", files: map[int64]media.File{
		10: {ID: 10, ObjectID: &objectID, ContentType: "image/jpeg", OriginalName: "a.jpg", CreatedAt: time.Now()},
		20: {ID: 20, SubmissionID: &submissionID, UploadedBy: &author, ContentType: "image/jpeg", OriginalName: "b.jpg", CreatedAt: time.Now()},
	}}
	r := chi.NewRouter()
	r.Route("/api/media", NewHandler(svc).Routes)
	return svc, r
}

func do(h http.Handler, method, url string, p *auth.Principal, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, nil)
	if p != nil {
		req = req.WithContext(middleware.WithPrincipal(req.Context(), *p))
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestDownloadVisibility(t *testing.T) {
	_, h := newServer()
	authorP := &auth.Principal{UserID: 7, Role: auth.RoleUser}
	otherP := &auth.Principal{UserID: 8, Role: auth.RoleUser}
	modP := &auth.Principal{UserID: 9, Role: auth.RoleModerator}

	// Файл объекта виден гостю и кэшируется публично.
	rec := do(h, http.MethodGet, "/api/media/10/file", nil, nil)
	if rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") != "public, max-age=86400" || rec.Body.String() != "0123456789" {
		t.Fatalf("файл объекта: %d %q %q", rec.Code, rec.Header().Get("Cache-Control"), rec.Body.String())
	}

	// Файл заявки для гостя и чужого пользователя «не существует».
	for _, p := range []*auth.Principal{nil, otherP} {
		rec = do(h, http.MethodGet, "/api/media/20/file", p, nil)
		var body map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&body)
		if rec.Code != http.StatusNotFound || body["error"] != "файл не найден" {
			t.Fatalf("файл заявки для %+v: %d %v", p, rec.Code, body)
		}
	}

	// Автору и модератору — виден, но не кэшируется.
	for _, p := range []*auth.Principal{authorP, modP} {
		rec = do(h, http.MethodGet, "/api/media/20/file", p, nil)
		if rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") != "private, no-store" {
			t.Fatalf("файл заявки для %+v: %d %q", p, rec.Code, rec.Header().Get("Cache-Control"))
		}
	}

	// Range: отдаётся только запрошенная часть.
	rec = do(h, http.MethodGet, "/api/media/10/file", nil, map[string]string{"Range": "bytes=2-4"})
	if rec.Code != http.StatusPartialContent || rec.Body.String() != "234" {
		t.Fatalf("Range: %d %q", rec.Code, rec.Body.String())
	}

	if rec = do(h, http.MethodGet, "/api/media/abc/file", nil, nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("некорректный id: %d", rec.Code)
	}
}

func TestDeleteRequiresModerator(t *testing.T) {
	svc, h := newServer()
	if rec := do(h, http.MethodDelete, "/api/media/10", nil, nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("гость: %d", rec.Code)
	}
	if rec := do(h, http.MethodDelete, "/api/media/10", &auth.Principal{UserID: 7, Role: auth.RoleUser}, nil); rec.Code != http.StatusForbidden {
		t.Fatalf("пользователь: %d", rec.Code)
	}
	if rec := do(h, http.MethodDelete, "/api/media/10", &auth.Principal{UserID: 9, Role: auth.RoleModerator}, nil); rec.Code != http.StatusNoContent || len(svc.deleted) != 1 {
		t.Fatalf("модератор: %d", rec.Code)
	}
}
