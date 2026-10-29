// Package errs — общий вид ошибок домена. Домен говорит, ЧТО случилось
// («не найдено», «конфликт»), а транспорт решает, КАК это показать (HTTP 404, 409).
// Поэтому домен и сервисы не знают про HTTP-коды.
package errs

import "errors"

// Kind — род ошибки. По нему transport/http выбирает код ответа.
type Kind int

const (
	Invalid      Kind = iota + 1 // неверные данные от клиента → 400
	Unauthorized                 // не выполнен вход или неверные учётные данные → 401
	Forbidden                    // вход выполнен, но прав нет → 403
	NotFound                     // такого нет (или его не положено видеть) → 404
	Conflict                     // противоречит текущему состоянию: дубликат, повторное решение → 409
	Unavailable                  // внешний сервис недоступен → 503
)

// Error — ошибка домена: род и сообщение, которое можно показать пользователю.
type Error struct {
	Kind    Kind
	Message string
}

func (e *Error) Error() string { return e.Message }

// New создаёт ошибку. Постоянные ошибки модуля объявляются переменными
// (var ErrNotFound = errs.New(errs.NotFound, "...")) и сравниваются через errors.Is.
func New(kind Kind, message string) *Error { return &Error{Kind: kind, Message: message} }

// As возвращает ошибку домена из цепочки err, если она там есть.
func As(err error) (*Error, bool) {
	var e *Error
	ok := errors.As(err, &e)
	return e, ok
}

// Is сообщает, что в цепочке err есть ошибка домена рода kind. Удобно в тестах.
func Is(err error, kind Kind) bool {
	e, ok := As(err)
	return ok && e.Kind == kind
}
