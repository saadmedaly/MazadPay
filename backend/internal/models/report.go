package models

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID         uuid.UUID  `db:"id"          json:"id"`
	AuctionID  uuid.UUID  `db:"auction_id"  json:"auction_id"`
	ReporterID uuid.UUID  `db:"reporter_id" json:"reporter_id"`
	Reason     string     `db:"reason"      json:"reason"`
	Status     string     `db:"status"      json:"status"`
	ReviewedBy *uuid.UUID `db:"reviewed_by" json:"reviewed_by"`
	ReviewedAt *time.Time `db:"reviewed_at" json:"reviewed_at"`
	AdminNotes *string    `db:"admin_notes" json:"admin_notes"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
}
