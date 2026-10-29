package routes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"heritage-backend/internal/auth"
	"heritage-backend/internal/catalog"
	"heritage-backend/internal/httpx"
	"heritage-backend/internal/osrm"
	"heritage-backend/internal/testdb"
)

// fakeOSRM отвечает как OSRM, но без расчёта: тесту важна работа с БД, а не маршрутизация.
func fakeOSRM(t *testing.T) *osrm.Client {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":"Ok","routes":[{"distance":420.5,"duration":300,
			"geometry":{"type":"LineString","coordinates":[[48.0285,46.3538],[48.0301,46.3532],[48.0308,46.3529]]}}]}`))
	}))
	t.Cleanup(srv.Close)
	return osrm.NewClient(srv.URL)
}

func wantStatus(t *testing.T, err error, status int) {
	t.Helper()
	var he *httpx.Error
	if !errors.As(err, &he) || he.Status != status {
		t.Fatalf("ожидалась ошибка %d, получено: %v", status, err)
	}
}

func TestSaveGetListDelete(t *testing.T) {
	db := testdb.Connect(t)
	svc := NewService(db, catalog.NewService(db), fakeOSRM(t))
	ctx := context.Background()
	order := []string{"dom-agamzhanovoy", "birzha", "azovo-donskoy-bank"}

	_, err := svc.Build(ctx, []string{"birzha", "no-such-object"})
	wantStatus(t, err, 400)

	slug := testdb.Unique("test-route")
	saved, err := svc.Save(ctx, SaveInput{Slug: slug, Title: "Тестовый маршрут", Objects: order}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if saved.DistanceM != 420.5 || len(saved.Geometry) == 0 || len(saved.Stops) != 3 {
		t.Fatalf("сохранённый маршрут: %+v", saved)
	}
	for i, st := range saved.Stops {
		if st.Slug != order[i] || st.Position != i+1 || st.Lat == 0 {
			t.Fatalf("остановка %d: %+v, ожидался %s", i+1, st, order[i])
		}
	}

	_, err = svc.Save(ctx, SaveInput{Slug: slug, Title: "Дубликат", Objects: order}, nil)
	wantStatus(t, err, 409)

	thematic, err := svc.ListThematic(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var inList *Route
	for i := range thematic {
		if thematic[i].Slug == slug {
			inList = &thematic[i]
		}
	}
	if inList == nil || len(inList.Stops) != 3 || inList.Stops[0].Slug != order[0] || inList.Geometry != nil {
		t.Fatalf("маршрут в списке тематических: %+v", inList)
	}

	// Пользовательский маршрут: владелец видит его в своём списке и может удалить, чужой — нет.
	users := auth.NewService(db)
	owner, err := users.Register(ctx, auth.RegisterInput{Email: testdb.Unique("owner") + "@example.ru", Password: "password1", Name: "Владелец"})
	if err != nil {
		t.Fatal(err)
	}
	mine, err := svc.Save(ctx, SaveInput{Title: "Мой маршрут", Objects: order[:2]}, &owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	list, err := svc.ListByOwner(ctx, owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != mine.ID || len(list[0].Stops) != 2 {
		t.Fatalf("мои маршруты: %+v", list)
	}

	wantStatus(t, svc.Delete(ctx, mine.ID, owner.ID+1000, false), 403)
	if err := svc.Delete(ctx, mine.ID, owner.ID, false); err != nil {
		t.Fatal(err)
	}
	wantStatus(t, svc.Delete(ctx, saved.ID, owner.ID, false), 403)
	if err := svc.Delete(ctx, saved.ID, 0, true); err != nil {
		t.Fatal(err)
	}
	_, err = svc.Get(ctx, slug)
	wantStatus(t, err, 404)
}
