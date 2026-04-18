package models

import (
	"time"

	"github.com/google/uuid"
)

type AdminInvitation struct {
	ID        uuid.UUID  `db:"id"         json:"id"`
	Token     string     `db:"token"      json:"token"`
	CreatedBy uuid.UUID  `db:"created_by" json:"created_by"`
	ExpiresAt time.Time  `db:"expires_at" json:"expires_at"`
	UsedAt    *time.Time `db:"used_at"    json:"used_at"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}
