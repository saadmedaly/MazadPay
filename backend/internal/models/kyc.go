package models

import (
	"time"

	"github.com/google/uuid"
)

type KYCVerification struct {
	UserID         uuid.UUID  `db:"user_id"           json:"user_id"`
	IDCardFrontURL *string    `db:"id_card_front_url" json:"id_card_front_url"`
	IDCardBackURL  *string    `db:"id_card_back_url"  json:"id_card_back_url"`
	NNINumber      *string    `db:"nni_number"        json:"nni_number"`
	Status         string     `db:"status"            json:"status"` // pending, approved, rejected
	AdminNotes     *string    `db:"admin_notes"       json:"admin_notes"`
	ReviewedBy     *uuid.UUID `db:"reviewed_by"       json:"reviewed_by"`
	ReviewedAt     *time.Time `db:"reviewed_at"       json:"reviewed_at"`
	ExpiresAt      *time.Time `db:"expires_at"        json:"expires_at"`
	CreatedAt      time.Time  `db:"created_at"        json:"created_at"`
}
