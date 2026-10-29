package postgres_test

import (
	"context"
	"errors"
	"testing"

	"heritage-backend/internal/domain/catalog"
	"heritage-backend/internal/infrastructure/postgres"
	repo "heritage-backend/internal/repository/postgres"
	"heritage-backend/internal/testdb"
)

// Тесты каталога опираются на стартовые данные из миграции 00002_seed_catalog.sql.

func slugSet(list []catalog.Object) map[string]bool {
	m := make(map[string]bool, len(list))
	for _, o := range list {
		m[o.Slug] = true
	}
	return m
}

func TestObjectsListFilters(t *testing.T) {
	objects := repo.NewObjects(testdb.Connect(t))
	ctx := context.Background()

	all, err := objects.List(ctx, catalog.Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if !slugSet(all)["birzha"] || all[0].FullHistory != "" || all[0].DistanceM != nil || all[0].UpdatedAt.IsZero() {
		t.Fatalf("список без фильтров: %+v", all[0])
	}

	modern, err := objects.List(ctx, catalog.Filter{Category: "modern"})
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range modern {
		if o.Category != "modern" {
			t.Fatalf("фильтр по категории пропустил %s (%s)", o.Slug, o.Category)
		}
	}

	found, err := objects.List(ctx, catalog.Filter{Query: "никольская", Status: "regional"})
	if err != nil {
		t.Fatal(err)
	}
	if s := slugSet(found); !s["dom-agamzhanovoy"] || s["azovo-donskoy-bank"] {
		t.Fatalf("поиск по адресу и статусу: %v", s)
	}

	// Биржа (48.0285, 46.3538) — в 500 м от точки рядом с ней, остальные дальше.
	near, err := objects.List(ctx, catalog.Filter{Near: &catalog.Coordinates{Lat: 46.3538, Lng: 48.0286}, RadiusM: 500})
	if err != nil {
		t.Fatal(err)
	}
	if len(near) < 2 || near[0].Slug != "birzha" {
		t.Fatalf("поиск рядом должен вернуть ближайшие первыми: %+v", near)
	}
	for i, o := range near {
		if o.DistanceM == nil || *o.DistanceM > 500 || (i > 0 && *near[i-1].DistanceM > *o.DistanceM) {
			t.Fatalf("радиус и порядок по расстоянию нарушены на %s", o.Slug)
		}
	}

	box, err := objects.List(ctx, catalog.Filter{BBox: &[4]float64{48.028, 46.353, 48.029, 46.354}})
	if err != nil {
		t.Fatal(err)
	}
	if s := slugSet(box); !s["birzha"] || s["dom-agamzhanovoy"] {
		t.Fatalf("видимая область карты: %v", s)
	}
}

func TestObjectsCreateUpdateDelete(t *testing.T) {
	pool := testdb.Connect(t)
	objects := repo.NewObjects(pool)
	ctx := context.Background()
	slug := testdb.Unique("test-object")
	architect := "Неизвестен"
	unpublished := false

	in := catalog.ObjectInput{
		Slug: slug, Title: "Тестовый дом", Category: "manor", ProtectionStatus: catalog.StatusIdentified,
		Architect: &architect, Address: "ул. Тестовая, 1", Tags: []string{},
		Coordinates: catalog.Coordinates{Lat: 46.35, Lng: 48.03}, FullHistory: "История", IsPublished: &unpublished,
	}
	o, err := objects.Create(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if o.ID == 0 || o.Category != "manor" || o.IsPublished || o.Coordinates.Lat != 46.35 ||
		o.Architect == nil || *o.Architect != architect || o.YearBuilt != nil || o.FullHistory != "История" {
		t.Fatalf("созданный объект: %+v", o)
	}

	// Неопубликованный объект попадает в список только по просьбе.
	if list, _ := objects.List(ctx, catalog.Filter{Query: "Тестовый дом"}); slugSet(list)[slug] {
		t.Fatal("неопубликованный объект в общем списке")
	}
	if list, _ := objects.List(ctx, catalog.Filter{Query: "Тестовый дом", IncludeUnpublished: true}); !slugSet(list)[slug] {
		t.Fatal("неопубликованного объекта нет в списке модератора")
	}

	if _, err := objects.Create(ctx, in); !errors.Is(err, catalog.ErrSlugTaken) {
		t.Fatalf("повторный slug: %v", err)
	}
	bad := in
	bad.Slug, bad.Category = testdb.Unique("bad"), "no-such-category"
	if _, err := objects.Create(ctx, bad); !errors.Is(err, catalog.ErrUnknownCategory) {
		t.Fatalf("неизвестная категория: %v", err)
	}

	published := true
	in.Title, in.IsPublished, in.Architect = "Переименованный дом", &published, nil
	o, err = objects.Update(ctx, o.ID, in)
	if err != nil {
		t.Fatal(err)
	}
	if o.Title != "Переименованный дом" || !o.IsPublished || o.Architect != nil {
		t.Fatalf("изменённый объект: %+v", o)
	}
	if _, err := objects.Update(ctx, -1, in); !errors.Is(err, catalog.ErrObjectNotFound) {
		t.Fatalf("изменение несуществующего: %v", err)
	}
	upd := in
	upd.Category = "no-such-category"
	if _, err := objects.Update(ctx, o.ID, upd); !errors.Is(err, catalog.ErrUnknownCategory) {
		t.Fatalf("изменение на неизвестную категорию: %v", err)
	}

	places, err := objects.PublishedPlaces(ctx, []string{slug, "birzha"})
	if err != nil || len(places) != 2 {
		t.Fatalf("PublishedPlaces: %+v, %v", places, err)
	}
	byID, err := objects.PlacesByIDs(ctx, []int64{o.ID})
	if err != nil || len(byID) != 1 || byID[0].Lng != 48.03 || byID[0].Slug != slug {
		t.Fatalf("PlacesByIDs: %+v, %v", byID, err)
	}

	if err := objects.Delete(ctx, o.ID); err != nil {
		t.Fatal(err)
	}
	if ok, err := objects.Exists(ctx, o.ID); err != nil || ok {
		t.Fatalf("после удаления Exists = %v, %v", ok, err)
	}
	if err := objects.Delete(ctx, o.ID); !errors.Is(err, catalog.ErrObjectNotFound) {
		t.Fatalf("повторное удаление: %v", err)
	}
	if _, err := objects.BySlug(ctx, slug); !errors.Is(err, catalog.ErrObjectNotFound) {
		t.Fatalf("удалённый объект: %v", err)
	}
}

// Объект, на который ссылается маршрут, удалить нельзя.
func TestObjectsDeleteInRoute(t *testing.T) {
	pool := testdb.Connect(t)
	objects := repo.NewObjects(pool)
	ctx := context.Background()

	o, err := objects.Create(ctx, catalog.ObjectInput{
		Slug: testdb.Unique("in-route"), Title: "Дом", Category: "manor", ProtectionStatus: catalog.StatusLost,
		Address: "ул. Тестовая, 2", Tags: []string{}, Coordinates: catalog.Coordinates{Lat: 46.35, Lng: 48.03},
	})
	if err != nil {
		t.Fatal(err)
	}
	var routeID int64
	err = pool.QueryRow(ctx, `INSERT INTO routes (slug, title, distance_m, duration_s, geometry)
		VALUES ($1, 'Тест', 1, 1, 'SRID=4326;LINESTRING(48.03 46.35, 48.04 46.36)') RETURNING id`,
		testdb.Unique("route")).Scan(&routeID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM routes WHERE id = $1`, routeID) })
	if _, err := pool.Exec(ctx, `INSERT INTO route_points (route_id, position, object_id) VALUES ($1, 0, $2)`, routeID, o.ID); err != nil {
		t.Fatal(err)
	}
	if err := objects.Delete(ctx, o.ID); !errors.Is(err, catalog.ErrObjectInRoute) {
		t.Fatalf("удаление объекта из маршрута: %v", err)
	}
}

// Объект, созданный в откаченной транзакции, не сохраняется, а внутри неё виден.
func TestObjectsCreateInTx(t *testing.T) {
	pool := testdb.Connect(t)
	objects := repo.NewObjects(pool)
	ctx := context.Background()
	slug := testdb.Unique("tx-object")

	boom := errors.New("откат")
	err := postgres.NewTxManager(pool).Do(ctx, func(ctx context.Context) error {
		o, err := objects.Create(ctx, catalog.ObjectInput{
			Slug: slug, Title: "Дом", Category: "manor", ProtectionStatus: catalog.StatusLost,
			Address: "ул. Тестовая, 3", Tags: []string{}, Coordinates: catalog.Coordinates{Lat: 46.35, Lng: 48.03},
		})
		if err != nil || o.Slug != slug {
			t.Fatalf("создание в транзакции: %+v, %v", o, err)
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatal(err)
	}
	if _, err := objects.BySlug(ctx, slug); !errors.Is(err, catalog.ErrObjectNotFound) {
		t.Fatalf("объект из откаченной транзакции сохранился: %v", err)
	}
}

func TestCategoriesRepository(t *testing.T) {
	cats := repo.NewCategories(testdb.Connect(t))
	ctx := context.Background()
	slug := testdb.Unique("cat")

	c, err := cats.Create(ctx, catalog.CategoryInput{Slug: slug, Name: "Тестовая"})
	if err != nil || c.ID == 0 {
		t.Fatalf("создание: %+v, %v", c, err)
	}
	if _, err := cats.Create(ctx, catalog.CategoryInput{Slug: slug, Name: "Дубль"}); !errors.Is(err, catalog.ErrCategorySlugTaken) {
		t.Fatalf("дубликат: %v", err)
	}
	if c, err = cats.Update(ctx, c.ID, catalog.CategoryInput{Slug: slug, Name: "Новая"}); err != nil || c.Name != "Новая" {
		t.Fatalf("изменение: %+v, %v", c, err)
	}
	if _, err := cats.Update(ctx, -1, catalog.CategoryInput{Slug: slug + "x", Name: "x"}); !errors.Is(err, catalog.ErrCategoryNotFound) {
		t.Fatalf("изменение несуществующей: %v", err)
	}
	if _, err := cats.Update(ctx, c.ID, catalog.CategoryInput{Slug: "manor", Name: "x"}); !errors.Is(err, catalog.ErrCategorySlugTaken) {
		t.Fatalf("изменение на занятый slug: %v", err)
	}

	list, err := cats.List(ctx)
	if err != nil || len(list) == 0 {
		t.Fatalf("список: %v", err)
	}
	var manorID int32
	for _, x := range list {
		if x.Slug == "manor" {
			manorID = x.ID
		}
	}
	if err := cats.Delete(ctx, manorID); !errors.Is(err, catalog.ErrCategoryNotEmpty) {
		t.Fatalf("удаление непустой категории: %v", err)
	}
	if err := cats.Delete(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	if err := cats.Delete(ctx, c.ID); !errors.Is(err, catalog.ErrCategoryNotFound) {
		t.Fatalf("повторное удаление: %v", err)
	}
}
