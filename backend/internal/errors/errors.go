package apperrors

import "errors"

var (
	// Général
	ErrNotFound        = errors.New("resource_not_found")
	ErrUserNotFound    = errors.New("user_not_found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrBadRequest   = errors.New("bad_request")

	// Settings (Admin Settings Phase B — key-based authorization)
	ErrSettingKeyUnknown        = errors.New("setting_key_unknown")
	ErrSettingRequiresSuperAdmin = errors.New("setting_requires_super_admin")

	// Settings (Admin Settings Phase C — validation)
	ErrSettingInvalidType  = errors.New("setting_invalid_type")
	ErrSettingInvalidValue = errors.New("setting_invalid_value")
	
	// Auth
	ErrOTPExpired               = errors.New("otp_expired")
	ErrOTPInvalid               = errors.New("otp_invalid")
	ErrOTPMaxAttempts           = errors.New("otp_max_attempts")
	ErrOTPRateLimited           = errors.New("otp_rate_limited")
	ErrDuplicatePhone           = errors.New("phone_already_registered")
	ErrInvalidPhone             = errors.New("invalid_phone")
	ErrInvalidPin               = errors.New("invalid_pin")
	ErrWeakPin                  = errors.New("weak_pin")
	ErrAccountBlocked           = errors.New("account_blocked")
	ErrPhoneUnavailable         = errors.New("phone_unavailable")
	ErrWablasNotConfigured      = errors.New("wablas_not_configured")
	ErrResetPasswordRateLimited = errors.New("reset_password_rate_limited")

	// Enchères
	ErrAuctionNotActive       = errors.New("auction_not_active")
	ErrAuctionEnded           = errors.New("auction_ended")
	ErrBidTooLow              = errors.New("bid_too_low")
	ErrBidConflict            = errors.New("bid_conflict") // Optimistic lock → le client retry
	ErrSelfBid                = errors.New("cannot_bid_own_auction")
	ErrInsuranceNotSet        = errors.New("insurance_not_set")        // audit V03 : insurance_amount <= 0
	ErrInsufficientForInsurance = errors.New("insufficient_for_insurance") // audit V03 : balance < insurance_amount
	// ErrRequestInsuranceNotSet (client feedback A7 follow-up): distinct from
	// ErrInsuranceNotSet above (which fires at bid time) -- this fires at
	// REQUEST-APPROVAL time. The user-facing create/request form no longer
	// collects insurance_amount (client requirement: staff/admin decides it
	// during review, never the user) -- so a request approved without an
	// admin having first set a valid insurance_amount via
	// AdminUpdateAuctionRequest would otherwise create a live "active"
	// auction that ErrInsuranceNotSet then blocks EVERY bidder from, with no
	// way for the seller/admin to notice until a bidder complains. Caught
	// earlier, at approval time, with a distinct code so the admin UI can
	// show the right message at the right step.
	ErrRequestInsuranceNotSet = errors.New("request_insurance_not_set")
	// ErrCrossMarketBid (migration 000046, country-scoped currency V1) : le
	// bidder et l'auction n'appartiennent pas au même marché (account_country_iso
	// != market_country_iso). Vérifié par COUNTRY, jamais par égalité de devise
	// seule — plusieurs pays partagent une devise (ex: SN/CI = XOF) mais doivent
	// rester des marchés distincts.
	ErrCrossMarketBid = errors.New("cross_market_bid_not_allowed")
	// ErrWalletCurrencyMismatch (migration 000046) : la devise du wallet du
	// bidder ne correspond pas à la devise de l'auction — garde-fou financier
	// supplémentaire, ne devrait normalement jamais se déclencher si
	// ErrCrossMarketBid est correctement vérifié en premier (un wallet est
	// toujours dans la devise du marché du compte), mais vérifié explicitement
	// par sécurité/défense en profondeur avant tout gel de caution.
	ErrWalletCurrencyMismatch = errors.New("wallet_currency_mismatch")
	// ErrDuplicateBidder (Customer #38, replacing the old client feedback #19
	// permanent rule): a user cannot place two CONSECUTIVE bids on the same
	// auction -- they may bid again once someone ELSE has bid in between.
	// Fires when the caller is already auctions.last_bidder_id; see
	// BidService.PlaceBid / AuctionRepository.TryClaimBidTurn. The wire code
	// ("bid_already_placed") is unchanged so existing/mobile clients don't
	// need a new error code to recognize this case, only updated messaging.
	ErrDuplicateBidder = errors.New("bid_already_placed")

	// Finance
	ErrInsufficientBalance = errors.New("insufficient_balance")
	ErrWalletLocked        = errors.New("wallet_locked")
	ErrReceiptRequired     = errors.New("receipt_required")
	// Client feedback (deposit min/max limits): distinct codes so the
	// mobile UI can show the exact right message, same pattern as
	// ErrBidTooLow -- fires only for the generic wallet-top-up deposit
	// path (WalletService.InitiateDeposit with auctionRequestID == nil);
	// the auction-subscription path's amount is server-stamped, never
	// client-supplied, so it can never trigger either of these.
	ErrDepositTooLow  = errors.New("deposit_too_low")
	ErrDepositTooHigh = errors.New("deposit_too_high")

	// Customer #31: admin manual winner-insurance refund. Distinct codes so
	// the admin UI can show the exact right message rather than a generic
	// failure for each of these genuinely different rejection reasons.
	ErrAuctionNotEnded  = errors.New("auction_not_ended")   // status != 'ended'
	ErrNotAuctionWinner = errors.New("not_auction_winner")  // target user_id != auctions.winner_id
	ErrNoActiveHold     = errors.New("no_active_hold")      // no active wallet_holds row for (winner, auction) -- already refunded, or never held

	// ErrDuplicateAdminCredit (Customer #35): a second add-balance attempt for
	// the same anchor transaction was rejected by uq_admin_credit_reference
	// (migration 000054) -- the DB-level compare-and-set that makes a
	// double-click/retry safe, since an admin_credit ledger row is created
	// already 'completed' in one step (no pending-then-approve status
	// transition to guard against re-entry, unlike deposit/withdraw).
	ErrDuplicateAdminCredit = errors.New("admin_credit_already_applied")

	// MAZADPAY -- admin wallet controls (deduct/disable): same
	// duplicate-guard shape as ErrDuplicateAdminCredit, mirrored for the new
	// admin_debit ledger type (uq_admin_debit_reference, migration 000058).
	ErrDuplicateAdminDebit = errors.New("admin_debit_already_applied")
	// ErrInsufficientBalanceForDebit: an admin attempted to deduct more than
	// the wallet's current (unfrozen) balance -- distinct from
	// ErrInsufficientBalance so the admin UI can show an admin-specific
	// message rather than reusing the user-facing withdrawal/bid message.
	ErrInsufficientBalanceForDebit = errors.New("insufficient_balance_for_debit")
	// ErrWalletDisabled: the target wallet's is_disabled flag (migration
	// 000058) is set -- blocks new spend attempts (bidding, withdrawal
	// requests) until an admin re-enables it. Admin credit/debit are
	// deliberately exempt from this check (see models.Wallet.IsDisabled doc
	// comment).
	ErrWalletDisabled = errors.New("wallet_disabled")

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
