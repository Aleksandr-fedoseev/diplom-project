package route

import (
	"fmt"

	"heritage-backend/internal/domain/errs"
)

// Ошибки модуля. Сравниваются через errors.Is; transport переводит их род в HTTP-код.
var (
	ErrNotFound      = errs.New(errs.NotFound, "маршрут не найден")
	ErrStopCount     = errs.New(errs.Invalid, fmt.Sprintf("в маршруте должно быть от %d до %d объектов", MinStops, MaxStops))
	ErrTitleRequired = errs.New(errs.Invalid, "укажите название маршрута")
	ErrSlugRequired  = errs.New(errs.Invalid, "укажите slug маршрута")
	ErrSlugTaken     = errs.New(errs.Conflict, "маршрут с таким slug уже есть")
	ErrDeleteDenied  = errs.New(errs.Forbidden, "удалить можно только свой маршрут")
	// Маршрутизатор (OSRM) не ответил или не построил маршрут.
	ErrRouterUnavailable = errs.New(errs.Unavailable, "сервис построения маршрутов недоступен")
)

// ObjectNotFound — в запросе slug объекта, которого нет в каталоге (или он не опубликован).
func ObjectNotFound(slug string) error {
	return errs.New(errs.Invalid, "объект не найден: "+slug)
}
