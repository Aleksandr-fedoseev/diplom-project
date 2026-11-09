package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/transport/http/middleware"
)

// fakeService — сервис пользователей без базы; методы, которые тест не вызывает, не реализованы.
type fakeService struct {
	auth.Service
}

func (fakeService) List(context.Context, int, int) ([]auth.User, error) {
	return []auth.User{{ID: 1, Email: "a@t", Role: auth.RoleUser}}, nil
}

func (fakeService) SetRole(_ context.Context, _ auth.Principal, id int64, role auth.Role) (auth.User, error) {
	return auth.User{ID: id, Role: role}, nil
}

func (fakeService) GetByID(_ context.Context, id int64) (auth.User, error) {
	return auth.User{ID: id, Name: "Иван", Role: auth.RoleUser}, nil
}

func (fakeService) UpdateName(_ context.Context, id int64, name string) (auth.User, error) {
	return auth.User{ID: id, Name: name, Role: auth.RoleUser}, nil
}

func (fakeService) OpenAvatar(_ context.Context, id int64) (io.ReadSeekCloser, string, error) {
	if id != 1 {
		return nil, "", auth.ErrNoAvatar
	}
	return nopCloser{strings.NewReader("png")}, "image/png", nil
}

type nopCloser struct{ io.ReadSeeker }

func (nopCloser) Close() error { return nil }

func (fakeService) SetBlocked(_ context.Context, _ auth.Principal, id int64, blocked bool) (auth.User, error) {
	return auth.User{ID: id, IsBlocked: blocked}, nil
}

// Кто какие адреса управления пользователями может вызвать. Управление пользователями —
// прецедент администратора; модератор блокирует автора только при отказе по заявке.
func TestAdminRoutesAccess(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/api/admin/users", NewHandler(fakeService{}).AdminRoutes)

	roles := map[string]*auth.Principal{
		"гость":         nil,
		"пользователь":  {UserID: 3, Role: auth.RoleUser},
		"модератор":     {UserID: 2, Role: auth.RoleModerator},
		"администратор": {UserID: 1, Role: auth.RoleAdmin},
	}
	cases := []struct {
		method, path, body string
		want               map[string]int
	}{
		{http.MethodGet, "/api/admin/users", "",
			map[string]int{"гость": 401, "пользователь": 403, "модератор": 403, "администратор": 200}},
		{http.MethodPut, "/api/admin/users/5/blocked", `{"blocked":true}`,
			map[string]int{"гость": 401, "пользователь": 403, "модератор": 403, "администратор": 200}},
		{http.MethodPut, "/api/admin/users/5/role", `{"role":"moderator"}`,
			map[string]int{"гость": 401, "пользователь": 403, "модератор": 403, "администратор": 200}},
	}
	for _, c := range cases {
		for name, p := range roles {
			req := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
			req.Header.Set("Content-Type", "application/json")
			if p != nil {
				req = req.WithContext(middleware.WithPrincipal(req.Context(), *p))
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != c.want[name] {
				t.Errorf("%s %s, %s: %d, ожидалось %d (%s)", c.method, c.path, name, rec.Code, c.want[name], rec.Body)
			}
		}
	}
}

func newRouter() http.Handler {
	h := NewHandler(fakeService{})
	r := chi.NewRouter()
	r.Route("/api/auth", h.Routes)
	r.Route("/api/users", h.UserRoutes)
	return r
}

func TestProfile(t *testing.T) {
	r := newRouter()
	user := &auth.Principal{UserID: 7, Role: auth.RoleUser}
	do := func(method, path, body string, p *auth.Principal) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if p != nil {
			req = req.WithContext(middleware.WithPrincipal(req.Context(), *p))
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}

	if rec := do(http.MethodPut, "/api/auth/me", `{"name":"Пётр"}`, nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("гость меняет имя: %d", rec.Code)
	}
	rec := do(http.MethodPut, "/api/auth/me", `{"name":"Пётр"}`, user)
	var got struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatarUrl"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil || rec.Code != 200 || got.ID != 7 || got.Name != "Пётр" {
		t.Fatalf("смена имени: %d %+v %v", rec.Code, got, err)
	}

	// Аватар открыт всем, даже гостю: его показывают рядом с именем.
	if rec := do(http.MethodGet, "/api/users/1/avatar", "", nil); rec.Code != 200 || rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("аватар: %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if rec := do(http.MethodGet, "/api/users/2/avatar", "", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("нет аватара: %d", rec.Code)
	}
}
