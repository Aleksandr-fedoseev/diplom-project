// Package media — домен медиафайлов: фото, панорамы, аудио и 3D-модели объектов и заявок.
// Здесь сущность файла, правила для каждого вида медиа и правило видимости.
// Только стандартная библиотека: ни S3, ни БД, ни HTTP здесь нет.
package media

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"heritage-backend/internal/domain/errs"
)

// Kind — вид медиа.
type Kind string

const (
	KindPhoto        Kind = "photo"         // современная фотография
	KindArchivePhoto Kind = "archive_photo" // архивный снимок
	KindAudio        Kind = "audio"         // аудиогид
	KindModel3D      Kind = "model3d"       // 3D-модель в формате GLB
	KindPanorama     Kind = "panorama"      // панорама: сферическая 360° или частичная
)

// Projection — проекция панорамы.
type Projection string

const (
	ProjectionFull360 Projection = "full_360" // полная сфера
	ProjectionPartial Projection = "partial"  // часть круга, угол обзора — HAngle
)

// Source — что это за изображение. ИИ-реконструкции (например, восстановленный по описаниям
// облик утраченного здания) сайт обязан подписывать «реконструкция ИИ».
type Source string

const (
	SourcePhoto            Source = "photo"
	SourceAIReconstruction Source = "ai_reconstruction"
)

func (s Source) Valid() bool { return s == SourcePhoto || s == SourceAIReconstruction }

type File struct {
	ID           int64
	ObjectID     *int64 // файл объекта — виден всем
	SubmissionID *int64 // файл заявки — виден автору и модераторам
	Kind         Kind
	StorageKey   string // ключ в хранилище; наружу не отдаётся
	OriginalName string
	ContentType  string
	SizeBytes    int64
	Caption      string
	YearTaken    string // год съёмки
	SourceNote   string // откуда материал: архив, книга, личный архив
	Source       Source
	Projection   *Projection // только у панорамы
	HAngle       *int        // угол обзора в градусах, только у частичной панорамы
	Author       string
	License      string
	UploadedBy   *int64
	CreatedAt    time.Time
}

// IsPublic — файл прикреплён к объекту и виден всем.
func (f File) IsPublic() bool { return f.ObjectID != nil }

// Viewer — кто запрашивает файл. Гость — нулевое значение (UserID 0).
type Viewer struct {
	UserID    int64
	Moderator bool // модератор или администратор
}

// CanView — правило видимости: файл объекта виден всем; файл заявки на модерации —
// только тому, кто его загрузил, и модераторам.
func (f File) CanView(v Viewer) bool {
	if f.IsPublic() || v.Moderator {
		return true
	}
	return v.UserID != 0 && f.UploadedBy != nil && *f.UploadedBy == v.UserID
}

// URL — адрес скачивания файла в API.
func (f File) URL() string { return fmt.Sprintf("/api/media/%d/file", f.ID) }

// Rule — какие файлы принимаются для вида медиа.
type Rule struct {
	MaxBytes   int64
	Allowed    []string // допустимые MIME-типы, определённые по содержимому файла
	DefaultExt string   // расширение для ключа в хранилище, если у имени файла его нет
}

// Allows — тип содержимого допустим для этого вида.
func (r Rule) Allows(contentType string) bool { return slices.Contains(r.Allowed, contentType) }

const MB = 1 << 20

var images = []string{"image/jpeg", "image/png", "image/webp"}

var rules = map[Kind]Rule{
	KindPhoto:        {MaxBytes: 15 * MB, Allowed: images, DefaultExt: ".jpg"},
	KindArchivePhoto: {MaxBytes: 15 * MB, Allowed: images, DefaultExt: ".jpg"},
	KindAudio: {MaxBytes: 40 * MB, DefaultExt: ".mp3",
		Allowed: []string{"audio/mpeg", "audio/ogg", "application/ogg", "audio/wave", "audio/mp4", "video/mp4"}},
	KindModel3D: {MaxBytes: 50 * MB, Allowed: []string{"model/gltf-binary"}, DefaultExt: ".glb"},
	// Панорамы большие (8–12 тыс. пикселей по ширине), а просмотрщики понимают JPEG и WebP.
	KindPanorama: {MaxBytes: 30 * MB, Allowed: []string{"image/jpeg", "image/webp"}, DefaultExt: ".jpg"},
}

// RuleFor возвращает правило для вида медиа; ok=false — вид неизвестен.
func RuleFor(k Kind) (Rule, bool) {
	r, ok := rules[k]
	return r, ok
}

// UploadInput — загружаемый файл. Заполняется ровно одно из ObjectID и SubmissionID.
// Права на заявку проверяет модуль заявок до вызова Upload.
type UploadInput struct {
	ObjectID     int64
	SubmissionID int64
	Kind         Kind
	Caption      string
	YearTaken    string
	SourceNote   string
	Source       Source // пусто — photo
	Projection   Projection
	HAngle       *int
	Author       string
	License      string
	OriginalName string
	Size         int64
	Body         io.Reader
	UploadedBy   int64
}

// Normalize приводит поля к одному виду и проверяет правила вида медиа
// (размер проверяется отдельно — см. CheckSize, тип содержимого — Rule.Allows).
func (in *UploadInput) Normalize() (Rule, error) {
	in.Caption = strings.TrimSpace(in.Caption)
	in.YearTaken = strings.TrimSpace(in.YearTaken)
	in.SourceNote = strings.TrimSpace(in.SourceNote)
	in.Author = strings.TrimSpace(in.Author)
	in.License = strings.TrimSpace(in.License)
	in.Projection = Projection(strings.TrimSpace(string(in.Projection)))
	if in.Source == "" {
		in.Source = SourcePhoto
	}

	rule, ok := RuleFor(in.Kind)
	if !ok {
		return Rule{}, ErrUnknownKind
	}
	if !in.Source.Valid() {
		return Rule{}, ErrInvalidSource
	}
	if err := checkPanorama(in.Kind, in.Projection, in.HAngle); err != nil {
		return Rule{}, err
	}
	return rule, nil
}

// checkPanorama — правила панорам: у панорамы обязательна проекция, у других видов её нет;
// угол обзора от 1 до 360 градусов указывается только для частичной панорамы (необязателен).
func checkPanorama(kind Kind, p Projection, hAngle *int) error {
	if kind != KindPanorama {
		if p != "" {
			return ErrProjectionNotAllowed
		}
		if hAngle != nil {
			return ErrInvalidHAngle
		}
		return nil
	}
	if p != ProjectionFull360 && p != ProjectionPartial {
		return ErrProjectionRequired
	}
	if hAngle != nil && (p != ProjectionPartial || *hAngle < 1 || *hAngle > 360) {
		return ErrInvalidHAngle
	}
	return nil
}

// CheckSize — файл не больше предела вида и общего предела загрузки maxAll.
func (r Rule) CheckSize(size, maxAll int64) error {
	limit := min(r.MaxBytes, maxAll)
	if size > limit {
		return errs.New(errs.Invalid, fmt.Sprintf("файл больше допустимых %d МБ", limit/MB))
	}
	return nil
}
