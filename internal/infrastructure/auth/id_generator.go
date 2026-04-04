package auth

import "github.com/google/uuid"

type UUIDGenerator struct{}

func (UUIDGenerator) NewString() string {
	return uuid.NewString()
}
