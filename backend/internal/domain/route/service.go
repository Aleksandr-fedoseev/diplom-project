package route

import "context"

// Service — сценарии модуля. Реализация — service/route; transport зависит только от этого интерфейса.
type Service interface {
	// Build строит маршрут по slug'ам объектов в заданном порядке, ничего не сохраняя.
	Build(ctx context.Context, slugs []string) (Route, error)
	// Save строит и сохраняет маршрут. ownerID == nil — тематический маршрут администратора.
	Save(ctx context.Context, in SaveInput, ownerID *int64) (Route, error)
	ListThematic(ctx context.Context) ([]Route, error)
	ListByOwner(ctx context.Context, ownerID int64) ([]Route, error)
	// Get возвращает маршрут с линией и остановками. viewer == nil — гость.
	// Чужой маршрут пользователя — ErrNotFound (не показываем даже, что он есть).
	Get(ctx context.Context, slug string, viewer *Actor) (Route, error)
	Delete(ctx context.Context, id int64, actor Actor) error
}
