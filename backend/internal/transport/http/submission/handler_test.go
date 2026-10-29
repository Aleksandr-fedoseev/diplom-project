package submission

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/submission"
	"heritage-backend/internal/transport/http/middleware"
)

// fakeService запоминает аргументы вызовов и возвращает заранее заданную ошибку.
type fakeService struct {
	err       error
	viewer    submission.Viewer
	mod       submission.Moderator
	approve   submission.ApproveInput
	media     submission.MediaInput
	body      string
	newSub    submission.NewSubmission
	queueArgs [3]any
}

var sample = submission.Submission{
	ID: 7, AuthorID: 1, Kind: submission.KindNewObject, Title: "Дом", Source: "Архив",
	Status: submission.StatusPending, Coordinates: &submission.Coordinates{Lat: 46, Lng: 48},
}

func (f *fakeService) Submit(_ context.Context, _ int64, in submission.NewSubmission) (submission.Submission, error) {
	f.newSub = in
	return sample, f.err
}

func (f *fakeService) AddMedia(_ context.Context, _, _ int64, in submission.MediaInput) (submission.File, error) {
	f.media = in
	b, _ := io.ReadAll(in.Body)
	f.body = string(b)
	return submission.File{ID: 1, SourceNote: in.SourceNote}, f.err
}

func (f *fakeService) Get(_ context.Context, _ int64, v submission.Viewer) (submission.Submission, error) {
	f.viewer = v
	return sample, f.err
}

func (f *fakeService) ListMine(context.Context, int64) ([]submission.Submission, error) {
	return nil, f.err
}

func (f *fakeService) Queue(_ context.Context, status submission.Status, limit, offset int) ([]submission.Submission, error) {
	f.queueArgs = [3]any{status, limit, offset}
	return []submission.Submission{sample}, f.err
}

func (f *fakeService) Approve(_ context.Context, _ int64, mod submission.Moderator, in submission.ApproveInput) (submission.Submission, error) {
	f.mod, f.approve = mod, in
	return sample, f.err
}

func (f *fakeService) Reject(_ context.Context, _ int64, mod submission.Moderator, _ submission.RejectInput) (submission.Submission, error) {
	f.mod = mod
	return sample, f.err
}

