package catalog

import (
	"context"
	"errors"
	"testing"

	"heritage-backend/internal/httpx"
	"heritage-backend/internal/testdb"
)

// Тесты опираются на стартовые данные из миграции 00002_seed_catalog.sql.

func wantStatus(t *testing.T, err error, status int) {
	t.Helper()
	var he *httpx.Error
	if !errors.As(err, &he) || he.Status != status {
		t.Fatalf("ожидалась ошибка %d, получено: %v", status, err)
	}
}

func slugs(list []Object) map[string]bool {
	m := make(map[string]bool, len(list))
	for _, o := range list {
		m[o.Slug] = true
	}
	return m
}

func TestListFilters(t *testing.T) {
	svc := NewService(testdb.Connect(t))
	ctx := context.Background()

	all, err := svc.List(ctx, Filter{})
	if err != nil {
		t.Fatal(err)
	}
	if !slugs(all)["birzha"] || all[0].FullHistory != "" || all[0].DistanceM != nil {
		t.Fatalf("список без фильтров: %+v", all[0])
	}

	modern, err := svc.List(ctx, Filter{Category: "modern"})
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range modern {
		if o.Category != "modern" {
			t.Fatalf("фильтр по категории пропустил %s (%s)", o.Slug, o.Category)
		}
	}

	found, err := svc.List(ctx, Filter{Query: "никольская", Status: "regional"})
	if err != nil {
		t.Fatal(err)
	}
	if s := slugs(found); !s["dom-agamzhanovoy"] || s["azovo-donskoy-bank"] {
		t.Fatalf("поиск по адресу и статусу: %v", s)
	}

	// Биржа (48.0285, 46.3538) — в 500 м от точки рядом с ней, банк чуть дальше.
	near, err := svc.List(ctx, Filter{Near: &Coordinates{Lat: 46.3538, Lng: 48.0286}, RadiusM: 500})
	if err != nil {
		t.Fatal(err)
	}
	if len(near) < 2 || near[0].Slug != "birzha" || near[0].DistanceM == nil || *near[0].DistanceM > *near[1].DistanceM {
		t.Fatalf("поиск рядом должен вернуть ближайшие первыми: %+v", near)
	}

	box, err := svc.List(ctx, Filter{BBox: &[4]float64{48.028, 46.353, 48.029, 46.354}})
	if err != nil {
		t.Fatal(err)
	}
	if s := slugs(box); !s["birzha"] || s["dom-agamzhanovoy"] {
		t.Fatalf("видимая область карты: %v", s)
	}
}

func TestCreateUpdateDelete(t *testing.T) {
	svc := NewService(testdb.Connect(t))
	ctx := context.Background()
	slug := testdb.Unique("test-object")
	architect := "Неизвестен"

	in := ObjectInput{
		Slug: slug, Title: "Тестовый дом", Category: "manor", ProtectionStatus: "identified",
		Architect: &architect, Address: "ул. Тестовая, 1",
		Coordinates: Coordinates{Lat: 46.35, Lng: 48.03}, FullHistory: "История",
	}
	o, err := svc.Create(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if o.ID == 0 || o.Category != "manor" || !o.IsPublished || o.Coordinates.Lat != 46.35 ||
		o.Architect == nil || *o.Architect != architect || o.YearBuilt != nil || o.FullHistory != "История" {
		t.Fatalf("созданный объект: %+v", o)
	}

	_, err = svc.Create(ctx, in)
	wantStatus(t, err, 409)

	bad := in
	bad.Slug, bad.Category = testdb.Unique("bad"), "no-such-category"
	_, err = svc.Create(ctx, bad)
	wantStatus(t, err, 400)

	unpublished := false
	in.Title, in.IsPublished, in.Architect = "Переименованный дом", &unpublished, nil
	o, err = svc.Update(ctx, o.ID, in)
	if err != nil {
		t.Fatal(err)
	}
	if o.Title != "Переименованный дом" || o.IsPublished || o.Architect != nil {
		t.Fatalf("изменённый объект: %+v", o)
	}
	_, err = svc.Update(ctx, -1, in)
	wantStatus(t, err, 404)

	// Неопубликованный объект не виден другим модулям при построении маршрута.
	places, err := svc.PublishedPlaces(ctx, []string{slug, "birzha"})
	if err != nil {
		t.Fatal(err)
	}
	if len(places) != 1 || places[0].Slug != "birzha" {
		t.Fatalf("PublishedPlaces: %+v", places)
	}
	byID, err := svc.PlacesByIDs(ctx, []int64{o.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(byID) != 1 || byID[0].Lng != 48.03 {
		t.Fatalf("PlacesByIDs: %+v", byID)
	}

	if err := svc.Delete(ctx, o.ID); err != nil {
		t.Fatal(err)
	}
	if ok, err := svc.Exists(ctx, o.ID); err != nil || ok {
		t.Fatalf("после удаления Exists = %v, %v", ok, err)
	}
	wantStatus(t, svc.Delete(ctx, o.ID), 404)
}
