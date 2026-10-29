package notification

import "context"

// Service — обработчики событий модуля. Реализация — service/notification.
//
// Handlers возвращает обработчик на каждое имя события. При сборке приложения
// их подписывают на outbox:
//
//	for eventType, h := range notifications.Handlers() {
//		processor.Subscribe(eventType, h)
//	}
//
// Обработчик получает payload в JSON. Ошибка — событие повторят позже. Событие может
// прийти повторно (доставка at-least-once), тогда письмо уйдёт ещё раз — это допустимо.
type Service interface {
	Handlers() map[string]func(ctx context.Context, payload []byte) error
}
