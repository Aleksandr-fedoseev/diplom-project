package route

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/route"
	"heritage-backend/internal/transport/http/middleware"
)

// Тесты обработчиков с фейковым сервисом: разбор запроса, права, формат JSON и перевод ошибок.

type fakeService struct {
	viewer      *route.Actor
	deleteID    int64
	deleteActor route.Actor
	saveOwner   *int64
	buildErr    error
}

func (f *fakeService) Build(context.Context, []string) (route.Route, error) {
	if f.buildErr != nil {
		return route.Route{}, f.buildErr
	}
	return route.Route{DistanceM: 10, DurationS: 5, Stops: []route.Stop{{Position: 1, Slug: "birzha"}}}, nil
}

func (f *fakeService) Save(_ context.Context, in route.SaveInput, ownerID *int64) (route.Route, error) {
	f.saveOwner = ownerID
	return route.Route{ID: 1, Slug: in.Slug, Title: in.Title}, nil
}

func (f *fakeService) ListThematic(context.Context) ([]route.Route, error) { return nil, nil }
func (f *fakeService) ListByOwner(context.Context, int64) ([]route.Route, error) {
	return nil, nil
}

func (f *fakeService) Get(_ context.Context, slug string, viewer *route.Actor) (route.Route, error) {
	f.viewer = viewer
	if slug == "missing" {
		return route.Route{}, route.ErrNotFound
	}
	return route.Route{ID: 1, Slug: slug}, nil
}

func (f *fakeService) Delete(_ context.Context, id int64, actor route.Actor) error {
	f.deleteID, f.deleteActor = id, actor
	return nil
}

func serve(svc route.Service, method, path, body string, p *auth.Principal) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	NewHandler(svc).Routes(r)
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if p != nil {
		req = req.WithContext(middleware.WithPrincipal(req.Context(), *p))
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestHandler(t *testing.T) {
	user := &auth.Principal{UserID: 7, Role: auth.RoleUser}
	admin := &auth.Principal{UserID: 1, Role: auth.RoleAdmin}

	cases := []struct {
		name, method, path, body string
		who                      *auth.Principal
		status                   int
		bodyHas                  string
	}{
		{"список пустой — []", "GET", "/routes", "", nil, 200, "[]"},
		{"построение без сохранения", "POST", "/routes/build", `{"objects":["a","b"]}`, nil, 200, `"stops":[{"position":1`},
		{"неизвестное поле", "POST", "/routes/build", `{"slugs":["a"]}`, nil, 400, "некорректный JSON"},
		{"чужой маршрут — 404", "GET", "/routes/missing", "", nil, 404, "маршрут не найден"},
		{"тематический — только админ", "POST", "/routes", `{"slug":"x","title":"X","objects":[]}`, user, 403, ""},
		{"тематический админом", "POST", "/routes", `{"slug":"x","title":"X","objects":[]}`, admin, 201, `"slug":"x"`},
		{"мои маршруты — только с входом", "GET", "/me/routes", "", nil, 401, ""},
		{"удаление гостем", "DELETE", "/routes/5", "", nil, 401, ""},
		{"удаление: не число", "DELETE", "/routes/abc", "", user, 400, ""},
		{"удаление", "DELETE", "/routes/5", "", user, 204, ""},
	}
	for _, c := range cases {
		rec := serve(&fakeService{}, c.method, c.path, c.body, c.who)
		if rec.Code != c.status || !strings.Contains(rec.Body.String(), c.bodyHas) {
			t.Errorf("%s: %d %s", c.name, rec.Code, rec.Body.String())
		}
	}
}

func TestHandlerPassesActor(t *testing.T) {
	svc := &fakeService{}
	serve(svc, "GET", "/routes/old-town", "", nil)
	if svc.viewer != nil {
		t.Fatal("гость должен передаваться как nil")
	}
	serve(svc, "GET", "/routes/old-town", "", &auth.Principal{UserID: 1, Role: auth.RoleAdmin})
	if svc.viewer == nil || !svc.viewer.IsAdmin {
		t.Fatalf("администратор: %+v", svc.viewer)
	}

	serve(svc, "DELETE", "/routes/5", "", &auth.Principal{UserID: 7, Role: auth.RoleModerator})
	if svc.deleteID != 5 || svc.deleteActor != (route.Actor{UserID: 7}) {
		t.Fatalf("удаление: id %d, %+v", svc.deleteID, svc.deleteActor)
	}

	serve(svc, "POST", "/me/routes", `{"title":"Мой","objects":["a","b"]}`, &auth.Principal{UserID: 7, Role: auth.RoleUser})
	if svc.saveOwner == nil || *svc.saveOwner != 7 {
		t.Fatalf("владелец: %v", svc.saveOwner)
	}
}

func TestHandlerRouterUnavailable(t *testing.T) {
	rec := serve(&fakeService{buildErr: route.ErrRouterUnavailable}, "POST", "/routes/build", `{"objects":["a","b"]}`, nil)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("маршрутизатор недоступен: %d", rec.Code)
	}
}
