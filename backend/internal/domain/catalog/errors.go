package catalog

import "heritage-backend/internal/domain/errs"

// Ошибки модуля. Сравниваются через errors.Is; transport переводит их род в HTTP-код.
var (
	ErrObjectNotFound  = errs.New(errs.NotFound, "объект не найден")
	ErrSlugTaken       = errs.New(errs.Conflict, "объект с таким slug уже есть")
	ErrUnknownCategory = errs.New(errs.Invalid, "неизвестная категория")
	// Объект нельзя удалить, пока на него ссылается маршрут.
	ErrObjectInRoute = errs.New(errs.Conflict, "объект входит в маршрут — сначала уберите его из маршрутов")

	ErrCategoryNotFound  = errs.New(errs.NotFound, "категория не найдена")
	ErrCategorySlugTaken = errs.New(errs.Conflict, "категория с таким slug уже есть")
	ErrCategoryNotEmpty  = errs.New(errs.Conflict, "в категории есть объекты — сначала перенесите их в другую")
)
