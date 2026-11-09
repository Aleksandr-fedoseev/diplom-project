// Package auth — сценарии модуля пользователей: регистрация, вход, роли, блокировка.
// Реализует auth.Service из домена.
package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"

	"heritage-backend/internal/domain/auth"
)

type Service struct {
	users   auth.UserRepository
	hasher  PasswordHasher
	tokens  TokenIssuer
	avatars AvatarStorage
}

// Проверка при сборке: Service реализует интерфейс домена.
var _ auth.Service = (*Service)(nil)

func NewService(users auth.UserRepository, hasher PasswordHasher, tokens TokenIssuer, avatars AvatarStorage) *Service {
	return &Service{users: users, hasher: hasher, tokens: tokens, avatars: avatars}
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

// Authenticate: токен доказывает только, кто это. Роль и блокировку берём из БД,
// чтобы снятие роли или блокировка действовали сразу, а не после истечения токена.
func (s *Service) Authenticate(ctx context.Context, token string) (auth.Principal, error) {
	id, role, err := s.tokens.Parse(token)
	if err != nil || !auth.Role(role).Valid() {
		return auth.Principal{}, auth.ErrInvalidToken
	}
	u, err := s.users.ByID(ctx, id)
	if errors.Is(err, auth.ErrUserNotFound) {
		return auth.Principal{}, auth.ErrInvalidToken
	}
	if err != nil {
		return auth.Principal{}, err
	}
	if u.IsBlocked {
		return auth.Principal{}, auth.ErrBlocked
	}
	return auth.Principal{UserID: u.ID, Role: u.Role}, nil
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

func (s *Service) UpdateName(ctx context.Context, id int64, name string) (auth.User, error) {
	name, err := auth.NormalizeName(name)
	if err != nil {
		return auth.User{}, err
	}
	return s.users.UpdateName(ctx, id, name)
}

// ChangePassword требует текущий пароль: оставленный открытым чужой сеанс не должен
// позволять сменить пароль и забрать аккаунт.
func (s *Service) ChangePassword(ctx context.Context, id int64, current, next string) error {
	u, err := s.users.ByID(ctx, id)
	if err != nil {
		return err
	}
	if s.hasher.Compare(u.PasswordHash, current) != nil {
		return auth.ErrWrongPassword
	}
	if err := auth.CheckPassword(next); err != nil {
		return err
	}
	if current == next {
		return auth.ErrSamePassword
	}
	hash, err := s.hasher.Hash(next)
	if err != nil {
		return fmt.Errorf("хэширование пароля: %w", err)
	}
	return s.users.UpdatePassword(ctx, id, hash)
}

// SetAvatar сохраняет новый файл, затем ссылку на него, и только потом удаляет старый:
// при сбое посередине у пользователя остаётся рабочий аватар.
func (s *Service) SetAvatar(ctx context.Context, id int64, body io.Reader, size int64) (auth.User, error) {
	if size > auth.MaxAvatarBytes {
		return auth.User{}, auth.ErrAvatarTooBig
	}
	head := make([]byte, 512)
	n, err := io.ReadFull(body, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return auth.User{}, auth.ErrAvatarType
	}
	contentType := http.DetectContentType(head[:n])
	ext, err := auth.AvatarExt(contentType)
	if err != nil {
		return auth.User{}, err
	}
	old, err := s.users.ByID(ctx, id)
	if err != nil {
		return auth.User{}, err
	}

	key := fmt.Sprintf("avatars/%d/%s%s", id, randomHex(8), ext)
	if err := s.avatars.Put(ctx, key, io.MultiReader(bytes.NewReader(head[:n]), body), size, contentType); err != nil {
		return auth.User{}, fmt.Errorf("сохранение аватара: %w", err)
	}
	u, err := s.users.SetAvatar(ctx, id, &key)
	if err != nil {
		_ = s.avatars.Remove(ctx, key)
		return auth.User{}, err
	}
	if old.AvatarKey != nil {
		_ = s.avatars.Remove(ctx, *old.AvatarKey) // не удалился — останется лишний файл, не беда
	}
	return u, nil
}

func (s *Service) RemoveAvatar(ctx context.Context, id int64) (auth.User, error) {
	old, err := s.users.ByID(ctx, id)
	if err != nil {
		return auth.User{}, err
	}
	u, err := s.users.SetAvatar(ctx, id, nil)
	if err != nil {
		return auth.User{}, err
	}
	if old.AvatarKey != nil {
		_ = s.avatars.Remove(ctx, *old.AvatarKey)
	}
	return u, nil
}

func (s *Service) OpenAvatar(ctx context.Context, id int64) (io.ReadSeekCloser, string, error) {
	u, err := s.users.ByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if u.AvatarKey == nil {
		return nil, "", auth.ErrNoAvatar
	}
	body, err := s.avatars.Get(ctx, *u.AvatarKey)
	if err != nil {
		return nil, "", err
	}
	return body, auth.AvatarContentType(*u.AvatarKey), nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
