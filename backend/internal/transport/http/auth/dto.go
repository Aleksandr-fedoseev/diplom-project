package auth

import (
	"time"

	"heritage-backend/internal/domain/auth"
)

// DTO — формат JSON в API. Отделён от сущности домена, чтобы ни хэш пароля,
// ни будущие внутренние поля не утекли в ответ случайно.

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	IsBlocked bool      `json:"isBlocked"`
	AvatarURL string    `json:"avatarUrl"` // пусто — аватара нет
	CreatedAt time.Time `json:"createdAt"`
}

func toUserResponse(u auth.User) userResponse {
	return userResponse{
		ID: u.ID, Email: u.Email, Name: u.Name, Role: string(u.Role), IsBlocked: u.IsBlocked,
		AvatarURL: u.AvatarURL(), CreatedAt: u.CreatedAt,
	}
}

type profileRequest struct {
	Name string `json:"name"`
}

type passwordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type loginResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

type roleRequest struct {
	Role string `json:"role"`
}

type blockedRequest struct {
	Blocked bool `json:"blocked"`
}
