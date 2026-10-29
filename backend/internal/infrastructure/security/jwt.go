package security

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// claims — то, что хранится внутри JWT. Роль кладём в токен,
// чтобы не ходить в БД на каждый запрос ради проверки прав.
type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// JWTIssuer выпускает и проверяет токены, подписанные секретом (HS256).
type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTIssuer(secret string, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret), ttl: ttl}
}

func (t *JWTIssuer) Issue(userID int64, role string) (string, error) {
	now := time.Now()
	c := claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(t.secret)
}

// Parse проверяет подпись и срок действия. Алгоритм фиксирован: токен
// с другим алгоритмом (например, «none») отклоняется.
func (t *JWTIssuer) Parse(raw string) (int64, string, error) {
	c := &claims{}
	_, err := jwt.ParseWithClaims(raw, c, func(*jwt.Token) (any, error) {
		return t.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return 0, "", fmt.Errorf("разбор токена: %w", err)
	}
	id, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("в токене некорректный id: %w", err)
	}
	return id, c.Role, nil
}
