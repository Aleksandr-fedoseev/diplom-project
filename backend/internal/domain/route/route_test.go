package route

import (
	"errors"
	"regexp"
	"testing"
)

func TestCheckStopCount(t *testing.T) {
	for n, ok := range map[int]bool{0: false, 1: false, 2: true, 25: true, 26: false} {
		if err := CheckStopCount(n); (err == nil) != ok {
			t.Errorf("CheckStopCount(%d) = %v", n, err)
		}
	}
}

func TestSaveInputNormalize(t *testing.T) {
	in := SaveInput{Slug: "  Old-Town ", Title: " Старый город "}
	if err := in.Normalize(nil); err != nil || in.Slug != "old-town" || in.Title != "Старый город" {
		t.Fatalf("тематический: %+v, %v", in, err)
	}

	if err := (&SaveInput{Slug: "x", Title: "  "}).Normalize(nil); !errors.Is(err, ErrTitleRequired) {
		t.Fatalf("без названия: %v", err)
	}
	if err := (&SaveInput{Title: "Маршрут"}).Normalize(nil); !errors.Is(err, ErrSlugRequired) {
		t.Fatalf("тематический без slug: %v", err)
	}

	// У маршрута пользователя slug служебный, присланный игнорируется.
	owner := int64(42)
	in = SaveInput{Slug: "my-slug", Title: "Мой"}
	if err := in.Normalize(&owner); err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^u42-[0-9a-f]{8}$`).MatchString(in.Slug) {
		t.Fatalf("служебный slug: %q", in.Slug)
	}
	if NewUserSlug(42) == NewUserSlug(42) {
		t.Fatal("служебные slug'и должны различаться")
	}
}

func TestCanView(t *testing.T) {
	owner := int64(1)
	thematic := Route{}
	mine := Route{OwnerID: &owner}

	if !thematic.CanView(nil) {
		t.Error("тематический маршрут виден гостю")
	}
	if mine.CanView(nil) || mine.CanView(&Actor{UserID: 2}) {
		t.Error("чужой маршрут пользователя не виден гостю и другим")
	}
	if !mine.CanView(&Actor{UserID: 1}) || !mine.CanView(&Actor{UserID: 3, IsAdmin: true}) {
		t.Error("маршрут виден владельцу и администратору")
	}
}

func TestCanDelete(t *testing.T) {
	owner := int64(1)
	cases := []struct {
		name  string
		owner *int64
		actor Actor
		want  bool
	}{
		{"владелец", &owner, Actor{UserID: 1}, true},
		{"чужой", &owner, Actor{UserID: 2}, false},
		{"админ — чужой", &owner, Actor{UserID: 2, IsAdmin: true}, true},
		{"тематический — пользователь", nil, Actor{UserID: 1}, false},
		{"тематический — админ", nil, Actor{UserID: 1, IsAdmin: true}, true},
	}
	for _, c := range cases {
		if got := CanDelete(c.owner, c.actor); got != c.want {
			t.Errorf("%s: %v", c.name, got)
		}
	}
}
