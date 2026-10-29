package catalog

import (
	"context"
	"errors"
	"testing"

	"heritage-backend/internal/domain/catalog"
	"heritage-backend/internal/domain/errs"
)

// fakeObjects — хранилище объектов в памяти: проверяем сценарии без БД.
type fakeObjects struct {
	bySlug     map[string]catalog.Object
	lastFilter catalog.Filter
	created    []catalog.ObjectInput
}

func newFakeObjects(objs ...catalog.Object) *fakeObjects {
	f := &fakeObjects{bySlug: map[string]catalog.Object{}}
	for _, o := range objs {
		f.bySlug[o.Slug] = o
	}
	return f
}

func (f *fakeObjects) List(_ context.Context, flt catalog.Filter) ([]catalog.Object, error) {
	f.lastFilter = flt
	var out []catalog.Object
	for _, o := range f.bySlug {
		if o.IsPublished || flt.IncludeUnpublished {
			out = append(out, o)
		}
	}
	return out, nil
}

func (f *fakeObjects) BySlug(_ context.Context, slug string) (catalog.Object, error) {
	o, ok := f.bySlug[slug]
	if !ok {
		return catalog.Object{}, catalog.ErrObjectNotFound
	}
	return o, nil
}

func (f *fakeObjects) Create(_ context.Context, in catalog.ObjectInput) (catalog.Object, error) {
	if _, ok := f.bySlug[in.Slug]; ok {
		return catalog.Object{}, catalog.ErrSlugTaken
	}
	f.created = append(f.created, in)
	o := catalog.Object{ID: int64(len(f.bySlug) + 1), Slug: in.Slug, Title: in.Title, IsPublished: *in.IsPublished}
	f.bySlug[o.Slug] = o
	return o, nil
}

func (f *fakeObjects) Update(_ context.Context, id int64, in catalog.ObjectInput) (catalog.Object, error) {
	for slug, o := range f.bySlug {
		if o.ID == id {
			delete(f.bySlug, slug)
			o.Slug, o.Title = in.Slug, in.Title
			f.bySlug[o.Slug] = o
			return o, nil
		}
	}
	return catalog.Object{}, catalog.ErrObjectNotFound
}

func (f *fakeObjects) Delete(context.Context, int64) error { return nil }
func (f *fakeObjects) Exists(context.Context, int64) (bool, error) {
	return true, nil
}
func (f *fakeObjects) PublishedPlaces(context.Context, []string) ([]catalog.Place, error) {
	return nil, nil
}
func (f *fakeObjects) PlacesByIDs(context.Context, []int64) ([]catalog.Place, error) {
	return nil, nil
}

type fakeCategories struct{ created []catalog.CategoryInput }

func (f *fakeCategories) List(context.Context) ([]catalog.Category, error) { return nil, nil }
func (f *fakeCategories) Create(_ context.Context, in catalog.CategoryInput) (catalog.Category, error) {
	f.created = append(f.created, in)
	return catalog.Category{ID: 1, Slug: in.Slug, Name: in.Name}, nil
}
func (f *fakeCategories) Update(_ context.Context, id int32, in catalog.CategoryInput) (catalog.Category, error) {
	return catalog.Category{ID: id, Slug: in.Slug, Name: in.Name}, nil
}
func (f *fakeCategories) Delete(context.Context, int32) error { return nil }

var (
	published = catalog.Object{ID: 1, Slug: "birzha", IsPublished: true}
	hidden    = catalog.Object{ID: 2, Slug: "chernovik", IsPublished: false}
	guest     = catalog.Viewer{}
	moderator = catalog.Viewer{Moderator: true}
)

func TestGetVisible(t *testing.T) {
	svc := NewService(newFakeObjects(published, hidden), &fakeCategories{})
	ctx := context.Background()

	if _, err := svc.GetVisible(ctx, "chernovik", guest); !errors.Is(err, catalog.ErrObjectNotFound) {
		t.Fatalf("гость видит неопубликованный объект: %v", err)
	}
	if o, err := svc.GetVisible(ctx, "chernovik", moderator); err != nil || o.ID != hidden.ID {
		t.Fatalf("модератор: %+v, %v", o, err)
	}
	if _, err := svc.GetVisible(ctx, "birzha", guest); err != nil {
		t.Fatal(err)
	}
	// Другим модулям объект доступен без проверки видимости.
	if _, err := svc.GetBySlug(ctx, "chernovik"); err != nil {
		t.Fatal(err)
	}
}

func TestListRespectsViewer(t *testing.T) {
	objects := newFakeObjects(published, hidden)
	svc := NewService(objects, &fakeCategories{})
	ctx := context.Background()

	list, err := svc.List(ctx, catalog.Filter{IncludeUnpublished: true}, guest)
	if err != nil || len(list) != 1 || objects.lastFilter.IncludeUnpublished {
		t.Fatalf("гость попросил неопубликованные: %+v, %v", list, err)
	}
	list, err = svc.List(ctx, catalog.Filter{IncludeUnpublished: true}, moderator)
	if err != nil || len(list) != 2 {
		t.Fatalf("модератор: %+v, %v", list, err)
	}
}

func TestCreateValidatesBeforeRepository(t *testing.T) {
	objects := newFakeObjects(published)
	svc := NewService(objects, &fakeCategories{})
	ctx := context.Background()

	_, err := svc.Create(ctx, catalog.ObjectInput{Slug: "bad slug", Title: "x"})
	if !errs.Is(err, errs.Invalid) || len(objects.created) != 0 {
		t.Fatalf("неверные данные дошли до хранилища: %v", err)
	}

	in := catalog.ObjectInput{
		Slug: " Dom ", Title: "Дом", Category: "manor", ProtectionStatus: catalog.StatusLost,
		Address: "ул. Тестовая, 1", Coordinates: catalog.Coordinates{Lat: 46, Lng: 48},
	}
	o, err := svc.Create(ctx, in)
	if err != nil || o.Slug != "dom" || !o.IsPublished || objects.created[0].Tags == nil {
		t.Fatalf("создание: %+v, %v", o, err)
	}
	if _, err := svc.Create(ctx, in); !errors.Is(err, catalog.ErrSlugTaken) {
		t.Fatalf("повторный slug: %v", err)
	}

	if _, err := svc.Update(ctx, 1, catalog.ObjectInput{}); !errs.Is(err, errs.Invalid) {
		t.Fatalf("изменение с пустыми полями: %v", err)
	}
}

func TestCreateCategoryNormalizes(t *testing.T) {
	cats := &fakeCategories{}
	svc := NewService(newFakeObjects(), cats)
	c, err := svc.CreateCategory(context.Background(), catalog.CategoryInput{Slug: " Modern ", Name: " Модерн "})
	if err != nil || c.Slug != "modern" || c.Name != "Модерн" {
		t.Fatalf("категория: %+v, %v", c, err)
	}
	if _, err := svc.CreateCategory(context.Background(), catalog.CategoryInput{Slug: "x"}); !errs.Is(err, errs.Invalid) || len(cats.created) != 1 {
		t.Fatalf("категория без названия: %v", err)
	}
}
