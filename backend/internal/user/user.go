package user

import (
	"time"
)

// User represents the persisted state of an application user.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
