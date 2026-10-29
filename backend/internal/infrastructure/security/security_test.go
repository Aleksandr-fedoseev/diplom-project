package security

import (
	"testing"
	"time"
)

func TestJWTRoundTripAndTampering(t *testing.T) {
	iss := NewJWTIssuer("0123456789abcdef0123456789abcdef", time.Hour)
	tok, err := iss.Issue(42, "moderator")
	if err != nil {
		t.Fatal(err)
	}
	id, role, err := iss.Parse(tok)
	if err != nil || id != 42 || role != "moderator" {
		t.Fatalf("разбор: %d %q %v", id, role, err)
	}
	if _, _, err := NewJWTIssuer("другой-секрет-0123456789abcdef0123", time.Hour).Parse(tok); err == nil {
		t.Fatal("токен с чужой подписью должен отклоняться")
	}
	expired, _ := NewJWTIssuer("0123456789abcdef0123456789abcdef", -time.Minute).Issue(1, "user")
	if _, _, err := iss.Parse(expired); err == nil {
		t.Fatal("просроченный токен должен отклоняться")
	}
}

func TestBcrypt(t *testing.T) {
	var b BcryptHasher
	h, err := b.Hash("password1")
	if err != nil || h == "password1" {
		t.Fatalf("хэш: %q %v", h, err)
	}
	if b.Compare(h, "password1") != nil || b.Compare(h, "wrong") == nil {
		t.Fatal("сравнение пароля с хэшем работает неверно")
	}
}
