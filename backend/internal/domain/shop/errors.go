package shop

import (
	"fmt"

	"heritage-backend/internal/domain/errs"
)

// Ошибки модуля. Сравниваются через errors.Is; transport переводит их род в HTTP-код.
var (
	ErrProductNotFound = errs.New(errs.NotFound, "товар не найден")
	ErrSlugTaken       = errs.New(errs.Conflict, "товар с таким slug уже есть")
	ErrOrderNotFound   = errs.New(errs.NotFound, "заказ не найден")

	ErrInvalidSlug     = errs.New(errs.Invalid, "slug обязателен и не должен содержать пробелы и / ? #")
	ErrTitleRequired   = errs.New(errs.Invalid, "укажите название")
	ErrInvalidPrice    = errs.New(errs.Invalid, "priceKopecks — цена в копейках, больше нуля")
	ErrNegativeStock   = errs.New(errs.Invalid, "stock не может быть отрицательным")
	ErrInvalidQuantity = errs.New(errs.Invalid, "quantity не может быть отрицательным")

	ErrInvalidAddress = errs.New(errs.Invalid, "укажите адрес доставки (до 500 символов)")
	ErrInvalidPhone   = errs.New(errs.Invalid, "некорректный номер телефона")
	ErrCartEmpty      = errs.New(errs.Invalid, "корзина пуста")

	ErrInvalidStatus     = errs.New(errs.Invalid, "status: pending, paid, shipped, completed или cancelled")
	ErrOrderNotPending   = errs.New(errs.Conflict, "заказ уже оплачен или отменён")
	ErrCannotCancel      = errs.New(errs.Conflict, "отменить можно только неоплаченный заказ")
	ErrStatusChangedRace = errs.New(errs.Conflict, "статус заказа только что изменился, обновите страницу")

	// Оплата. Раньше «платёжный сервис недоступен» давал 502; в errs нет отдельного рода
	// для 502, поэтому теперь обе ошибки — «внешний сервис недоступен» (503).
	ErrPaymentsDisabled = errs.New(errs.Unavailable, "приём оплаты временно не работает")
	ErrPaymentFailed    = errs.New(errs.Unavailable, "платёжный сервис недоступен, попробуйте позже")
)

// Ошибки с подробностями (название товара, остаток) создаются функциями.

func errNotEnoughForCart(title string, stock int) error {
	return errs.New(errs.Conflict, fmt.Sprintf("в наличии только %d шт. товара «%s»", stock, title))
}

func errProductInactive(title string) error {
	return errs.New(errs.Conflict, fmt.Sprintf("товар «%s» снят с продажи — уберите его из корзины", title))
}

func errNotEnoughForOrder(title string, stock int) error {
	return errs.New(errs.Conflict, fmt.Sprintf("товара «%s» осталось %d шт.", title, stock))
}

func errTransitionNotAllowed(from, to Status) error {
	return errs.New(errs.Conflict, fmt.Sprintf("нельзя перевести заказ из статуса %s в %s", from, to))
}
