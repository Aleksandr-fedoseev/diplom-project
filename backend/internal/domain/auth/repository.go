package auth

import "context"

// UserRepository — хранилище пользователей. Реализация — repository/postgres.
// Методы возвращают ошибки домена: ErrUserNotFound, если пользователя нет,
// ErrEmailTaken при повторном email.
type UserRepository interface {
	Create(ctx context.Context, u User) (User, error)
	ByID(ctx context.Context, id int64) (User, error)
	ByEmail(ctx context.Context, email string) (User, error)
	List(ctx context.Context, limit, offset int) ([]User, error)
	SetRole(ctx context.Context, id int64, role Role) (User, error)
	SetBlocked(ctx context.Context, id int64, blocked bool) (User, error)
	UpdateName(ctx context.Context, id int64, name string) (User, error)
	UpdatePassword(ctx context.Context, id int64, hash string) error
	// SetAvatar меняет ключ файла аватара; nil — убрать аватар.
	SetAvatar(ctx context.Context, id int64, key *string) (User, error)
}
