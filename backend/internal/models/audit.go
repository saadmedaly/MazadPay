package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID         uuid.UUID `db:"id" json:"id"`
	AdminID    uuid.UUID `db:"admin_id" json:"admin_id"`
	Action     string    `db:"action" json:"action"`
	EntityType string    `db:"entity_type" json:"entity_type"`
	EntityID   *uuid.UUID `db:"entity_id" json:"entity_id"`
	Details    string    `db:"details" json:"details"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}