// do выполняет запрос от имени пользователя p (nil — гость).
func do(t *testing.T, svc *fakeService, p *auth.Principal, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	r := chi.NewRouter()
	NewHandler(svc, 10<<20).Routes(r)
	if p != nil {
		req = req.WithContext(middleware.WithPrincipal(req.Context(), *p))
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

var (
	user = &auth.Principal{UserID: 1, Role: auth.RoleUser}
	mod  = &auth.Principal{UserID: 10, Role: auth.RoleModerator}
)

func TestRightsAndErrors(t *testing.T) {
	svc := &fakeService{}
	if w := do(t, svc, nil, httptest.NewRequest(http.MethodGet, "/me/submissions", nil)); w.Code != http.StatusUnauthorized {
		t.Fatalf("гость: %d", w.Code)
	}
	if w := do(t, svc, user, httptest.NewRequest(http.MethodGet, "/moderation/submissions", nil)); w.Code != http.StatusForbidden {
		t.Fatalf("пользователь в очереди модерации: %d", w.Code)
	}
	if w := do(t, svc, user, httptest.NewRequest(http.MethodGet, "/me/submissions", nil)); w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("пустой список: %d %s", w.Code, w.Body)
	}

	for err, code := range map[error]int{
		submission.ErrNotFound:       http.StatusNotFound,
		submission.ErrAlreadyDecided: http.StatusConflict,
		submission.ErrOwnSubmission:  http.StatusForbidden,
		submission.ErrReasonRequired: http.StatusBadRequest,
	} {
		svc.err = err
		w := do(t, svc, mod, httptest.NewRequest(http.MethodPost, "/moderation/submissions/7/reject", strings.NewReader(`{"reason":"x"}`)))
		if w.Code != code || !strings.Contains(w.Body.String(), err.Error()) {
			t.Fatalf("%v: %d %s", err, w.Code, w.Body)
		}
	}
}

func TestCreateAndGetJSON(t *testing.T) {
	svc := &fakeService{}
	body := `{"kind":"addition","objectSlug":"birzha","source":"Книга","coordinates":{"lat":46,"lng":48}}`
	w := do(t, svc, user, httptest.NewRequest(http.MethodPost, "/submissions", strings.NewReader(body)))
	if w.Code != http.StatusCreated || svc.newSub.Kind != submission.KindAddition || svc.newSub.Coordinates == nil {
		t.Fatalf("создание: %d %s %+v", w.Code, w.Body, svc.newSub)
	}
	var got map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if got["id"] != float64(7) || got["status"] != "pending" || got["coordinates"] == nil || got["author"] != nil {
		t.Fatalf("JSON заявки: %v", got)
	}
	// Неизвестное поле отклоняется.
	if w := do(t, svc, user, httptest.NewRequest(http.MethodPost, "/submissions", strings.NewReader(`{"titel":"x"}`))); w.Code != http.StatusBadRequest {
		t.Fatalf("опечатка в поле: %d", w.Code)
	}

	do(t, svc, mod, httptest.NewRequest(http.MethodGet, "/submissions/7", nil))
	if !svc.viewer.IsModerator || svc.viewer.UserID != 10 {
		t.Fatalf("модератор как зритель: %+v", svc.viewer)
	}
	do(t, svc, user, httptest.NewRequest(http.MethodGet, "/submissions/7", nil))
	if svc.viewer.IsModerator {
		t.Fatal("пользователь не модератор")
	}
	if w := do(t, svc, user, httptest.NewRequest(http.MethodGet, "/submissions/abc", nil)); w.Code != http.StatusBadRequest {
		t.Fatalf("неверный id: %d", w.Code)
	}
}

func TestQueueAndApprove(t *testing.T) {
	svc := &fakeService{}
	do(t, svc, mod, httptest.NewRequest(http.MethodGet, "/moderation/submissions?limit=5&offset=10", nil))
	if svc.queueArgs != [3]any{submission.StatusPending, 5, 10} {
		t.Fatalf("параметры очереди: %v", svc.queueArgs)
	}

	// Для дополнения тело не нужно.
	if w := do(t, svc, mod, httptest.NewRequest(http.MethodPost, "/moderation/submissions/7/approve", nil)); w.Code != http.StatusOK {
		t.Fatalf("одобрение без тела: %d %s", w.Code, w.Body)
	}
	if svc.mod != (submission.Moderator{UserID: 10, Role: "moderator"}) {
		t.Fatalf("модератор: %+v", svc.mod)
	}
	body := `{"slug":"dom","category":"manor","protectionStatus":"identified","yearBuilt":"1890","coordinates":{"lat":1,"lng":2}}`
	do(t, svc, mod, httptest.NewRequest(http.MethodPost, "/moderation/submissions/7/approve", strings.NewReader(body)))
	if svc.approve.Slug != "dom" || svc.approve.YearBuilt == nil || svc.approve.Coordinates == nil || svc.approve.Coordinates.Lng != 2 {
		t.Fatalf("поля объекта: %+v", svc.approve)
	}
}

func multipartRequest(t *testing.T, fields map[string]string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = mw.WriteField(k, v)
	}
	fw, _ := mw.CreateFormFile("file", "photo.jpg")
	_, _ = fw.Write([]byte("jpeg"))
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/submissions/7/media", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestUploadMedia(t *testing.T) {
	svc := &fakeService{}
	w := do(t, svc, user, multipartRequest(t, map[string]string{"kind": "photo", "sourceNote": "Архив", "yearTaken": "1950"}))
	if w.Code != http.StatusCreated || svc.media.SourceNote != "Архив" || svc.media.OriginalName != "photo.jpg" ||
		svc.body != "jpeg" || !strings.Contains(w.Body.String(), `"sourceNote":"Архив"`) {
		t.Fatalf("загрузка: %d %s %+v", w.Code, w.Body, svc.media)
	}
	// Прежнее имя поля source тоже принимается.
	do(t, svc, user, multipartRequest(t, map[string]string{"kind": "photo", "source": "Книга"}))
	if svc.media.SourceNote != "Книга" {
		t.Fatalf("поле source: %+v", svc.media)
	}
}
