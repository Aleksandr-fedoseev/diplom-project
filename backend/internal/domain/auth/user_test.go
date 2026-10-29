package auth

import (
	"errors"
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
