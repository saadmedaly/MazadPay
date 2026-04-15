package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID            uuid.UUID  `db:"id"             json:"id"`
	UserID        uuid.UUID  `db:"user_id"        json:"user_id"`
	Type          string     `db:"type"           json:"type"`
	Title         string     `db:"title"          json:"title"`
	Body          *string    `db:"body"           json:"body"`
	IsRead        bool       `db:"is_read"        json:"is_read"`
	ReferenceID   *uuid.UUID `db:"reference_id"   json:"reference_id"`
	ReferenceType *string    `db:"reference_type" json:"reference_type"`
	Data          JSONB      `db:"data"           json:"data"`
	CreatedAt     time.Time  `db:"created_at"     json:"created_at"`
}
