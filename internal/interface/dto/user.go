package dto

import (
	"time"

	"clean_architecture/internal/domain/entities"
)

type RegisterUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginUserRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	DeviceName string `json:"device_name"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RevokeSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

type SessionResponse struct {
	ID            string     `json:"id"`
	DeviceName    string     `json:"device_name"`
	UserAgent     string     `json:"user_agent"`
	IPAddress     string     `json:"ip_address"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     time.Time  `json:"expires_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	CompromisedAt *time.Time `json:"compromised_at,omitempty"`
}

func ToUserResponse(user *entities.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt,
	}
}

func ToSessionResponse(token *entities.RefreshToken) SessionResponse {
	return SessionResponse{
		ID:            token.ID,
		DeviceName:    token.DeviceName,
		UserAgent:     token.UserAgent,
		IPAddress:     token.IPAddress,
		CreatedAt:     token.CreatedAt,
		ExpiresAt:     token.ExpiresAt,
		RevokedAt:     token.RevokedAt,
		CompromisedAt: token.CompromisedAt,
	}
}
