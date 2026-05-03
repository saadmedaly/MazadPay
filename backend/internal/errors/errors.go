package apperrors

import "errors"

var (
	// Général
	ErrNotFound     = errors.New("resource_not_found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrBadRequest   = errors.New("bad_request")
	
	// Auth
	ErrOTPExpired               = errors.New("otp_expired")
	ErrOTPInvalid               = errors.New("otp_invalid")
	ErrOTPMaxAttempts           = errors.New("otp_max_attempts")
	ErrOTPRateLimited           = errors.New("otp_rate_limited")
	ErrDuplicatePhone           = errors.New("phone_already_registered")
	ErrInvalidPin               = errors.New("invalid_pin")
	ErrWeakPin                  = errors.New("weak_pin")
	ErrAccountBlocked           = errors.New("account_blocked")
	ErrTwilioNotConfigured      = errors.New("twilio_not_configured")
	ErrResetPasswordRateLimited = errors.New("reset_password_rate_limited")

	// Enchères
	ErrAuctionNotActive = errors.New("auction_not_active")
	ErrAuctionEnded     = errors.New("auction_ended")
	ErrBidTooLow        = errors.New("bid_too_low")
	ErrBidConflict      = errors.New("bid_conflict") // Optimistic lock → le client retry
	ErrSelfBid          = errors.New("cannot_bid_own_auction")

	// Finance
	ErrInsufficientBalance = errors.New("insufficient_balance")
	ErrWalletLocked        = errors.New("wallet_locked")
	ErrReceiptRequired     = errors.New("receipt_required")

	// Chat / Messagerie
	ErrConversationNotFound     = errors.New("conversation_not_found")
	ErrNotConversationMember    = errors.New("not_conversation_member")
	ErrAlreadyInConversation    = errors.New("already_in_conversation")
	ErrCannotEditMessage      = errors.New("cannot_edit_message")
	ErrCannotDeleteMessage    = errors.New("cannot_delete_message")
	ErrInvalidMessageType     = errors.New("invalid_message_type")
	ErrMessageTooLarge        = errors.New("message_too_large")
	ErrFileTooLarge           = errors.New("file_too_large") // > 10MB
	ErrDirectConversationExists = errors.New("direct_conversation_already_exists")
	ErrorResponse = errors.New("error_response")
)
