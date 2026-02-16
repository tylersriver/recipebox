package entity

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser creates a new User with a generated UUID v7 and timestamps.
func NewUser(email, passwordHash string) (*User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		ID:           id.String(),
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
