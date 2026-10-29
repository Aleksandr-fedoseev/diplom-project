// Package auth — домен пользователей и ролей: сущности, правила и контракты модуля.
// Только стандартная библиотека: ни БД, ни HTTP, ни bcrypt здесь нет.
package auth

import (
	"net/mail"
	"strings"
	"time"

	"heritage-backend/internal/domain/errs"
)

// Role — роль пользователя. Каждая следующая включает права предыдущей:
// модератор может всё, что пользователь, администратор — всё, что модератор.
type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

// Уровни повторяют roles.level в БД.
var roleLevel = map[Role]int{RoleUser: 1, RoleModerator: 2, RoleAdmin: 3}

func (r Role) Valid() bool { return roleLevel[r] > 0 }

// AtLeast — роль не ниже min.
func (r Role) AtLeast(min Role) bool { return r.Valid() && roleLevel[r] >= roleLevel[min] }

type User struct {
	ID           int64
	Email        string
	Name         string
	Role         Role
	IsBlocked    bool
	CreatedAt    time.Time
	PasswordHash string // наружу не отдаётся: transport выводит только нужные поля
}

// Principal — кто делает запрос: достаётся из токена, без обращения к БД.
type Principal struct {
	UserID int64
	Role   Role
}

const MinPasswordLen = 8

// NormalizeEmail приводит email к одному виду, чтобы «Ivan@Test.ru» и «ivan@test.ru» были одним адресом.
func NormalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

// Registration — данные регистрации.
type Registration struct {
	Email    string
	Password string
	Name     string
}

// Normalize приводит поля к одному виду и проверяет правила регистрации.
func (r *Registration) Normalize() error {
	r.Email = NormalizeEmail(r.Email)
	r.Name = strings.TrimSpace(r.Name)
	switch {
	case !validEmail(r.Email):
		return errs.New(errs.Invalid, "некорректный email")
	case len(r.Password) < MinPasswordLen:
		return errs.New(errs.Invalid, "пароль должен быть не короче 8 символов")
	case r.Name == "":
		return errs.New(errs.Invalid, "укажите имя")
	}
	return nil
}

func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s // «Иван <a@b.ru>» адресом не считаем
}

// CheckRoleChange — правило смены роли: только допустимая роль и не себе
// (иначе единственный администратор мог бы лишить себя прав, и вернуть их было бы некому).
func CheckRoleChange(actor Principal, targetID int64, role Role) error {
	if !role.Valid() {
		return ErrInvalidRole
	}
	if actor.UserID == targetID {
		return ErrSelfRoleChange
	}
	return nil
}

// CheckBlock — правило блокировки: не себя; модератор блокирует только обычных пользователей,
// администратор — кого угодно.
func CheckBlock(actor Principal, target User) error {
	if actor.UserID == target.ID {
		return ErrSelfBlock
	}
	if actor.Role != RoleAdmin && target.Role != RoleUser {
		return ErrCannotBlockStaff
	}
	return nil
}
