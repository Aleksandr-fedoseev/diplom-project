package auth

import "heritage-backend/internal/domain/errs"

// Ошибки модуля. Сравниваются через errors.Is; transport переводит их род в HTTP-код.
var (
	ErrUserNotFound = errs.New(errs.NotFound, "пользователь не найден")
	ErrEmailTaken   = errs.New(errs.Conflict, "пользователь с таким email уже есть")
	// На неверный email и неверный пароль отвечаем одинаково,
	// чтобы по ответу нельзя было узнать, зарегистрирован ли email.
	ErrInvalidCredentials = errs.New(errs.Unauthorized, "неверный email или пароль")
	ErrInvalidToken       = errs.New(errs.Unauthorized, "токен недействителен или истёк")
	ErrBlocked            = errs.New(errs.Forbidden, "пользователь заблокирован")
	ErrInvalidRole        = errs.New(errs.Invalid, "role: user, moderator или admin")
	ErrSelfRoleChange     = errs.New(errs.Invalid, "нельзя менять роль самому себе")
	ErrSelfBlock          = errs.New(errs.Invalid, "нельзя заблокировать самого себя")
	ErrCannotBlockStaff   = errs.New(errs.Forbidden, "модератор может блокировать только пользователей")
)
