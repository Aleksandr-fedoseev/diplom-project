package submission

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

func newObject() NewSubmission {
	return NewSubmission{
		Kind:        KindNewObject,
		Title:       " Дом купца Тестова ",
		Description: "Двухэтажный деревянный дом.",
		Address:     "ул. Тестовая, 1",
		Coordinates: &Coordinates{Lat: 46.35, Lng: 48.04},
		Period:      "1890-е",
		Source:      "Личный архив",
	}
}

func TestNewSubmissionNormalize(t *testing.T) {
	cases := map[string]struct {
		mutate func(*NewSubmission)
		want   error
	}{
		"неизвестный вид":        {func(in *NewSubmission) { in.Kind = "other" }, ErrInvalidKind},
		"без источника":          {func(in *NewSubmission) { in.Source = "  " }, ErrSourceRequired},
		"без названия":           {func(in *NewSubmission) { in.Title = "" }, ErrTitleRequired},
		"без места":              {func(in *NewSubmission) { in.Address, in.Coordinates = "", nil }, ErrPlaceRequired},
		"без периода":            {func(in *NewSubmission) { in.Period = "" }, ErrPeriodRequired},
		"плохие координаты":      {func(in *NewSubmission) { in.Coordinates = &Coordinates{Lat: 100, Lng: 48} }, ErrInvalidCoordinates},
		"дополнение без объекта": {func(in *NewSubmission) { in.Kind = KindAddition }, ErrObjectSlugRequired},
	}
	for name, c := range cases {
		in := newObject()
		c.mutate(&in)
		if err := in.Normalize(); !errors.Is(err, c.want) {
			t.Errorf("%s: ожидалась %v, получено %v", name, c.want, err)
		}
	}

	// Достаточно адреса без координат; пробелы обрезаются.
	in := newObject()
	in.Coordinates = nil
	if err := in.Normalize(); err != nil || in.Title != "Дом купца Тестова" {
		t.Fatalf("адрес без координат: %+v, %v", in, err)
	}
	// Дополнению хватает объекта и источника; slug приводится к нижнему регистру.
	add := NewSubmission{Kind: KindAddition, ObjectSlug: " Birzha ", Source: "Книга"}
	if err := add.Normalize(); err != nil || add.ObjectSlug != "birzha" {
		t.Fatalf("дополнение: %+v, %v", add, err)
	}
}

func TestCanDecide(t *testing.T) {
	s := Submission{AuthorID: 1, Status: StatusPending}
	if err := s.CanDecide(2); err != nil {
		t.Fatal(err)
	}
	if err := s.CanDecide(1); !errors.Is(err, ErrOwnSubmission) {
		t.Fatalf("решение по своей заявке: %v", err)
	}
	for _, st := range []Status{StatusApproved, StatusRejected} {
		s.Status = st
		if err := s.CanDecide(2); !errors.Is(err, ErrAlreadyDecided) {
			t.Fatalf("повторное решение (%s): %v", st, err)
		}
	}
}

func TestCanAddMediaAndLimit(t *testing.T) {
	s := Submission{AuthorID: 1, Status: StatusPending}
	if err := s.CanAddMedia(1); err != nil {
		t.Fatal(err)
	}
	if err := s.CanAddMedia(2); !errors.Is(err, ErrNotFound) {
		t.Fatalf("чужая заявка: %v", err)
	}
	s.Status = StatusRejected
	if err := s.CanAddMedia(1); !errors.Is(err, ErrDecidedNoMedia) {
		t.Fatalf("после решения: %v", err)
	}
	if err := CheckFilesLimit(MaxFiles - 1); err != nil {
		t.Fatal(err)
	}
	if err := CheckFilesLimit(MaxFiles); !errors.Is(err, ErrTooManyFiles) {
		t.Fatalf("лимит файлов: %v", err)
	}
}

func TestVisibleTo(t *testing.T) {
	s := Submission{AuthorID: 1}
	if !s.VisibleTo(Viewer{UserID: 1}) || !s.VisibleTo(Viewer{UserID: 2, IsModerator: true}) || s.VisibleTo(Viewer{UserID: 2}) {
		t.Fatal("видимость заявки: автор и модератор — да, остальные — нет")
	}
}

func TestRejectInputNormalize(t *testing.T) {
	in := RejectInput{Reason: "   "}
	if err := in.Normalize(); !errors.Is(err, ErrReasonRequired) {
		t.Fatalf("пустая причина: %v", err)
	}
	in = RejectInput{Reason: " Дубликат "}
	if err := in.Normalize(); err != nil || in.Reason != "Дубликат" {
		t.Fatalf("причина: %+v, %v", in, err)
	}
}

func TestApproveInputPrefill(t *testing.T) {
	long := strings.Repeat("я", ShortDescLen+10)
	s := Submission{Title: "Дом", Address: "ул. 1", Coordinates: &Coordinates{Lat: 46, Lng: 48}, Period: "1890-е", Description: long}

	got, err := ApproveInput{Slug: "dom"}.Prefill(s)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Дом" || got.Address != "ул. 1" || *got.Coordinates != *s.Coordinates ||
		got.YearBuilt == nil || *got.YearBuilt != "1890-е" || got.FullHistory != long ||
		utf8.RuneCountInString(got.ShortDescription) != ShortDescLen+1 || got.IsPublished == nil || !*got.IsPublished {
		t.Fatalf("заполнение из заявки: %+v", got)
	}

	// Поля модератора не перезаписываются.
	hidden := false
	got, _ = ApproveInput{Title: "Особняк", Coordinates: &Coordinates{Lat: 1, Lng: 2}, IsPublished: &hidden}.Prefill(s)
	if got.Title != "Особняк" || got.Coordinates.Lat != 1 || *got.IsPublished {
		t.Fatalf("поля модератора: %+v", got)
	}

	// Без координат ни в заявке, ни у модератора — ошибка; нулевые координаты считаются незаданными.
	s.Coordinates = nil
	if _, err := (ApproveInput{Coordinates: &Coordinates{}}).Prefill(s); !errors.Is(err, ErrNoCoordinates) {
		t.Fatalf("без координат: %v", err)
	}
}
