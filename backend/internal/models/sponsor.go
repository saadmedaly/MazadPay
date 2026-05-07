package models

import (
	"time"

	"github.com/google/uuid"
)

type Sponsor struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	Name      string    `db:"name"       json:"name"`
	Phone     string    `db:"phone"      json:"phone"`
	ImageURL  string    `db:"image_url"  json:"image_url"`
	LinkURL   *string   `db:"link_url"   json:"link_url"`
	IsActive  bool      `db:"is_active"  json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
