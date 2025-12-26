package dto

import (
	"lab/internal/app/role"
	"time"
)

type UserRegistration struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserDataResposne struct {
	ID    uint   `json:"id"`
	Login string `json:"login"`
}

type ChangeUserData struct {
	Login string `json:"login,omitempty"`
}

type LoginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRes struct {
	ExpiresIn   time.Duration `json:"expires_in"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
	IsModerator role.Role     `json:"is_moderator"`
}
