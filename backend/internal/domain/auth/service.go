package auth

import "context"

// Service — сценарии модуля. Реализация — service/auth; transport зависит только от этого интерфейса.
type Service interface {
	Register(ctx context.Context, in Registration) (User, error)
	// Login возвращает токен для заголовка Authorization: Bearer <token>.
	Login(ctx context.Context, email, password string) (token string, u User, err error)
	// Authenticate проверяет токен и возвращает, кто делает запрос.
	Authenticate(ctx context.Context, token string) (Principal, error)
	GetByID(ctx context.Context, id int64) (User, error)
	// EnsureActive — ErrBlocked, если пользователь заблокирован. Токен заблокированного действует
	// до истечения срока, поэтому модули проверяют блокировку перед записью.
	EnsureActive(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]User, error)
	SetRole(ctx context.Context, actor Principal, userID int64, role Role) (User, error)
	SetBlocked(ctx context.Context, actor Principal, userID int64, blocked bool) (User, error)
	// EnsureAdmin создаёт первого администратора, если такого email ещё нет.
	EnsureAdmin(ctx context.Context, email, password string) error
}
