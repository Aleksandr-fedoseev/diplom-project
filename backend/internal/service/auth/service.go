// Package auth — сценарии модуля пользователей: регистрация, вход, роли, блокировка.
// Реализует auth.Service из домена.
package auth

import (
	"context"
	"errors"
	"fmt"

	"heritage-backend/internal/domain/auth"
)

type Service struct {
	users  auth.UserRepository
	hasher PasswordHasher
	tokens TokenIssuer
}

// Проверка при сборке: Service реализует интерфейс домена.
var _ auth.Service = (*Service)(nil)

func NewService(users auth.UserRepository, hasher PasswordHasher, tokens TokenIssuer) *Service {
	return &Service{users: users, hasher: hasher, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, in auth.Registration) (auth.User, error) {
	if err := in.Normalize(); err != nil {
		return auth.User{}, err
	}
	return s.create(ctx, in, auth.RoleUser)
}

func (s *Service) create(ctx context.Context, in auth.Registration, role auth.Role) (auth.User, error) {
	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return auth.User{}, fmt.Errorf("хэширование пароля: %w", err)
	}
	return s.users.Create(ctx, auth.User{Email: in.Email, Name: in.Name, Role: role, PasswordHash: hash})
}

func (s *Service) Login(ctx context.Context, email, password string) (string, auth.User, error) {
	u, err := s.users.ByEmail(ctx, auth.NormalizeEmail(email))
	if errors.Is(err, auth.ErrUserNotFound) {
		return "", auth.User{}, auth.ErrInvalidCredentials
	}
	if err != nil {
		return "", auth.User{}, err
	}
	if s.hasher.Compare(u.PasswordHash, password) != nil {
		return "", auth.User{}, auth.ErrInvalidCredentials
	}
	if u.IsBlocked {
		return "", auth.User{}, auth.ErrBlocked
	}
	token, err := s.tokens.Issue(u.ID, string(u.Role))
	if err != nil {
		return "", auth.User{}, fmt.Errorf("выпуск токена: %w", err)
	}
	return token, u, nil
}

func (s *Service) Authenticate(_ context.Context, token string) (auth.Principal, error) {
	id, role, err := s.tokens.Parse(token)
	if err != nil || !auth.Role(role).Valid() {
		return auth.Principal{}, auth.ErrInvalidToken
	}
	return auth.Principal{UserID: id, Role: auth.Role(role)}, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (auth.User, error) {
	return s.users.ByID(ctx, id)
}

func (s *Service) EnsureActive(ctx context.Context, id int64) error {
	u, err := s.users.ByID(ctx, id)
	if err != nil {
		return err
	}
	if u.IsBlocked {
		return auth.ErrBlocked
	}
	return nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]auth.User, error) {
	return s.users.List(ctx, limit, offset)
}

func (s *Service) SetRole(ctx context.Context, actor auth.Principal, userID int64, role auth.Role) (auth.User, error) {
	if err := auth.CheckRoleChange(actor, userID, role); err != nil {
		return auth.User{}, err
	}
	return s.users.SetRole(ctx, userID, role)
}

// SetBlocked вызывается и в транзакции модуля заявок (отказ с блокировкой автора):
// репозиторий берёт транзакцию из ctx, поэтому отдельный вариант метода не нужен.
func (s *Service) SetBlocked(ctx context.Context, actor auth.Principal, userID int64, blocked bool) (auth.User, error) {
	target, err := s.users.ByID(ctx, userID)
	if err != nil {
		return auth.User{}, err
	}
	if err := auth.CheckBlock(actor, target); err != nil {
		return auth.User{}, err
	}
	return s.users.SetBlocked(ctx, userID, blocked)
}

func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	if email == "" || password == "" {
		return nil
	}
	in := auth.Registration{Email: email, Password: password, Name: "Администратор"}
	if err := in.Normalize(); err != nil {
		return fmt.Errorf("ADMIN_EMAIL / ADMIN_PASSWORD: %w", err)
	}
	_, err := s.create(ctx, in, auth.RoleAdmin)
	if errors.Is(err, auth.ErrEmailTaken) {
		return nil // уже существует
	}
	return err
}
