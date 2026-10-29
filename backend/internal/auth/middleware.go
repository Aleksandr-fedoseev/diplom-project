package auth

import (
	"context"
	"net/http"

	domainauth "heritage-backend/internal/domain/auth"
	"heritage-backend/internal/transport/http/middleware"
)

// Переходник для модулей, которые ещё не переехали на новую раскладку:
// они продолжают звать auth.FromContext и auth.RequireRole, а работу делает
// transport/http/middleware. Пакет удаляется, когда переедут submissions и shop.

// Principal — кто делает запрос.
type Principal struct {
	UserID int64
	Role   string
}

// FromContext возвращает текущего пользователя; ok=false для гостя.
func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := middleware.PrincipalFrom(ctx)
	return Principal{UserID: p.UserID, Role: string(p.Role)}, ok
}

// RequireRole пропускает только пользователей с ролью не ниже minRole.
func RequireRole(minRole string) func(http.Handler) http.Handler {
	return middleware.RequireRole(domainauth.Role(minRole))
}

// AtLeast — роль не ниже minRole.
func AtLeast(role, minRole string) bool {
	return domainauth.Role(role).AtLeast(domainauth.Role(minRole))
}
