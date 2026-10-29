// Package catalog — домен каталога: объекты наследия, категории и правила их заполнения.
// Модуль — владелец объектов: другие модули получают сведения о них только через catalog.Service.
// Только стандартная библиотека: ни БД, ни HTTP здесь нет.
package catalog

import (
	"strings"
	"time"

	"heritage-backend/internal/domain/errs"
)

type Category struct {
	ID   int32
	Slug string
	Name string
}

// Coordinates — точка в градусах WGS 84.
type Coordinates struct {
	Lat float64
	Lng float64
}

// Valid — широта от −90 до 90, долгота от −180 до 180.
func (c Coordinates) Valid() bool {
	return c.Lat >= -90 && c.Lat <= 90 && c.Lng >= -180 && c.Lng <= 180
}

// ProtectionStatus — охранный статус объекта.
type ProtectionStatus string

const (
	StatusFederal    ProtectionStatus = "federal"    // объект федерального значения
	StatusRegional   ProtectionStatus = "regional"   // регионального значения
	StatusIdentified ProtectionStatus = "identified" // выявленный объект
	StatusLost       ProtectionStatus = "lost"       // утраченный
)

func (s ProtectionStatus) Valid() bool {
	switch s {
	case StatusFederal, StatusRegional, StatusIdentified, StatusLost:
		return true
	}
	return false
}

type Object struct {
	ID                 int64
	Slug               string
	Title              string
	Category           string // slug категории
	ProtectionStatus   ProtectionStatus
	RegistryNumber     *string // необязательные сведения: nil — неизвестно
	YearBuilt          *string
	Architect          *string
	ArchitecturalStyle *string
	Address            string
	Coordinates        Coordinates
	ShortDescription   string
	FullHistory        string // в списке не заполняется, есть только в карточке
	Tags               []string
	IsPublished        bool
	// DistanceM заполняется, только если искали объекты рядом с точкой.
	DistanceM *float64
	UpdatedAt time.Time
}

// Viewer — кто смотрит каталог. Транспорт заполняет его по токену; гость — нулевое значение.
type Viewer struct {
	Moderator bool // модератор или администратор
}

// VisibleTo — правило видимости: неопубликованный объект видят только модераторы и администраторы.
// Остальным он не показывается вовсе (как будто его нет).
func (o Object) VisibleTo(v Viewer) bool { return o.IsPublished || v.Moderator }

// Place — минимум сведений об объекте для других модулей (например, остановка маршрута).
type Place struct {
	ID    int64
	Slug  string
	Title string
	Lat   float64
	Lng   float64
}

// Filter — параметры списка объектов. Все поля необязательны и комбинируются через И.
type Filter struct {
	Category string
	Status   string
	Query    string // поиск по названию и адресу

	// Объекты в радиусе RadiusM метров от точки Near, ближайшие первыми.
	Near    *Coordinates
	RadiusM float64

	// Объекты в прямоугольнике видимой области карты: [minLng, minLat, maxLng, maxLat].
	BBox *[4]float64

	// IncludeUnpublished — просьба показать и неопубликованные объекты.
	// Выполняется только для модератора (см. ForViewer).
	IncludeUnpublished bool
}

// ForViewer применяет правило видимости к фильтру: неопубликованные объекты
// попадают в список, только если об этом попросили и смотрит модератор.
func (f Filter) ForViewer(v Viewer) Filter {
	f.IncludeUnpublished = f.IncludeUnpublished && v.Moderator
	return f
}

// ObjectInput — данные для создания и изменения объекта.
type ObjectInput struct {
	Slug               string
	Title              string
	Category           string
	ProtectionStatus   ProtectionStatus
	RegistryNumber     *string
	YearBuilt          *string
	Architect          *string
	ArchitecturalStyle *string
	Address            string
	Coordinates        Coordinates
	ShortDescription   string
	FullHistory        string
	Tags               []string
	IsPublished        *bool // nil — опубликовать сразу
}

// Normalize приводит поля к одному виду и проверяет правила заполнения объекта.
func (in *ObjectInput) Normalize() error {
	in.Slug = normalizeSlug(in.Slug)
	switch {
	case !validSlug(in.Slug):
		return errs.New(errs.Invalid, "slug обязателен и не должен содержать пробелы и / ? #")
	case strings.TrimSpace(in.Title) == "":
		return errs.New(errs.Invalid, "укажите название")
	case strings.TrimSpace(in.Address) == "":
		return errs.New(errs.Invalid, "укажите адрес (для утраченного объекта — исторический)")
	case !in.ProtectionStatus.Valid():
		return errs.New(errs.Invalid, "protectionStatus: federal, regional, identified или lost")
	case !in.Coordinates.Valid():
		return errs.New(errs.Invalid, "некорректные координаты")
	}
	if in.Tags == nil {
		in.Tags = []string{}
	}
	if in.IsPublished == nil {
		published := true
		in.IsPublished = &published
	}
	return nil
}

// CategoryInput — данные для создания и изменения категории.
type CategoryInput struct {
	Slug string
	Name string
}

// Normalize приводит поля к одному виду и проверяет правила заполнения категории.
func (in *CategoryInput) Normalize() error {
	in.Slug = normalizeSlug(in.Slug)
	in.Name = strings.TrimSpace(in.Name)
	if !validSlug(in.Slug) {
		return errs.New(errs.Invalid, "slug обязателен и не должен содержать пробелы и / ? #")
	}
	if in.Name == "" {
		return errs.New(errs.Invalid, "укажите название категории")
	}
	return nil
}

// slug — часть адреса страницы (/objects/birzha), поэтому без пробелов и служебных символов URL.
func normalizeSlug(s string) string { return strings.TrimSpace(strings.ToLower(s)) }

func validSlug(s string) bool { return s != "" && !strings.ContainsAny(s, " /?#") }
