package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Wallet struct {
	UserID       uuid.UUID       `db:"user_id"       json:"user_id"`
	Balance      decimal.Decimal `db:"balance"       json:"balance"`
	FrozenAmount decimal.Decimal `db:"frozen_amount" json:"frozen_amount"`
	Version      int             `db:"version"       json:"version"`
	UpdatedAt    time.Time       `db:"updated_at"    json:"updated_at"`
}

type Transaction struct {
	ID           uuid.UUID        `db:"id"            json:"id"`
	UserID       uuid.UUID        `db:"user_id"       json:"user_id"`
	AuctionID    *uuid.UUID       `db:"auction_id"    json:"auction_id"`
	Type         string           `db:"type"          json:"type"`
	Amount       decimal.Decimal  `db:"amount"        json:"amount"`
	Gateway      *string          `db:"gateway"       json:"gateway"`
	Status       string           `db:"status"        json:"status"`
	Reference    *string          `db:"reference"     json:"reference"`
	ReceiptURL   *string          `db:"receipt_url"   json:"receipt_url"`
	AdminNotes   *string          `db:"admin_notes"   json:"admin_notes"`
	ReviewedBy   *uuid.UUID       `db:"reviewed_by"   json:"reviewed_by"`
	ReviewedAt   *time.Time       `db:"reviewed_at"   json:"reviewed_at"`
	WalletHoldID *uuid.UUID       `db:"wallet_hold_id" json:"wallet_hold_id"`
	CreatedAt    time.Time        `db:"created_at"    json:"created_at"`
}

type WalletHold struct {
	ID            uuid.UUID  `db:"id"             json:"id"`
	UserID        uuid.UUID  `db:"user_id"        json:"user_id"`
	AuctionID     uuid.UUID  `db:"auction_id"     json:"auction_id"`
	Amount        string     `db:"amount"         json:"amount"`
	Status        string     `db:"status"         json:"status"`
	TransactionID *uuid.UUID `db:"transaction_id" json:"transaction_id"`
	ReleasedAt    *time.Time `db:"released_at"    json:"released_at"`
	CreatedAt     time.Time  `db:"created_at"     json:"created_at"`
}
