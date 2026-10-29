package media

import (
	"errors"
	"testing"

	"heritage-backend/internal/domain/errs"
)

func ptr[T any](v T) *T { return &v }

func TestCanView(t *testing.T) {
	author := int64(7)
	public := File{ObjectID: ptr(int64(1))}
	private := File{SubmissionID: ptr(int64(2)), UploadedBy: &author}

	cases := []struct {
		name string
		f    File
		v    Viewer
		want bool
	}{
		{"гость — файл объекта", public, Viewer{}, true},
		{"гость — файл заявки", private, Viewer{}, false},
		{"автор — свой файл заявки", private, Viewer{UserID: 7}, true},
		{"чужой пользователь — файл заявки", private, Viewer{UserID: 8}, false},
		{"модератор — файл заявки", private, Viewer{UserID: 9, Moderator: true}, true},
		{"гость — файл без автора", File{SubmissionID: ptr(int64(2))}, Viewer{}, false},
	}
	for _, c := range cases {
		if got := c.f.CanView(c.v); got != c.want {
			t.Errorf("%s: CanView = %v", c.name, got)
		}
	}
}

func TestUploadInputNormalize(t *testing.T) {
	in := UploadInput{Kind: KindPhoto, Caption: "  Биржа  "}
	rule, err := in.Normalize()
	if err != nil || in.Source != SourcePhoto || in.Caption != "Биржа" || !rule.Allows("image/png") {
		t.Fatalf("фото: %+v, %v", in, err)
	}

	cases := []struct {
		name string
		in   UploadInput
		want error
	}{
		{"неизвестный вид", UploadInput{Kind: "video"}, ErrUnknownKind},
		{"неизвестный источник", UploadInput{Kind: KindPhoto, Source: "drawing"}, ErrInvalidSource},
		{"панорама без проекции", UploadInput{Kind: KindPanorama}, ErrProjectionRequired},
		{"панорама с неизвестной проекцией", UploadInput{Kind: KindPanorama, Projection: "cube"}, ErrProjectionRequired},
		{"проекция у фото", UploadInput{Kind: KindPhoto, Projection: ProjectionFull360}, ErrProjectionNotAllowed},
		{"угол у фото", UploadInput{Kind: KindPhoto, HAngle: ptr(90)}, ErrInvalidHAngle},
		{"угол у полной панорамы", UploadInput{Kind: KindPanorama, Projection: ProjectionFull360, HAngle: ptr(180)}, ErrInvalidHAngle},
		{"угол 0", UploadInput{Kind: KindPanorama, Projection: ProjectionPartial, HAngle: ptr(0)}, ErrInvalidHAngle},
		{"угол 361", UploadInput{Kind: KindPanorama, Projection: ProjectionPartial, HAngle: ptr(361)}, ErrInvalidHAngle},
	}
	for _, c := range cases {
		if _, err := c.in.Normalize(); !errors.Is(err, c.want) {
			t.Errorf("%s: ожидалось %v, получено %v", c.name, c.want, err)
		}
	}

	for _, ok := range []UploadInput{
		{Kind: KindPanorama, Projection: ProjectionFull360},
		{Kind: KindPanorama, Projection: ProjectionPartial},
		{Kind: KindPanorama, Projection: ProjectionPartial, HAngle: ptr(1)},
		{Kind: KindPanorama, Projection: ProjectionPartial, HAngle: ptr(360), Source: SourceAIReconstruction},
	} {
		if _, err := ok.Normalize(); err != nil {
			t.Errorf("%+v: %v", ok, err)
		}
	}
}

func TestPanoramaRule(t *testing.T) {
	rule, ok := RuleFor(KindPanorama)
	if !ok || !rule.Allows("image/jpeg") || !rule.Allows("image/webp") || rule.Allows("image/png") {
		t.Fatalf("панорама: JPEG и WebP, без PNG: %+v", rule)
	}
	if err := rule.CheckSize(30*MB, 100*MB); err != nil {
		t.Fatal(err)
	}
	if err := rule.CheckSize(30*MB+1, 100*MB); !errs.Is(err, errs.Invalid) {
		t.Fatalf("панорама больше 30 МБ: %v", err)
	}
	// Общий предел меньше предела вида — действует общий.
	if err := rule.CheckSize(20*MB, 10*MB); !errs.Is(err, errs.Invalid) || err.Error() != "файл больше допустимых 10 МБ" {
		t.Fatalf("общий предел: %v", err)
	}
}
