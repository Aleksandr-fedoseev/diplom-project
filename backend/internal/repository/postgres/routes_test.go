package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/domain/route"
	"heritage-backend/internal/infrastructure/postgres"
	repo "heritage-backend/internal/repository/postgres"
	"heritage-backend/internal/testdb"
)

// Тест репозитория маршрутов на настоящей БД: сохранение в транзакции, линия через PostGIS,
// порядок точек, списки и удаление.
func TestRoutesRepository(t *testing.T) {
	pool := testdb.Connect(t)
	routes := repo.NewRoutes(pool)
	tm := postgres.NewTxManager(pool)
	ctx := context.Background()

	// Три объекта из стартовых данных каталога — в нужном нам порядке.
	var objectIDs []int64
	for _, slug := range []string{"dom-agamzhanovoy", "birzha", "azovo-donskoy-bank"} {
		var id int64
		if err := pool.QueryRow(ctx, "SELECT id FROM catalog.heritage_objects WHERE slug = $1", slug).Scan(&id); err != nil {
			t.Fatalf("объект %s: %v", slug, err)
		}
		objectIDs = append(objectIDs, id)
	}
	stops := make([]route.Stop, len(objectIDs))
	for i, id := range objectIDs {
		stops[i] = route.Stop{Position: i + 1, ObjectID: id}
	}
	geometry := json.RawMessage(`{"type":"LineString","coordinates":[[48.0285,46.3538],[48.0301,46.3532],[48.0308,46.3529]]}`)

	// Маршрут и точки — в одной транзакции.
	slug := testdb.Unique("repo-route")
	var id int64
	err := tm.Do(ctx, func(ctx context.Context) error {
		var err error
		if id, err = routes.Create(ctx, route.Route{Slug: slug, Title: "Тест", DistanceM: 420.5, DurationS: 300, Geometry: geometry}); err != nil {
			return err
		}
		return routes.AddStops(ctx, id, stops)
	})
	if err != nil || id == 0 {
		t.Fatalf("сохранение: %d, %v", id, err)
	}

	got, err := routes.BySlug(ctx, slug)
	if err != nil || got.ID != id || got.OwnerID != nil || got.DistanceM != 420.5 || got.CreatedAt.IsZero() {
		t.Fatalf("по slug: %+v, %v", got, err)
	}
	var line struct {
		Type        string      `json:"type"`
		Coordinates [][]float64 `json:"coordinates"`
	}
	if err := json.Unmarshal(got.Geometry, &line); err != nil || line.Type != "LineString" || len(line.Coordinates) != 3 {
		t.Fatalf("линия из БД: %s, %v", got.Geometry, err)
	}

	if _, err := routes.Create(ctx, route.Route{Slug: slug, Title: "Дубль", Geometry: geometry}); !errors.Is(err, route.ErrSlugTaken) {
		t.Fatalf("дубликат slug: %v", err)
	}
	if _, err := routes.BySlug(ctx, "no-such-route"); !errors.Is(err, route.ErrNotFound) {
		t.Fatalf("несуществующий: %v", err)
	}

	// Сбой в транзакции — ни маршрута, ни точек.
	failedSlug := testdb.Unique("repo-route-failed")
	err = tm.Do(ctx, func(ctx context.Context) error {
		rid, err := routes.Create(ctx, route.Route{Slug: failedSlug, Title: "Сбой", Geometry: geometry})
		if err != nil {
			return err
		}
		return routes.AddStops(ctx, rid, []route.Stop{{Position: 1, ObjectID: -1}}) // нет такого объекта
	})
	if err == nil {
		t.Fatal("ожидалась ошибка внешнего ключа")
	}
	if _, err := routes.BySlug(ctx, failedSlug); !errors.Is(err, route.ErrNotFound) {
		t.Fatalf("маршрут после отката: %v", err)
	}

	// Маршрут пользователя.
	users := repo.NewUsers(pool)
	owner, err := users.Create(ctx, auth.User{Email: testdb.Unique("route-owner") + "@example.ru", Name: "Иван", Role: auth.RoleUser, PasswordHash: "h"})
	if err != nil {
		t.Fatal(err)
	}
	mineID, err := routes.Create(ctx, route.Route{Slug: testdb.Unique("u-route"), Title: "Мой", OwnerID: &owner.ID, Geometry: geometry})
	if err != nil {
		t.Fatal(err)
	}
	if err := routes.AddStops(ctx, mineID, stops[:2]); err != nil {
		t.Fatal(err)
	}

	// Точки двух маршрутов одним запросом — по маршруту и номеру остановки.
	refs, err := routes.StopRefs(ctx, []int64{mineID, id})
	if err != nil || len(refs) != 5 {
		t.Fatalf("точки: %+v, %v", refs, err)
	}
	first, second := id, mineID
	if mineID < id {
		first, second = mineID, id
	}
	for i, ref := range refs {
		wantRoute := first
		if (first == id && i >= 3) || (first == mineID && i >= 2) {
			wantRoute = second
		}
		if ref.RouteID != wantRoute {
			t.Fatalf("точка %d из маршрута %d, ожидался %d", i, ref.RouteID, wantRoute)
		}
	}
	for _, ref := range refs {
		if ref.RouteID == id && ref.ObjectID != objectIDs[ref.Position-1] {
			t.Fatalf("порядок остановок: %+v", refs)
		}
	}

	thematic, err := routes.ListThematic(ctx)
	if err != nil || !containsRoute(thematic, id) || containsRoute(thematic, mineID) {
		t.Fatalf("тематические: %v", err)
	}
	for _, r := range thematic {
		if r.Geometry != nil {
			t.Fatal("в списке не должно быть линий")
		}
	}
	mine, err := routes.ListByOwner(ctx, owner.ID)
	if err != nil || len(mine) != 1 || mine[0].ID != mineID || *mine[0].OwnerID != owner.ID {
		t.Fatalf("маршруты пользователя: %+v, %v", mine, err)
	}

	if o, err := routes.Owner(ctx, mineID); err != nil || o == nil || *o != owner.ID {
		t.Fatalf("владелец: %v, %v", o, err)
	}
	if o, err := routes.Owner(ctx, id); err != nil || o != nil {
		t.Fatalf("владелец тематического: %v, %v", o, err)
	}
	if _, err := routes.Owner(ctx, -1); !errors.Is(err, route.ErrNotFound) {
		t.Fatalf("владелец несуществующего: %v", err)
	}

	// Удаление убирает и точки (ON DELETE CASCADE).
	for _, rid := range []int64{id, mineID} {
		if err := routes.Delete(ctx, rid); err != nil {
			t.Fatal(err)
		}
	}
	if refs, err := routes.StopRefs(ctx, []int64{id, mineID}); err != nil || len(refs) != 0 {
		t.Fatalf("точки после удаления: %+v, %v", refs, err)
	}
	if _, err := routes.BySlug(ctx, slug); !errors.Is(err, route.ErrNotFound) {
		t.Fatalf("после удаления: %v", err)
	}
}

func containsRoute(list []route.Route, id int64) bool {
	for _, r := range list {
		if r.ID == id {
			return true
		}
	}
	return false
}
