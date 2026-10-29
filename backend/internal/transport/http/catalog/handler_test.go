package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/catalog"
	"heritage-backend/internal/domain/media"
	mediahttp "heritage-backend/internal/transport/http/media"
	"heritage-backend/internal/transport/http/middleware"
)

// fakeCatalog — каталог из двух объектов; правило видимости — настоящее, из домена.
type fakeCatalog struct {
	catalog.Service // методы, которые тест не вызывает, не реализованы
	objects         map[string]catalog.Object
	lastFilter      catalog.Filter
	lastViewer      catalog.Viewer
}

func (c *fakeCatalog) GetVisible(_ context.Context, slug string, v catalog.Viewer) (catalog.Object, error) {
	o, ok := c.objects[slug]
	if !ok || !o.VisibleTo(v) {
		return catalog.Object{}, catalog.ErrObjectNotFound
	}
	return o, nil
}

func (c *fakeCatalog) List(_ context.Context, f catalog.Filter, v catalog.Viewer) ([]catalog.Object, error) {
	c.lastFilter, c.lastViewer = f, v
	return nil, nil
}

type fakeMedia struct {
	media.Service
	uploaded *media.UploadInput
	body     string
}

func (m *fakeMedia) ListByObject(context.Context, int64) ([]media.File, error) { return nil, nil }
func (m *fakeMedia) MaxUploadBytes() int64                                     { return 10 * media.MB }

func (m *fakeMedia) Upload(_ context.Context, in media.UploadInput) (media.File, error) {
	b, _ := io.ReadAll(in.Body)
	m.uploaded, m.body = &in, string(b)
	p := in.Projection
	return media.File{ID: 5, ObjectID: &in.ObjectID, Kind: in.Kind, Source: in.Source, Projection: &p, HAngle: in.HAngle}, nil
}

func newServer() (*fakeCatalog, *fakeMedia, http.Handler) {
	cat := &fakeCatalog{objects: map[string]catalog.Object{
		"birzha":    {ID: 1, Slug: "birzha", IsPublished: true},
		"chernovik": {ID: 2, Slug: "chernovik", IsPublished: false},
	}}
	med := &fakeMedia{}
	r := chi.NewRouter()
	r.Route("/api", NewHandler(cat, med, mediahttp.NewHandler(med)).Routes)
	return cat, med, r
}

func serve(h http.Handler, req *http.Request, p *auth.Principal) *httptest.ResponseRecorder {
	if p != nil {
		req = req.WithContext(middleware.WithPrincipal(req.Context(), *p))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

var (
	userP = &auth.Principal{UserID: 7, Role: auth.RoleUser}
	modP  = &auth.Principal{UserID: 9, Role: auth.RoleModerator}
)

func TestCardVisibility(t *testing.T) {
	_, _, h := newServer()

	rec := serve(h, httptest.NewRequest(http.MethodGet, "/api/objects/birzha", nil), nil)
	var card struct {
		Object map[string]any `json:"object"`
		Media  []any          `json:"media"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&card); err != nil || rec.Code != http.StatusOK ||
		card.Object["slug"] != "birzha" || card.Media == nil {
		t.Fatalf("карточка: %d %+v %v", rec.Code, card, err)
	}

	for _, p := range []*auth.Principal{nil, userP} {
		if rec := serve(h, httptest.NewRequest(http.MethodGet, "/api/objects/chernovik", nil), p); rec.Code != http.StatusNotFound {
			t.Fatalf("неопубликованный объект для %+v: %d", p, rec.Code)
		}
	}
	if rec := serve(h, httptest.NewRequest(http.MethodGet, "/api/objects/chernovik", nil), modP); rec.Code != http.StatusOK {
		t.Fatalf("неопубликованный объект для модератора: %d", rec.Code)
	}
}

func TestListParsesFilter(t *testing.T) {
	cat, _, h := newServer()
	rec := serve(h, httptest.NewRequest(http.MethodGet, "/api/objects?all=true&lat=46.35&lng=48.03&bbox=1,2,3,4", nil), modP)
	if rec.Code != http.StatusOK || rec.Body.String() != "[]\n" {
		t.Fatalf("список: %d %q", rec.Code, rec.Body.String())
	}
	f := cat.lastFilter
	if !f.IncludeUnpublished || f.Near == nil || f.RadiusM != 1000 || f.BBox == nil || f.BBox[3] != 4 || !cat.lastViewer.Moderator {
		t.Fatalf("фильтр: %+v, viewer %+v", f, cat.lastViewer)
	}

	for _, bad := range []string{"lat=x&lng=1", "lat=1&lng=1&radius=0", "bbox=1,2,3"} {
		if rec := serve(h, httptest.NewRequest(http.MethodGet, "/api/objects?"+bad, nil), nil); rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: %d", bad, rec.Code)
		}
	}
}

func TestUploadPanorama(t *testing.T) {
	_, med, h := newServer()

	form := func() (*bytes.Buffer, string) {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		for k, v := range map[string]string{"kind": "panorama", "projection": "partial", "hAngle": "180",
			"source": "ai_reconstruction", "sourceNote": "Книга", "author": "Иванов", "license": "CC BY"} {
			_ = mw.WriteField(k, v)
		}
		fw, _ := mw.CreateFormFile("file", "pano.jpg")
		_, _ = fw.Write([]byte("содержимое"))
		_ = mw.Close()
		return &buf, mw.FormDataContentType()
	}

	body, ct := form()
	req := httptest.NewRequest(http.MethodPost, "/api/objects/1/media", body)
	req.Header.Set("Content-Type", ct)
	if rec := serve(h, req, userP); rec.Code != http.StatusForbidden {
		t.Fatalf("загрузка пользователем: %d", rec.Code)
	}

	body, ct = form()
	req = httptest.NewRequest(http.MethodPost, "/api/objects/1/media", body)
	req.Header.Set("Content-Type", ct)
	rec := serve(h, req, modP)
	if rec.Code != http.StatusCreated {
		t.Fatalf("загрузка модератором: %d %s", rec.Code, rec.Body.String())
	}
	in := med.uploaded
	if in.ObjectID != 1 || in.Kind != media.KindPanorama || in.Projection != media.ProjectionPartial || in.HAngle == nil ||
		*in.HAngle != 180 || in.Source != media.SourceAIReconstruction || in.SourceNote != "Книга" || in.Author != "Иванов" ||
		in.License != "CC BY" || in.UploadedBy != 9 || in.OriginalName != "pano.jpg" || med.body != "содержимое" {
		t.Fatalf("разбор формы: %+v", in)
	}
	var out map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&out)
	if out["source"] != "ai_reconstruction" || out["projection"] != "partial" || out["hAngle"] != float64(180) || out["url"] != "/api/media/5/file" {
		t.Fatalf("ответ: %v", out)
	}
}
