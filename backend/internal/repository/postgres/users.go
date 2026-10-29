// Package postgres — реализации репозиториев домена на PostgreSQL.
// SQL лежит в queries/*.sql, Go-код для него генерирует sqlc в пакет sqlc.
// Модели sqlc отсюда не выходят: репозитории отдают сущности домена.
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/infrastructure/postgres"
	"heritage-backend/internal/repository/postgres/sqlc"
)

// Коды ошибок PostgreSQL, которые репозитории переводят в ошибки домена.
const (
	codeUniqueViolation     = "23505"
	codeForeignKeyViolation = "23503"
	codeNotNullViolation    = "23502"
)

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// queries — запросы sqlc на текущем соединении: транзакции из ctx (TxManager.Do) или пуле.
func queries(ctx context.Context, pool *pgxpool.Pool) *sqlc.Queries {
	return sqlc.New(postgres.Conn(ctx, pool))
}

type Users struct {
	pool *pgxpool.Pool
}

var _ auth.UserRepository = (*Users)(nil)

func NewUsers(pool *pgxpool.Pool) *Users { return &Users{pool: pool} }

func toUser(u sqlc.User) auth.User {
	return auth.User{
		ID:           u.ID,
		Email:        u.Email,
		Name:         u.Name,
		Role:         auth.Role(u.Role),
		IsBlocked:    u.IsBlocked,
		CreatedAt:    u.CreatedAt,
		PasswordHash: u.PasswordHash,
	}
}

// oneUser переводит результат запроса одной строки: нет строки — пользователь не найден.
func oneUser(u sqlc.User, err error) (auth.User, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.User{}, auth.ErrUserNotFound
	}
	if err != nil {
		return auth.User{}, err
	}
	return toUser(u), nil
}

func (r *Users) Create(ctx context.Context, u auth.User) (auth.User, error) {
	row, err := queries(ctx, r.pool).CreateUser(ctx, sqlc.CreateUserParams{
		Email: u.Email, PasswordHash: u.PasswordHash, Name: u.Name, Role: string(u.Role),
	})
	if pgCode(err) == codeUniqueViolation {
		return auth.User{}, auth.ErrEmailTaken
	}
	return oneUser(row, err)
}

func (r *Users) ByID(ctx context.Context, id int64) (auth.User, error) {
	return oneUser(queries(ctx, r.pool).GetUserByID(ctx, id))
}

func (r *Users) ByEmail(ctx context.Context, email string) (auth.User, error) {
	return oneUser(queries(ctx, r.pool).GetUserByEmail(ctx, email))
}

func (r *Users) List(ctx context.Context, limit, offset int) ([]auth.User, error) {
	rows, err := queries(ctx, r.pool).ListUsers(ctx, sqlc.ListUsersParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, err
	}
	users := make([]auth.User, len(rows))
	for i, u := range rows {
		users[i] = toUser(u)
	}
	return users, nil
}

func (r *Users) SetRole(ctx context.Context, id int64, role auth.Role) (auth.User, error) {
	return oneUser(queries(ctx, r.pool).SetUserRole(ctx, sqlc.SetUserRoleParams{ID: id, Role: string(role)}))
}

func (r *Users) SetBlocked(ctx context.Context, id int64, blocked bool) (auth.User, error) {
	return oneUser(queries(ctx, r.pool).SetUserBlocked(ctx, sqlc.SetUserBlockedParams{ID: id, IsBlocked: blocked}))
}
