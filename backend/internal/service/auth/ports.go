package auth

// Технические порты сервиса. Реализации — infrastructure/security;
// сервис о bcrypt и JWT не знает, а в тестах подставляются фейки.

// PasswordHasher хэширует пароли и сверяет пароль с хэшем.
type PasswordHasher interface {
	Hash(password string) (string, error)
	// Compare возвращает ошибку, если пароль не подходит к хэшу.
	Compare(hash, password string) error
}

// TokenIssuer выпускает и проверяет токены доступа. Работает с простыми типами,
// чтобы инфраструктура не зависела от домена.
type TokenIssuer interface {
	Issue(userID int64, role string) (string, error)
	Parse(token string) (userID int64, role string, err error)
}
