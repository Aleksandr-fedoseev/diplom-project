package auth

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"heritage-backend/internal/httpx"
)

// Principal — кто делает запрос. Кладётся в контекст запроса middleware'ом.
type Principal struct {
	UserID int64
	Role   string
}

type ctxKey struct{}

// FromContext возвращает текущего пользователя; ok=false для гостя.
func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}

// Authenticate читает заголовок «Authorization: Bearer <token>».
// Без токена запрос проходит как гостевой; с неверным токеном — отклоняется.
func Authenticate(tokens *TokenIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				next.ServeHTTP(w, r)
				return
			}
			raw, found := strings.CutPrefix(header, "Bearer ")
			if !found {
				httpx.WriteError(w, r, httpx.Unauthorized("ожидается заголовок Authorization: Bearer <token>"))
				return
			}
			claims, err := tokens.Parse(raw)
			if err != nil {
				httpx.WriteError(w, r, httpx.Unauthorized("токен недействителен или истёк"))
				return
			}
			id, err := strconv.ParseInt(claims.Subject, 10, 64)
			if err != nil {
				httpx.WriteError(w, r, httpx.Unauthorized("токен недействителен"))
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, Principal{UserID: id, Role: claims.Role})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AtLeast — роль не ниже minRole: модератор может всё, что пользователь, и т. д.
func AtLeast(role, minRole string) bool { return roleLevel[role] >= roleLevel[minRole] }

// RequireRole пропускает только пользователей с ролью не ниже minRole.
func RequireRole(minRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := FromContext(r.Context())
			if !ok {
				httpx.WriteError(w, r, httpx.Unauthorized("требуется вход"))
				return
			}
			if !AtLeast(p.Role, minRole) {
				httpx.WriteError(w, r, httpx.Forbidden("недостаточно прав"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
