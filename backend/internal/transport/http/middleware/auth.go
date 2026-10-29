// Package middleware — общие обработчики-обёртки: проверка токена и ролей.
// Логи, CORS и защита от паники берутся из chi (см. router.go).
package middleware

import (
	"context"
	"net/http"
	"strings"

	"heritage-backend/internal/domain/auth"
	"heritage-backend/internal/transport/http/httpx"
)

// Authenticator — то, что middleware нужно от модуля auth: проверить токен.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (auth.Principal, error)
}

type ctxKey struct{}

// PrincipalFrom возвращает текущего пользователя; ok=false для гостя.
func PrincipalFrom(ctx context.Context) (auth.Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(auth.Principal)
	return p, ok
}

// WithPrincipal кладёт пользователя в контекст. Нужен middleware и тестам обработчиков.
func WithPrincipal(ctx context.Context, p auth.Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

// Authenticate читает заголовок «Authorization: Bearer <token>».
// Без токена запрос проходит как гостевой; с неверным токеном — отклоняется.
func Authenticate(a Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}
			raw, found := strings.CutPrefix(header, "Bearer ")
			if !found {
				httpx.WriteError(w, r, auth.ErrInvalidToken)
				return
			}
			p, err := a.Authenticate(r.Context(), raw)
			if err != nil {
				httpx.WriteError(w, r, err)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), p)))
		})
	}
}

var (
	errLoginRequired = httpx.Unauthorized("требуется вход")
	errNoRights      = httpx.Forbidden("недостаточно прав")
)

// RequireRole пропускает только пользователей с ролью не ниже min.
func RequireRole(min auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PrincipalFrom(r.Context())
			if !ok {
				httpx.WriteError(w, r, errLoginRequired)
				return
			}
			if !p.Role.AtLeast(min) {
				httpx.WriteError(w, r, errNoRights)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
