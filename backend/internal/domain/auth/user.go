// Package auth — домен пользователей и ролей: сущности, правила и контракты модуля.
// Только стандартная библиотека: ни БД, ни HTTP, ни bcrypt здесь нет.
package auth

import (
	"fmt"
	"net/mail"
	"path"
	"strings"
	"time"
	"unicode/utf8"

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
	PasswordHash string  // наружу не отдаётся: transport выводит только нужные поля
	AvatarKey    *string // ключ файла аватара в хранилище; nil — аватара нет
}

// AvatarURL — адрес аватара в API; пусто, если аватара нет. Параметр v меняется вместе
// с файлом, поэтому браузер не покажет старый аватар из кэша.
func (u User) AvatarURL() string {
	if u.AvatarKey == nil {
		return ""
	}
	return fmt.Sprintf("/api/users/%d/avatar?v=%s", u.ID, path.Base(*u.AvatarKey))
}

// Principal — кто делает запрос: достаётся из токена, без обращения к БД.
type Principal struct {
	UserID int64
	Role   Role
}

const (
	MinPasswordLen = 8 // символов, а не байт: «пароль» по-русски — 6 символов, но 12 байт
	// bcrypt учитывает только первые 72 байта и отказывается хешировать пароль длиннее.
	MaxPasswordBytes = 72
)

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
	if !validEmail(r.Email) {
		return errs.New(errs.Invalid, "некорректный email")
	}
	if err := CheckPassword(r.Password); err != nil {
		return err
	}
	name, err := NormalizeName(r.Name)
	r.Name = name
	return err
}

const MaxNameLen = 100 // символов

// NormalizeName — имя без пробелов по краям, не пустое и не длиннее MaxNameLen.
func NormalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	switch {
	case name == "":
		return "", errs.New(errs.Invalid, "укажите имя")
	case utf8.RuneCountInString(name) > MaxNameLen:
		return "", errs.New(errs.Invalid, "имя не длиннее 100 символов")
	}
	return name, nil
}

// CheckPassword — правило пароля при регистрации и смене.
func CheckPassword(p string) error {
	switch {
	case utf8.RuneCountInString(p) < MinPasswordLen:
		return errs.New(errs.Invalid, "пароль должен быть не короче 8 символов")
	case len(p) > MaxPasswordBytes:
		return errs.New(errs.Invalid, "пароль слишком длинный: не больше 72 латинских или 36 русских букв")
	}
	return nil
}

// Аватар: небольшая картинка; формат определяется по содержимому файла, а не по имени.
const MaxAvatarBytes = 2 << 20 // 2 МБ

var avatarTypes = map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}

// AvatarExt — расширение файла для типа содержимого; ошибка, если тип не картинка.
func AvatarExt(contentType string) (string, error) {
	ext, ok := avatarTypes[contentType]
	if !ok {
		return "", ErrAvatarType
	}
	return ext, nil
}

// AvatarContentType — тип содержимого по ключу файла (для отдачи аватара).
func AvatarContentType(key string) string {
	for t, ext := range avatarTypes {
		if strings.HasSuffix(key, ext) {
			return t
		}
	}
	return "application/octet-stream"
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
