package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"password_hash"`
	FirstName    string    `db:"first_name" json:"first_name"`
	Surname      string    `db:"last_name" json:"last_name"`
	Role         string    `db:"role" json:"role"`
	CreatedAt    time.Time `db:"create_at" json:"create_at"`
	UpdatedAt    time.Time `db:"update_at" json:"update_at"`
}
