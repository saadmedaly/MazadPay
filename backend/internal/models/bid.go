package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Bid struct {
	ID            uuid.UUID        `db:"id"             json:"id"`
	AuctionID     uuid.UUID        `db:"auction_id"     json:"auction_id"`
	UserID        uuid.UUID        `db:"user_id"        json:"user_id"`
	Amount        decimal.Decimal  `db:"amount"         json:"amount"`
	PreviousPrice *decimal.Decimal `db:"previous_price" json:"previous_price"`
	IsWinning     bool             `db:"is_winning"     json:"is_winning"`
	BidderPhone   string           `db:"bidder_phone"   json:"bidder_phone,omitempty"` // Pour masquage ####xxxx
	CreatedAt     time.Time        `db:"created_at"     json:"created_at"`
}
