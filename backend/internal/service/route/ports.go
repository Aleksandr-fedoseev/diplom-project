package route

import (
	"context"
	"encoding/json"
)

// Порты сервиса: то, что ему нужно от других модулей и от техники.
// Типы здесь свои, чтобы сервис не зависел ни от каталога, ни от OSRM;
// переходники пишутся при сборке приложения, а в тестах подставляются фейки.

// Place — объект каталога, каким его видит модуль маршрутов.
type Place struct {
	ID    int64
	Slug  string
	Title string
	Lat   float64
	Lng   float64
}

// Places — то, что маршрутам нужно от каталога: названия и координаты объектов.
// Таблицу объектов модуль маршрутов сам не читает.
type Places interface {
	// PublishedPlaces — опубликованные объекты с такими slug (порядок любой; ненайденных нет в ответе).
	PublishedPlaces(ctx context.Context, slugs []string) ([]Place, error)
	// PlacesByIDs — объекты по id, в том числе снятые с публикации: остановка сохранённого маршрута не должна пропадать.
	PlacesByIDs(ctx context.Context, ids []int64) ([]Place, error)
}

// Point — точка для маршрутизатора.
type Point struct {
	Lat float64
	Lng float64
}

// RouteLine — то, что построил маршрутизатор.
type RouteLine struct {
	DistanceM float64         // длина в метрах
	DurationS float64         // время в пути в секундах
	Geometry  json.RawMessage // GeoJSON LineString
}

// Router строит пеший маршрут по улицам через точки в заданном порядке. Реализация — OSRM.
type Router interface {
	Route(ctx context.Context, points []Point) (RouteLine, error)
}

// TxManager выполняет fn в одной транзакции: ошибка — откат всего.
type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
