package route

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"time"

	"heritage-backend/internal/domain/errs"
	"heritage-backend/internal/domain/route"
)

// Тесты сервиса без БД и HTTP: репозиторий, каталог, маршрутизатор и транзакции — фейки в памяти.

type txKey struct{}

func inTx(ctx context.Context) bool { return ctx.Value(txKey{}) != nil }

// fakeRepo хранит маршруты и точки в памяти и запоминает, сколько раз читали точки.
type fakeRepo struct {
	routes        map[int64]route.Route
	refs          []route.StopRef
	nextID        int64
	stopRefsCalls int
	failAddStops  bool
	writesOutside int // записи вне транзакции — их быть не должно
}

func newFakeRepo() *fakeRepo { return &fakeRepo{routes: map[int64]route.Route{}} }

func (f *fakeRepo) Create(ctx context.Context, r route.Route) (int64, error) {
	if !inTx(ctx) {
		f.writesOutside++
	}
	for _, x := range f.routes {
		if x.Slug == r.Slug {
			return 0, route.ErrSlugTaken
		}
	}
	f.nextID++
	r.ID, r.Stops, r.CreatedAt = f.nextID, nil, time.Now()
	f.routes[r.ID] = r
	return r.ID, nil
}

func (f *fakeRepo) AddStops(ctx context.Context, routeID int64, stops []route.Stop) error {
	if !inTx(ctx) {
		f.writesOutside++
	}
	if f.failAddStops {
		return errors.New("сбой БД")
	}
	for _, st := range stops {
		f.refs = append(f.refs, route.StopRef{RouteID: routeID, Position: st.Position, ObjectID: st.ObjectID})
	}
	return nil
}

func (f *fakeRepo) BySlug(_ context.Context, slug string) (route.Route, error) {
	for _, r := range f.routes {
		if r.Slug == slug {
			return r, nil
		}
	}
	return route.Route{}, route.ErrNotFound
}

func (f *fakeRepo) list(keep func(route.Route) bool) []route.Route {
	var out []route.Route
	for _, r := range f.routes {
		if keep(r) {
			r.Geometry = nil
			out = append(out, r)
		}
	}
	slices.SortFunc(out, func(a, b route.Route) int { return int(a.ID - b.ID) })
	return out
}

func (f *fakeRepo) ListThematic(context.Context) ([]route.Route, error) {
	return f.list(func(r route.Route) bool { return r.OwnerID == nil }), nil
}

func (f *fakeRepo) ListByOwner(_ context.Context, ownerID int64) ([]route.Route, error) {
	return f.list(func(r route.Route) bool { return r.OwnerID != nil && *r.OwnerID == ownerID }), nil
}

func (f *fakeRepo) StopRefs(_ context.Context, routeIDs []int64) ([]route.StopRef, error) {
	f.stopRefsCalls++
	var out []route.StopRef
	for _, ref := range f.refs {
		if slices.Contains(routeIDs, ref.RouteID) {
			out = append(out, ref)
		}
	}
	return out, nil
}

func (f *fakeRepo) Owner(_ context.Context, id int64) (*int64, error) {
	r, ok := f.routes[id]
	if !ok {
		return nil, route.ErrNotFound
	}
	return r.OwnerID, nil
}

func (f *fakeRepo) Delete(_ context.Context, id int64) error {
	delete(f.routes, id)
	return nil
}

// fakeTx помечает ctx как транзакционный; при ошибке откатывает состояние репозитория.
type fakeTx struct{ repo *fakeRepo }

func (t fakeTx) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	saved := *t.repo
	saved.routes = make(map[int64]route.Route, len(t.repo.routes))
	for k, v := range t.repo.routes {
		saved.routes[k] = v
	}
	saved.refs = slices.Clone(t.repo.refs)
	if err := fn(context.WithValue(ctx, txKey{}, true)); err != nil {
		t.repo.routes, t.repo.refs = saved.routes, saved.refs
		return err
	}
	return nil
}

// fakePlaces — каталог из трёх объектов; «hidden» снят с публикации.
type fakePlaces struct{ byIDsCalls int }

var catalog = []Place{
	{ID: 1, Slug: "birzha", Title: "Биржа", Lat: 46.35, Lng: 48.03},
	{ID: 2, Slug: "kreml", Title: "Кремль", Lat: 46.34, Lng: 48.02},
	{ID: 3, Slug: "hidden", Title: "Скрытый", Lat: 46.33, Lng: 48.01},
}

