package models

import (
	"time"

	"github.com/google/uuid"
)

type AppRating struct {
	ID        uuid.UUID  `db:"id"             json:"id"`
	UserID    uuid.UUID  `db:"user_id"        json:"user_id"`
	AuctionID uuid.UUID  `db:"auction_id"     json:"auction_id"`
	Rating    int        `db:"rating"          json:"rating"`
	Comment   *string    `db:"comment"         json:"comment"`
	CreatedAt time.Time  `db:"created_at"      json:"created_at"`
}
