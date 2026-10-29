// Package route — домен маршрутов: тематические маршруты администратора и сохранённые
// маршруты пользователей. Сущности, правила и контракты модуля; только стандартная библиотека.
// Пакет называется route (в единственном числе), чтобы не путать с HTTP-маршрутами.
package route

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Сколько объектов может быть в маршруте.
const (
	MinStops = 2
	MaxStops = 25
)

// Stop — остановка маршрута: объект каталога и его место в порядке посещения (с 1).
type Stop struct {
	Position int
	ObjectID int64
	Slug     string
	Title    string
	Lat      float64
	Lng      float64
}

// Route — маршрут. У построенного, но не сохранённого маршрута (Build) ID, Slug и Title пустые.
type Route struct {
	ID          int64
	Slug        string
	Title       string
	Description string
	OwnerID     *int64 // nil — тематический маршрут администратора, иначе — маршрут пользователя
	DistanceM   float64
	DurationS   float64
	Geometry    json.RawMessage // GeoJSON LineString; в списках не заполняется — линии тяжёлые
	Stops       []Stop
	CreatedAt   time.Time
}

// StopRef — точка маршрута, как она хранится в БД: только ссылка на объект.
// Название и координаты объекта сервис получает у каталога.
type StopRef struct {
	RouteID  int64
	Position int
	ObjectID int64
}

// Actor — кто смотрит или удаляет маршрут. Свой тип, чтобы не зависеть от домена auth.
type Actor struct {
	UserID  int64
	IsAdmin bool
}

// SaveInput — данные для сохранения маршрута.
type SaveInput struct {
	Slug        string
	Title       string
	Description string
	Objects     []string // slug'и объектов в порядке посещения
}

// CheckStopCount — в маршруте от MinStops до MaxStops объектов.
func CheckStopCount(n int) error {
	if n < MinStops || n > MaxStops {
		return ErrStopCount
	}
	return nil
}

// Normalize приводит поля к одному виду и проверяет правила сохранения.
// ownerID == nil — тематический маршрут: slug задаёт администратор.
// У маршрута пользователя slug служебный и создаётся здесь: придумывать его пользователю незачем.
func (in *SaveInput) Normalize(ownerID *int64) error {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return ErrTitleRequired
	}
	if ownerID != nil {
		in.Slug = NewUserSlug(*ownerID)
		return nil
	}
	in.Slug = strings.ToLower(strings.TrimSpace(in.Slug))
	if in.Slug == "" {
		return ErrSlugRequired
	}
	return nil
}

// NewUserSlug — служебный slug маршрута пользователя вида «u<id владельца>-<8 случайных символов>».
func NewUserSlug(ownerID int64) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b) // crypto/rand.Read не возвращает ошибок
	return fmt.Sprintf("u%d-%s", ownerID, hex.EncodeToString(b))
}

// CanView — тематический маршрут виден всем, маршрут пользователя — владельцу и администратору.
// viewer == nil — гость.
func (r Route) CanView(viewer *Actor) bool {
	if r.OwnerID == nil {
		return true
	}
	return viewer != nil && (viewer.IsAdmin || viewer.UserID == *r.OwnerID)
}

// CanDelete — свой маршрут удаляет владелец, любой (в том числе тематический) — администратор.
func CanDelete(ownerID *int64, actor Actor) bool {
	return actor.IsAdmin || (ownerID != nil && *ownerID == actor.UserID)
}
