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

type Service struct {
	db     *pgxpool.Pool
	tokens *TokenIssuer
}

func NewService(db *pgxpool.Pool, tokens *TokenIssuer) *Service {
	return &Service{db: db, tokens: tokens}
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

	var u User
	err = s.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, name, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, name, role, is_blocked, created_at`,
		in.Email, string(hash), in.Name, role,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.IsBlocked, &u.CreatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return User{}, httpx.Conflict("пользователь с таким email уже есть")
	}
	if err != nil {
		return User{}, fmt.Errorf("создание пользователя: %w", err)
	}
	return u, nil
}

// Login возвращает JWT. На неверный email и неверный пароль отвечаем одинаково,
// чтобы по ответу нельзя было узнать, зарегистрирован ли email.
func (s *Service) Login(ctx context.Context, email, password string) (string, User, error) {
	var (
		u    User
		hash string
	)
	err := s.db.QueryRow(ctx, `
		SELECT id, email, name, role, is_blocked, created_at, password_hash
		FROM users WHERE email = $1`,
		strings.ToLower(strings.TrimSpace(email)),
	).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.IsBlocked, &u.CreatedAt, &hash)

	invalid := httpx.Unauthorized("неверный email или пароль")
	if errors.Is(err, pgx.ErrNoRows) {
		return "", User{}, invalid
	}
	if err != nil {
		return "", User{}, fmt.Errorf("поиск пользователя: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", User{}, invalid
	}
	if u.IsBlocked {
		return "", User{}, httpx.Forbidden("пользователь заблокирован")
	}

	token, err := s.tokens.Issue(u.ID, u.Role)
	if err != nil {
		return "", User{}, fmt.Errorf("выпуск токена: %w", err)
	}
	return token, u, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, name, role, is_blocked, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.IsBlocked, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, httpx.NotFound("пользователь не найден")
	}
	return u, err
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
