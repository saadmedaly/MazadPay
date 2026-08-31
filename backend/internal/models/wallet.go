package models

import (
	"encoding/json"
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
	// CurrencyCode (migration 000046): the wallet's single, immutable
	// denomination for this V1 -- set once at wallet-creation time from the
	// owner's account market, never changed afterward (no multi-currency
	// wallets in this release). Nullable for wallets created before migration
	// 000046; EffectiveCurrencyCode below is the fallback-aware accessor.
	CurrencyCode *string `db:"currency_code" json:"currency_code,omitempty"`
}

// EffectiveCurrencyCode falls back to DefaultCurrencyCode for wallets
// predating migration 000046. Callers (bid/deposit/withdraw logic) must use
// this, never CurrencyCode directly.
func (w *Wallet) EffectiveCurrencyCode() string {
	if w.CurrencyCode != nil && *w.CurrencyCode != "" {
		return *w.CurrencyCode
	}
	return DefaultCurrencyCode
}

// MarshalJSON ensures currency_code is ALWAYS present when *Wallet is
// serialized directly (e.g. handlers.OK(c, wallet)), using EffectiveCurrencyCode
// for legacy pre-migration-000046 wallets instead of omitting the key.
func (w Wallet) MarshalJSON() ([]byte, error) {
	type alias Wallet
	return json.Marshal(struct {
		alias
		CurrencyCode string `json:"currency_code"`
	}{
		alias:        alias(w),
		CurrencyCode: w.EffectiveCurrencyCode(),
	})
}

type Transaction struct {
	ID               uuid.UUID        `db:"id"                 json:"id"`
	UserID           uuid.UUID        `db:"user_id"            json:"user_id"`
	AuctionID        *uuid.UUID       `db:"auction_id"         json:"auction_id"`
	Type             string           `db:"type"               json:"type"`
	Amount           decimal.Decimal  `db:"amount"             json:"amount"`
	Gateway          *string          `db:"gateway"            json:"gateway"`
	Status           string           `db:"status"             json:"status"`
	Reference        *string          `db:"reference"          json:"reference"`
	// ReceiptURL n'est jamais exposé directement en JSON (audit de sécurité) : le bucket
	// R2 est public, donc cette URL fonctionnerait sans authentification pour quiconque
	// la connaîtrait. Utiliser GET /wallet/transactions/:id/receipt-url pour obtenir une
	// URL présignée temporaire à la place.
	ReceiptURL       *string          `db:"receipt_url"        json:"-"`
	AdminNotes       *string          `db:"admin_notes"        json:"admin_notes"`
	ReviewedBy       *uuid.UUID       `db:"reviewed_by"        json:"reviewed_by"`
	ReviewedAt       *time.Time       `db:"reviewed_at"        json:"reviewed_at"`
	WalletHoldID     *uuid.UUID       `db:"wallet_hold_id"     json:"wallet_hold_id"`
	ReceiptImageTemp *string          `db:"receipt_image_temp" json:"receipt_image_temp"`
	PaymentMethod    *string          `db:"payment_method"     json:"payment_method"`
	FeeAmount        *decimal.Decimal `db:"fee_amount"         json:"fee_amount"`
	NetAmount        *decimal.Decimal `db:"net_amount"         json:"net_amount"`
	// UserFullName et UserPhone sont remplis via JOIN uniquement par GetByID (vue admin
	// détaillée) — permet à la web admin d'afficher le vrai nom de l'utilisateur au lieu
	// de son UUID tronqué. Absents (nil) pour les autres requêtes (ListPaginated,
	// FindByID) qui ne font pas ce JOIN.
	UserFullName     *string          `db:"user_full_name"     json:"user_full_name,omitempty"`
	UserPhone        *string          `db:"user_phone"         json:"user_phone,omitempty"`
	Description      *string          `db:"description"        json:"description"`
	FailureReason    *string          `db:"failure_reason"     json:"failure_reason"`
	CreatedAt        time.Time        `db:"created_at"         json:"created_at"`
	// CurrencyCode (migration 000046): stamped at transaction-creation time from
	// the wallet's currency, so this historical record remains correctly
	// denominated even if the user's account market changes later. Nullable for
	// transactions predating migration 000046 (fallback: DefaultCurrencyCode).
	CurrencyCode *string `db:"currency_code" json:"currency_code,omitempty"`
}

// EffectiveCurrencyCode falls back to DefaultCurrencyCode for transactions predating
// migration 000046. Callers displaying/exporting a transaction's currency must use this,
// never CurrencyCode directly.
func (t *Transaction) EffectiveCurrencyCode() string {
	if t.CurrencyCode != nil && *t.CurrencyCode != "" {
		return *t.CurrencyCode
	}
	return DefaultCurrencyCode
}

// MarshalJSON ensures currency_code is ALWAYS present when *Transaction is
// serialized directly (e.g. handlers.OK/PaginatedOK(c, tx(s))), using
// EffectiveCurrencyCode for legacy pre-migration-000046 transactions instead
// of omitting the key.
func (t Transaction) MarshalJSON() ([]byte, error) {
	type alias Transaction
	return json.Marshal(struct {
		alias
		CurrencyCode string `json:"currency_code"`
	}{
		alias:        alias(t),
		CurrencyCode: t.EffectiveCurrencyCode(),
	})
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
