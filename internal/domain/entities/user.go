package entities

import (
	"strings"
	"time"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

var AllUserRoles = []UserRole{
	RoleUser,
	RoleAdmin,
}

type User struct {
	ID           string
	Name         string
	Email        string
	Role         UserRole
	PasswordHash string
	CreatedAt    time.Time
}

func NewUser(id, name, email, passwordHash string, role UserRole, now time.Time) (*User, error) {
	if role == "" {
		role = RoleUser
	}
	user := &User{
		ID:           strings.TrimSpace(id),
		Name:         strings.TrimSpace(name),
		Email:        strings.ToLower(strings.TrimSpace(email)),
		Role:         role,
		PasswordHash: passwordHash,
		CreatedAt:    now.UTC(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *User) Validate() error {
	if u.Name == "" || u.Email == "" || u.PasswordHash == "" {
		return ErrInvalidInput
	}
	if !IsValidRole(u.Role) {
		return ErrInvalidInput
	}

	return nil
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func IsValidRole(role UserRole) bool {
	for _, candidate := range AllUserRoles {
		if role == candidate {
			return true
		}
	}
	return false
}
