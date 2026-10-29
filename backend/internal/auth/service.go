// Package auth — регистрация, вход, JWT и проверка ролей.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"heritage-backend/internal/auth/store"
	"heritage-backend/internal/httpx"
)

// Роли из таблицы roles. Уровень повторяет roles.level:
// модератор может всё, что пользователь, администратор — всё, что модератор.
const (
	RoleUser      = "user"
	RoleModerator = "moderator"
	RoleAdmin     = "admin"
)

var roleLevel = map[string]int{RoleUser: 1, RoleModerator: 2, RoleAdmin: 3}

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	IsBlocked bool      `json:"isBlocked"`
	CreatedAt time.Time `json:"createdAt"`
}

// toUser переводит строку таблицы в ответ API. Хэш пароля наружу не отдаётся.
func toUser(u store.User) User {
	return User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		IsBlocked: u.IsBlocked,
		CreatedAt: u.CreatedAt,
	}
}

type Service struct {
	q      *store.Queries // запросы из queries.sql, код сгенерирован sqlc
	tokens *TokenIssuer
}

func NewService(db *pgxpool.Pool, tokens *TokenIssuer) *Service {
	return &Service{q: store.New(db), tokens: tokens}
}

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (User, error) {
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Name = strings.TrimSpace(in.Name)
	if _, err := mail.ParseAddress(in.Email); err != nil {
		return User{}, httpx.BadRequest("некорректный email")
	}
	if len(in.Password) < 8 {
		return User{}, httpx.BadRequest("пароль должен быть не короче 8 символов")
	}
	if in.Name == "" {
		return User{}, httpx.BadRequest("укажите имя")
	}
	return s.create(ctx, in, RoleUser)
}

func (s *Service) create(ctx context.Context, in RegisterInput, role string) (User, error) {
	// bcrypt сам добавляет «соль», поэтому одинаковые пароли дают разные хэши.
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("хэширование пароля: %w", err)
	}

	u, err := s.q.CreateUser(ctx, store.CreateUserParams{
		Email:        in.Email,
		PasswordHash: string(hash),
		Name:         in.Name,
		Role:         role,
	})

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return User{}, httpx.Conflict("пользователь с таким email уже есть")
	}
	if err != nil {
		return User{}, fmt.Errorf("создание пользователя: %w", err)
	}
	return toUser(u), nil
}

// Login возвращает JWT. На неверный email и неверный пароль отвечаем одинаково,
// чтобы по ответу нельзя было узнать, зарегистрирован ли email.
func (s *Service) Login(ctx context.Context, email, password string) (string, User, error) {
	u, err := s.q.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))

	invalid := httpx.Unauthorized("неверный email или пароль")
	if errors.Is(err, pgx.ErrNoRows) {
		return "", User{}, invalid
	}
	if err != nil {
		return "", User{}, fmt.Errorf("поиск пользователя: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", User{}, invalid
	}
	if u.IsBlocked {
		return "", User{}, httpx.Forbidden("пользователь заблокирован")
	}

	token, err := s.tokens.Issue(u.ID, u.Role)
	if err != nil {
		return "", User{}, fmt.Errorf("выпуск токена: %w", err)
	}
	return token, toUser(u), nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (User, error) {
	u, err := s.q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, httpx.NotFound("пользователь не найден")
	}
	if err != nil {
		return User{}, fmt.Errorf("поиск пользователя: %w", err)
	}
	return toUser(u), nil
}

// EnsureAdmin создаёт первого администратора, если его ещё нет.
// Без этого некому было бы назначать модераторов.
func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	if email == "" || password == "" {
		return nil
	}
	_, err := s.create(ctx, RegisterInput{Email: strings.ToLower(email), Password: password, Name: "Администратор"}, RoleAdmin)
	var he *httpx.Error
	if errors.As(err, &he) && he.Status == 409 {
		return nil // уже существует
	}
	return err
}

// EnsureActive возвращает 403, если пользователь заблокирован. Токен заблокированного
// действует до истечения срока, поэтому модули проверяют блокировку перед записью
// (подача заявки, оформление заказа).
func (s *Service) EnsureActive(ctx context.Context, userID int64) error {
	u, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u.IsBlocked {
		return httpx.Forbidden("пользователь заблокирован")
	}
	return nil
}

// Управление пользователями (администратор и модератор).

func (s *Service) List(ctx context.Context, limit, offset int32) ([]User, error) {
	rows, err := s.q.ListUsers(ctx, store.ListUsersParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	users := make([]User, len(rows))
	for i, u := range rows {
		users[i] = toUser(u)
	}
	return users, nil
}

// SetRole назначает роль. Себе роль менять нельзя: иначе единственный администратор
// мог бы случайно лишить себя прав и некому было бы их вернуть.
func (s *Service) SetRole(ctx context.Context, actorID, userID int64, role string) (User, error) {
	if _, ok := roleLevel[role]; !ok {
		return User{}, httpx.BadRequest("role: user, moderator или admin")
	}
	if actorID == userID {
		return User{}, httpx.BadRequest("нельзя менять роль самому себе")
	}
	u, err := s.q.SetUserRole(ctx, store.SetUserRoleParams{ID: userID, Role: role})
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, httpx.NotFound("пользователь не найден")
	}
	if err != nil {
		return User{}, err
	}
	return toUser(u), nil
}

// SetBlocked блокирует или разблокирует пользователя. Модератор может блокировать
// только обычных пользователей, администратор — всех, кроме себя.
func (s *Service) SetBlocked(ctx context.Context, actorID int64, actorRole string, userID int64, blocked bool) (User, error) {
	return setBlocked(ctx, s.q, actorID, actorRole, userID, blocked)
}

// SetBlockedTx — то же в транзакции вызывающего модуля
// (отказ по заявке и блокировка её автора сохраняются вместе).
func (s *Service) SetBlockedTx(ctx context.Context, tx pgx.Tx, actorID int64, actorRole string, userID int64, blocked bool) (User, error) {
	return setBlocked(ctx, s.q.WithTx(tx), actorID, actorRole, userID, blocked)
}

func setBlocked(ctx context.Context, q *store.Queries, actorID int64, actorRole string, userID int64, blocked bool) (User, error) {
	if actorID == userID {
		return User{}, httpx.BadRequest("нельзя заблокировать самого себя")
	}
	target, err := q.GetUserByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, httpx.NotFound("пользователь не найден")
	}
	if err != nil {
		return User{}, err
	}
	if actorRole != RoleAdmin && target.Role != RoleUser {
		return User{}, httpx.Forbidden("модератор может блокировать только пользователей")
	}
	u, err := q.SetUserBlocked(ctx, store.SetUserBlockedParams{ID: userID, IsBlocked: blocked})
	if err != nil {
		return User{}, err
	}
	return toUser(u), nil
}
