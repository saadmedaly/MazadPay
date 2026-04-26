package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BidHistoryEntry représente une entrée dans l'historique des bids
// avec les informations de l'utilisateur incluse pour éviter les requêtes supplémentaires
type BidHistoryEntry struct {
	ID            uuid.UUID       `db:"id"              json:"id"`
	AuctionID     uuid.UUID       `db:"auction_id"      json:"auction_id"`
	UserID        uuid.UUID       `db:"user_id"         json:"user_id"`
	Amount        decimal.Decimal `db:"amount"          json:"amount"`
	PreviousPrice *decimal.Decimal `db:"previous_price"  json:"previous_price,omitempty"`
	IsWinning     bool            `db:"is_winning"      json:"is_winning"`
	CreatedAt     time.Time       `db:"created_at"      json:"created_at"`
	// Infos utilisateur (via JOIN)
	BidderName  string `db:"bidder_name"  json:"bidder_name,omitempty"`
	BidderPhone string `db:"bidder_phone" json:"bidder_phone,omitempty"`
	IsAnonymous bool   `db:"is_anonymous" json:"is_anonymous"`
}
