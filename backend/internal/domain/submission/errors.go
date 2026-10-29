package submission

import (
	"fmt"

	"heritage-backend/internal/domain/errs"
)

// Ошибки модуля. Сравниваются через errors.Is; transport переводит их род в HTTP-код.
var (
	// Чужую заявку не показываем: отвечаем «не найдена», чтобы по ответу нельзя было
	// узнать, что такая заявка есть.
	ErrNotFound = errs.New(errs.NotFound, "заявка не найдена")
	// О скрытых (неопубликованных) объектах тоже не сообщаем.
	ErrObjectNotFound = errs.New(errs.NotFound, "объект не найден")
	ErrObjectDeleted  = errs.New(errs.Conflict, "объект, к которому относится дополнение, удалён")

	ErrAlreadyDecided = errs.New(errs.Conflict, "по заявке уже принято решение")
	ErrOwnSubmission  = errs.New(errs.Forbidden, "нельзя принимать решение по своей заявке")
	ErrDecidedNoMedia = errs.New(errs.Conflict, "по заявке уже принято решение, добавлять файлы нельзя")
	ErrTooManyFiles   = errs.New(errs.Conflict, fmt.Sprintf("к заявке можно прикрепить не больше %d файлов", MaxFiles))

	ErrReasonRequired = errs.New(errs.Invalid, "укажите причину отказа")
	ErrNoCoordinates  = errs.New(errs.Invalid, "в заявке нет координат — укажите их для объекта (coordinates)")
	ErrInvalidStatus  = errs.New(errs.Invalid, "status: pending, approved или rejected")

	// Ошибки проверки новой заявки.
	ErrInvalidKind        = errs.New(errs.Invalid, "kind: new_object или addition")
	ErrInvalidCoordinates = errs.New(errs.Invalid, "некорректные координаты")
	ErrSourceRequired     = errs.New(errs.Invalid, "укажите источник сведений (архив, книга, личный архив и т. п.)")
	ErrTitleRequired      = errs.New(errs.Invalid, "укажите название объекта")
	ErrPlaceRequired      = errs.New(errs.Invalid, "укажите место: адрес или координаты")
	ErrPeriodRequired     = errs.New(errs.Invalid, "укажите период или дату")
	ErrObjectSlugRequired = errs.New(errs.Invalid, "укажите объект, к которому относится дополнение (objectSlug)")
)
