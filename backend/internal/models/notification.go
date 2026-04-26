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
	Priority      *string    `db:"priority"       json:"priority"`
	ActionURL     *string    `db:"action_url"     json:"action_url"`
	ActionLabel   *string    `db:"action_label"   json:"action_label"`
	CreatedAt     time.Time  `db:"created_at"     json:"created_at"`
}

type PushToken struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	UserID    uuid.UUID `db:"user_id"    json:"user_id"`
	FCMToken  string    `db:"fcm_token"  json:"fcm_token"`
	DeviceID  string    `db:"device_id"  json:"device_id"`
	Platform  string    `db:"platform"   json:"platform"` // android, ios, web
	IsActive  bool      `db:"is_active"  json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
