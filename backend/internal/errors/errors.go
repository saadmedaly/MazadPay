package apperrors

import "errors"

var (
	// Général
	ErrNotFound    = errors.New("resource_not_found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden   = errors.New("forbidden")
	ErrConflict    = errors.New("conflict")
	ErrBadRequest  = errors.New("bad_request")

	// Auth
	ErrOTPExpired      = errors.New("otp_expired")
	ErrOTPInvalid      = errors.New("otp_invalid")
	ErrOTPMaxAttempts  = errors.New("otp_max_attempts")
	ErrOTPRateLimited  = errors.New("otp_rate_limited")
	ErrDuplicatePhone  = errors.New("phone_already_registered")
	ErrInvalidPin      = errors.New("invalid_pin")
	ErrAccountBlocked  = errors.New("account_blocked")

	// Enchères
	ErrAuctionNotActive = errors.New("auction_not_active")
	ErrAuctionEnded    = errors.New("auction_ended")
	ErrBidTooLow       = errors.New("bid_too_low")
	ErrBidConflict     = errors.New("bid_conflict")        // Optimistic lock → le client retry
	ErrSelfBid         = errors.New("cannot_bid_own_auction")

	// Finance
	ErrInsufficientBalance = errors.New("insufficient_balance")
	ErrWalletLocked        = errors.New("wallet_locked")
	ErrReceiptRequired     = errors.New("receipt_required")
)
