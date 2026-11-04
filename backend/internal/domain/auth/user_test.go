package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestRoleAtLeast(t *testing.T) {
	cases := []struct {
		role, min Role
		want      bool
	}{
		{RoleAdmin, RoleModerator, true},
		{RoleModerator, RoleModerator, true},
		{RoleUser, RoleModerator, false},
		{Role("hacker"), RoleUser, false},
	}
	for _, c := range cases {
		if got := c.role.AtLeast(c.min); got != c.want {
			t.Errorf("%s.AtLeast(%s) = %v", c.role, c.min, got)
		}
	}
}

func TestRegistrationNormalize(t *testing.T) {
	r := Registration{Email: "  Ivan@Test.RU ", Password: "password1", Name: " Иван "}
	if err := r.Normalize(); err != nil || r.Email != "ivan@test.ru" || r.Name != "Иван" {
		t.Fatalf("нормализация: %+v, %v", r, err)
	}
	for _, bad := range []Registration{
		{Email: "не-email", Password: "password1", Name: "Иван"},
		{Email: "Иван <a@b.ru>", Password: "password1", Name: "Иван"},
		{Email: "a@b.ru", Password: "short", Name: "Иван"},
		{Email: "a@b.ru", Password: "password1", Name: "  "},
	} {
		if err := bad.Normalize(); err == nil {
			t.Errorf("ожидалась ошибка для %+v", bad)
		}
	}
}

func TestCheckBlock(t *testing.T) {
	mod := Principal{UserID: 1, Role: RoleModerator}
	admin := Principal{UserID: 2, Role: RoleAdmin}
	user := User{ID: 3, Role: RoleUser}
	otherMod := User{ID: 4, Role: RoleModerator}

	if err := CheckBlock(mod, user); err != nil {
		t.Fatal(err)
	}
	if err := CheckBlock(mod, otherMod); !errors.Is(err, ErrCannotBlockStaff) {
		t.Fatalf("модератор блокирует модератора: %v", err)
	}
	if err := CheckBlock(admin, otherMod); err != nil {
		t.Fatal(err)
	}
	if err := CheckBlock(mod, User{ID: 1, Role: RoleModerator}); !errors.Is(err, ErrSelfBlock) {
		t.Fatalf("блокировка себя: %v", err)
	}
}

func TestCheckRoleChange(t *testing.T) {
	admin := Principal{UserID: 1, Role: RoleAdmin}
	if err := CheckRoleChange(admin, 2, RoleModerator); err != nil {
		t.Fatal(err)
	}
	if err := CheckRoleChange(admin, 1, RoleUser); !errors.Is(err, ErrSelfRoleChange) {
		t.Fatalf("смена роли себе: %v", err)
	}
	if err := CheckRoleChange(admin, 2, Role("root")); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("неизвестная роль: %v", err)
	}
}

// Длина пароля считается в символах, а предел bcrypt — в байтах.
func TestPasswordLength(t *testing.T) {
	reg := func(pw string) error {
		r := Registration{Email: "a@test.ru", Name: "А", Password: pw}
		return r.Normalize()
	}
	if err := reg("пароль1"); err == nil { // 7 символов, но 13 байт
		t.Error("7 русских символов должны отклоняться, хотя это больше 8 байт")
	}
	if err := reg("парольку"); err != nil { // 8 символов
		t.Errorf("8 русских символов: %v", err)
	}
	if err := reg(strings.Repeat("я", 36)); err != nil { // 72 байта
		t.Errorf("36 русских символов (72 байта): %v", err)
	}
	if err := reg(strings.Repeat("я", 37)); err == nil { // 74 байта — bcrypt не примет
		t.Error("пароль длиннее 72 байт должен отклоняться понятной ошибкой")
	}
}
