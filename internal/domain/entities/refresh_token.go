package entities

import "time"

type RefreshToken struct {
	ID            string
	FamilyID      string
	UserID        string
	TokenHash     string
	DeviceName    string
	UserAgent     string
	IPAddress     string
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	ReplacedByID  *string
	CompromisedAt *time.Time
	CreatedAt     time.Time
}

func NewRefreshToken(id, familyID, userID, tokenHash, deviceName, userAgent, ipAddress string, expiresAt, now time.Time) (*RefreshToken, error) {
	token := &RefreshToken{
		ID:         id,
		FamilyID:   familyID,
		UserID:     userID,
		TokenHash:  tokenHash,
		DeviceName: deviceName,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
		ExpiresAt:  expiresAt.UTC(),
		CreatedAt:  now.UTC(),
	}
	if err := token.Validate(); err != nil {
		return nil, err
	}
	return token, nil
}

func (t *RefreshToken) Validate() error {
	if t.ID == "" || t.FamilyID == "" || t.UserID == "" || t.TokenHash == "" {
		return ErrInvalidInput
	}
	if !t.ExpiresAt.After(t.CreatedAt) {
		return ErrInvalidInput
	}
	return nil
}

func (t *RefreshToken) IsExpired(now time.Time) bool {
	return !t.ExpiresAt.After(now.UTC())
}

func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t *RefreshToken) Revoke(now time.Time) {
	ts := now.UTC()
	t.RevokedAt = &ts
}

func (t *RefreshToken) Rotate(nextID string, now time.Time) {
	t.Revoke(now)
	t.ReplacedByID = &nextID
}

func (t *RefreshToken) MarkCompromised(now time.Time) {
	ts := now.UTC()
	t.CompromisedAt = &ts
}

func (t *RefreshToken) WasRotated() bool {
	return t.ReplacedByID != nil
}
