package notification

import "heritage-backend/internal/domain/errs"

// Ошибки модуля. Пользователю они не показываются (у модуля нет HTTP), но по роду
// ошибки видно, повторять ли попытку: битое событие повтор не исправит.
var (
	// ErrRecipientNotFound — получателя нет (удалён). Письмо не отправляется, это не ошибка доставки.
	// Адаптер пользователей может вернуть её или любую ошибку рода errs.NotFound.
	ErrRecipientNotFound = errs.New(errs.NotFound, "получатель письма не найден")
	ErrMalformedEvent    = errs.New(errs.Invalid, "некорректные данные события")
	ErrNoRecipientEmail  = errs.New(errs.Invalid, "у получателя нет адреса почты")
	ErrHeaderInjection   = errs.New(errs.Invalid, "перевод строки в адресе или теме письма")
)
