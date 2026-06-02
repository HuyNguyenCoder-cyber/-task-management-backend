package dto

import "time"

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type RegisterResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
	User        AuthUserInfo `json:"user"`
}

type AuthUserInfo struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}