func (f *fakePlaces) PublishedPlaces(_ context.Context, slugs []string) ([]Place, error) {
	var out []Place
	// Обратный порядок — проверяем, что сервис сам восстанавливает порядок пользователя.
	for i := len(catalog) - 1; i >= 0; i-- {
		if p := catalog[i]; p.Slug != "hidden" && slices.Contains(slugs, p.Slug) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakePlaces) PlacesByIDs(_ context.Context, ids []int64) ([]Place, error) {
	f.byIDsCalls++
	var out []Place
	for _, p := range catalog {
		if slices.Contains(ids, p.ID) {
			out = append(out, p)
		}
	}
	return out, nil
}

type fakeRouter struct {
	fail   bool
	points []Point
}

func (f *fakeRouter) Route(_ context.Context, points []Point) (RouteLine, error) {
	if f.fail {
		return RouteLine{}, errors.New("OSRM: таймаут")
	}
	f.points = points
	return RouteLine{DistanceM: 420.5, DurationS: 300, Geometry: json.RawMessage(`{"type":"LineString"}`)}, nil
}

type fixture struct {
	svc    *Service
	repo   *fakeRepo
	places *fakePlaces
	router *fakeRouter
}

func newFixture() fixture {
	repo := newFakeRepo()
	f := fixture{repo: repo, places: &fakePlaces{}, router: &fakeRouter{}}
	f.svc = NewService(repo, f.places, f.router, fakeTx{repo: repo})
	return f
}

func TestBuild(t *testing.T) {
	f := newFixture()
	ctx := context.Background()

	r, err := f.svc.Build(ctx, []string{"kreml", "birzha"})
	if err != nil {
		t.Fatal(err)
	}
	if r.DistanceM != 420.5 || len(r.Geometry) == 0 || len(r.Stops) != 2 {
		t.Fatalf("маршрут: %+v", r)
	}
	if r.Stops[0].Slug != "kreml" || r.Stops[0].Position != 1 || r.Stops[1].Slug != "birzha" || r.Stops[1].Position != 2 {
		t.Fatalf("порядок остановок: %+v", r.Stops)
	}
	if f.router.points[0] != (Point{Lat: 46.34, Lng: 48.02}) {
		t.Fatalf("точки маршрутизатору: %+v", f.router.points)
	}

	if _, err := f.svc.Build(ctx, []string{"birzha"}); !errors.Is(err, route.ErrStopCount) {
		t.Fatalf("одна остановка: %v", err)
	}
	if _, err := f.svc.Build(ctx, []string{"birzha", "hidden"}); !errs.Is(err, errs.Invalid) {
		t.Fatalf("неопубликованный объект: %v", err)
	}

	f.router.fail = true
	if _, err := f.svc.Build(ctx, []string{"birzha", "kreml"}); !errors.Is(err, route.ErrRouterUnavailable) {
		t.Fatalf("маршрутизатор недоступен: %v", err)
	}
}

func TestSaveAndGet(t *testing.T) {
	f := newFixture()
	ctx := context.Background()

	saved, err := f.svc.Save(ctx, route.SaveInput{Slug: " Center ", Title: " Центр ", Objects: []string{"birzha", "kreml"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == 0 || saved.Slug != "center" || saved.Title != "Центр" || len(saved.Stops) != 2 || saved.Stops[1].Title != "Кремль" {
		t.Fatalf("сохранённый: %+v", saved)
	}
	if f.repo.writesOutside != 0 {
		t.Fatal("запись вне транзакции")
	}
	if _, err := f.svc.Save(ctx, route.SaveInput{Slug: "center", Title: "Дубль", Objects: []string{"birzha", "kreml"}}, nil); !errors.Is(err, route.ErrSlugTaken) {
		t.Fatalf("дубликат slug: %v", err)
	}
	if _, err := f.svc.Save(ctx, route.SaveInput{Title: "Без slug", Objects: []string{"birzha", "kreml"}}, nil); !errors.Is(err, route.ErrSlugRequired) {
		t.Fatalf("без slug: %v", err)
	}

	// Маршрут пользователя: служебный slug, виден только владельцу и администратору.
	owner := int64(7)
	mine, err := f.svc.Save(ctx, route.SaveInput{Title: "Мой", Objects: []string{"kreml", "birzha"}}, &owner)
	if err != nil || mine.OwnerID == nil || *mine.OwnerID != 7 || mine.Slug == "" {
		t.Fatalf("маршрут пользователя: %+v, %v", mine, err)
	}
	if _, err := f.svc.Get(ctx, mine.Slug, nil); !errors.Is(err, route.ErrNotFound) {
		t.Fatalf("гость видит чужой маршрут: %v", err)
	}
	if _, err := f.svc.Get(ctx, mine.Slug, &route.Actor{UserID: 8}); !errors.Is(err, route.ErrNotFound) {
		t.Fatalf("другой пользователь видит чужой маршрут: %v", err)
	}
	if got, err := f.svc.Get(ctx, mine.Slug, &route.Actor{UserID: 7}); err != nil || got.ID != mine.ID {
		t.Fatalf("владелец: %+v, %v", got, err)
	}
	if _, err := f.svc.Get(ctx, mine.Slug, &route.Actor{UserID: 1, IsAdmin: true}); err != nil {
		t.Fatalf("администратор: %v", err)
	}
	if _, err := f.svc.Get(ctx, "center", nil); err != nil {
		t.Fatalf("тематический гостю: %v", err)
	}
}

func TestSaveRollsBack(t *testing.T) {
	f := newFixture()
	f.repo.failAddStops = true
	_, err := f.svc.Save(context.Background(), route.SaveInput{Slug: "x", Title: "X", Objects: []string{"birzha", "kreml"}}, nil)
	if err == nil || len(f.repo.routes) != 0 {
		t.Fatalf("после сбоя маршрут остался: %v, %+v", err, f.repo.routes)
	}
}

func TestListsFillStopsWithoutNPlusOne(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	for _, slug := range []string{"a", "b", "c"} {
		if _, err := f.svc.Save(ctx, route.SaveInput{Slug: slug, Title: slug, Objects: []string{"birzha", "kreml"}}, nil); err != nil {
			t.Fatal(err)
		}
	}
	f.repo.stopRefsCalls, f.places.byIDsCalls = 0, 0

	list, err := f.svc.ListThematic(ctx)
	if err != nil || len(list) != 3 {
		t.Fatalf("список: %+v, %v", list, err)
	}
	if f.repo.stopRefsCalls != 1 || f.places.byIDsCalls != 1 {
		t.Fatalf("запросов точек %d, к каталогу %d — ожидалось по одному", f.repo.stopRefsCalls, f.places.byIDsCalls)
	}
	for _, r := range list {
		if len(r.Stops) != 2 || r.Stops[0].Slug != "birzha" || r.Geometry != nil {
			t.Fatalf("маршрут списка: %+v", r)
		}
	}

	// Пустой список: ни точек, ни каталога не спрашиваем, остановки — пустой срез.
	mine, err := f.svc.ListByOwner(ctx, 99)
	if err != nil || len(mine) != 0 || f.repo.stopRefsCalls != 1 {
		t.Fatalf("пустой список: %+v, %v", mine, err)
	}
}

func TestDelete(t *testing.T) {
	f := newFixture()
	ctx := context.Background()
	owner := int64(7)
	thematic, _ := f.svc.Save(ctx, route.SaveInput{Slug: "t", Title: "T", Objects: []string{"birzha", "kreml"}}, nil)
	mine, _ := f.svc.Save(ctx, route.SaveInput{Title: "M", Objects: []string{"birzha", "kreml"}}, &owner)

	if err := f.svc.Delete(ctx, mine.ID, route.Actor{UserID: 8}); !errors.Is(err, route.ErrDeleteDenied) {
		t.Fatalf("чужой маршрут: %v", err)
	}
	if err := f.svc.Delete(ctx, thematic.ID, route.Actor{UserID: 7}); !errors.Is(err, route.ErrDeleteDenied) {
		t.Fatalf("тематический пользователем: %v", err)
	}
	if err := f.svc.Delete(ctx, mine.ID, route.Actor{UserID: 7}); err != nil {
		t.Fatalf("свой: %v", err)
	}
	if err := f.svc.Delete(ctx, thematic.ID, route.Actor{UserID: 1, IsAdmin: true}); err != nil {
		t.Fatalf("тематический администратором: %v", err)
	}
	if err := f.svc.Delete(ctx, thematic.ID, route.Actor{UserID: 1, IsAdmin: true}); !errors.Is(err, route.ErrNotFound) {
		t.Fatalf("повторное удаление: %v", err)
	}
}
