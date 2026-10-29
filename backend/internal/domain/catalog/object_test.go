package catalog

import (
	"testing"

	"heritage-backend/internal/domain/errs"
)

func validInput() ObjectInput {
	return ObjectInput{
		Slug: "  Birzha ", Title: "Биржа", Category: "modern", ProtectionStatus: StatusRegional,
		Address: "ул. Советская, 1", Coordinates: Coordinates{Lat: 46.35, Lng: 48.03},
	}
}

func TestObjectInputNormalize(t *testing.T) {
	in := validInput()
	if err := in.Normalize(); err != nil {
		t.Fatal(err)
	}
	if in.Slug != "birzha" || in.Tags == nil || in.IsPublished == nil || !*in.IsPublished {
		t.Fatalf("нормализация: %+v", in)
	}

	bad := map[string]func(*ObjectInput){
		"пустой slug":        func(in *ObjectInput) { in.Slug = "  " },
		"slug с пробелом":    func(in *ObjectInput) { in.Slug = "dom kupca" },
		"slug со слэшем":     func(in *ObjectInput) { in.Slug = "a/b" },
		"без названия":       func(in *ObjectInput) { in.Title = " " },
		"без адреса":         func(in *ObjectInput) { in.Address = "" },
		"неизвестный статус": func(in *ObjectInput) { in.ProtectionStatus = "unesco" },
		"широта за пределом": func(in *ObjectInput) { in.Coordinates.Lat = 91 },
		"долгота за пределом": func(in *ObjectInput) {
			in.Coordinates.Lng = -181
		},
	}
	for name, mutate := range bad {
		in := validInput()
		mutate(&in)
		if err := in.Normalize(); !errs.Is(err, errs.Invalid) {
			t.Errorf("%s: ожидалась ошибка Invalid, получено %v", name, err)
		}
	}
}

func TestCategoryInputNormalize(t *testing.T) {
	in := CategoryInput{Slug: " Modern ", Name: " Модерн "}
	if err := in.Normalize(); err != nil || in.Slug != "modern" || in.Name != "Модерн" {
		t.Fatalf("нормализация: %+v, %v", in, err)
	}
	for _, bad := range []CategoryInput{{Slug: "", Name: "x"}, {Slug: "a#b", Name: "x"}, {Slug: "ok", Name: " "}} {
		if err := bad.Normalize(); !errs.Is(err, errs.Invalid) {
			t.Errorf("ожидалась ошибка для %+v", bad)
		}
	}
}

func TestVisibility(t *testing.T) {
	guest, mod := Viewer{}, Viewer{Moderator: true}
	hidden := Object{IsPublished: false}
	if hidden.VisibleTo(guest) || !hidden.VisibleTo(mod) || !(Object{IsPublished: true}).VisibleTo(guest) {
		t.Fatal("неопубликованный объект виден только модератору")
	}

	f := Filter{IncludeUnpublished: true}
	if f.ForViewer(guest).IncludeUnpublished || !f.ForViewer(mod).IncludeUnpublished {
		t.Fatal("неопубликованные в списке — только для модератора")
	}
	if (Filter{}).ForViewer(mod).IncludeUnpublished {
		t.Fatal("без просьбы неопубликованные не показываются и модератору")
	}
}
