// Package security — хэширование паролей (bcrypt) и токены доступа (JWT).
// О домене пакет не знает: работает со строками и числами.
package security

import "golang.org/x/crypto/bcrypt"

// BcryptHasher — хэширование паролей. bcrypt сам добавляет «соль»,
// поэтому одинаковые пароли дают разные хэши.
type BcryptHasher struct{}

func (BcryptHasher) Hash(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

func (BcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
