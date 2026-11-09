package auth

import (
	"context"
	"io"
)

// Service — сценарии модуля. Реализация — service/auth; transport зависит только от этого интерфейса.
type Service interface {
	Register(ctx context.Context, in Registration) (User, error)
	// Login возвращает токен для заголовка Authorization: Bearer <token>.
	Login(ctx context.Context, email, password string) (token string, u User, err error)
	// Authenticate проверяет токен и возвращает, кто делает запрос. Роль — из БД, а не из токена;
	// заблокированному — ErrBlocked, удалённому — ErrInvalidToken.
	Authenticate(ctx context.Context, token string) (Principal, error)
	GetByID(ctx context.Context, id int64) (User, error)
	// EnsureActive — ErrBlocked, если пользователь заблокирован. Нужен модулям, которые проверяют
	// пользователя внутри своей транзакции, а не только на входе запроса.
	EnsureActive(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]User, error)
	SetRole(ctx context.Context, actor Principal, userID int64, role Role) (User, error)
	SetBlocked(ctx context.Context, actor Principal, userID int64, blocked bool) (User, error)
	// Профиль: пользователь меняет своё имя, пароль и аватар.
	UpdateName(ctx context.Context, id int64, name string) (User, error)
	ChangePassword(ctx context.Context, id int64, current, next string) error
	SetAvatar(ctx context.Context, id int64, body io.Reader, size int64) (User, error)
	RemoveAvatar(ctx context.Context, id int64) (User, error)
	// OpenAvatar — поток файла аватара и его тип; ErrNoAvatar, если аватара нет.
	OpenAvatar(ctx context.Context, id int64) (io.ReadSeekCloser, string, error)
	// EnsureAdmin создаёт первого администратора, если такого email ещё нет.
	EnsureAdmin(ctx context.Context, email, password string) error
}
