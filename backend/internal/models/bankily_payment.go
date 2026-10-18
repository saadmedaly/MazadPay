package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Bankily B-PAY attempt statuses (migration 000061).
const (
	BankilyStatusInitiated           = "INITIATED"
	BankilyStatusSubmitted           = "SUBMITTED"
	BankilyStatusPendingVerification = "PENDING_VERIFICATION"
	BankilyStatusSuccess             = "SUCCESS"
	BankilyStatusFailed              = "FAILED"

	// BankilyGateway marks the transactions ledger rows written by the
	// automatic B-PAY path. Those rows are created already 'completed' and are
	// never subject to manual admin approval.
	BankilyGateway = "bankily_bpay"
)

// BankilyPayment is one B-PAY payment attempt. There is deliberately no
// passcode field: the passcode is single-use input forwarded to Bankily and
// never stored.
type BankilyPayment struct {
	ID                   uuid.UUID       `db:"id"                     json:"id"`
	UserID               uuid.UUID       `db:"user_id"                json:"user_id"`
	OperationID          string          `db:"operation_id"           json:"operation_id"`
	IdempotencyKey       *string         `db:"idempotency_key"        json:"-"`
	Amount               decimal.Decimal `db:"amount"                 json:"amount"`
	CurrencyCode         string          `db:"currency_code"          json:"currency_code"`
	ClientPhone          string          `db:"client_phone"           json:"client_phone"`
	Language             string          `db:"language"               json:"language"`
	Status               string          `db:"status"                 json:"status"`
	BankilyTransactionID *string         `db:"bankily_transaction_id" json:"bankily_transaction_id"`
	BankilyErrorCode     *string         `db:"bankily_error_code"     json:"bankily_error_code"`
	BankilyErrorMessage  *string         `db:"bankily_error_message"  json:"bankily_error_message"`
	BankilyStatus        *string         `db:"bankily_status"         json:"bankily_status"`
	FailureReason        *string         `db:"failure_reason"         json:"failure_reason"`
	WalletTransactionID  *uuid.UUID      `db:"wallet_transaction_id"  json:"wallet_transaction_id"`
	CheckAttempts        int             `db:"check_attempts"         json:"check_attempts"`
	NeedsReview          bool            `db:"needs_review"           json:"needs_review"`
	LastCheckedAt        *time.Time      `db:"last_checked_at"        json:"last_checked_at"`
	NextCheckAt          *time.Time      `db:"next_check_at"          json:"next_check_at"`
	SubmittedAt          *time.Time      `db:"submitted_at"           json:"submitted_at"`
	CompletedAt          *time.Time      `db:"completed_at"           json:"completed_at"`
	CreatedAt            time.Time       `db:"created_at"             json:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at"             json:"updated_at"`
}

// IsTerminal reports whether the attempt can no longer change.
func (p *BankilyPayment) IsTerminal() bool {
	return p.Status == BankilyStatusSuccess || p.Status == BankilyStatusFailed
}

// BankilyPaymentAdminRow adds the owner's identity for the admin list.
type BankilyPaymentAdminRow struct {
	BankilyPayment
	UserFullName *string `db:"user_full_name" json:"user_full_name"`
	UserPhone    *string `db:"user_phone"     json:"user_phone"`
}

// MaskPhone keeps only the first two and last two digits ("48****11"), for
// logs and the customer-facing API.
func MaskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	masked := []byte(phone)
	for i := 2; i < len(masked)-2; i++ {
		masked[i] = '*'
	}
	return string(masked)
}
