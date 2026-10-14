-- Customer Request #21: deposit/withdrawal in-app notifications never appeared
-- because their types were never added to chk_notif_type after 000038, so every
-- INSERT attempt using them silently failed the CHECK constraint (the error was
-- logged and swallowed in NotificationService.SendPush, never surfacing to the
-- caller). deposit_submitted/withdrawal_submitted are new, dedicated submission-
-- acknowledgment types -- deposit_confirmed/deposit_rejected/withdrawal_processed
-- must never be reused for a still-pending request (see notification_service.go
-- callers).
--
-- bid_outbid (already used by bid_service.go:247, already localized in
-- notification_localizations.go) and banner_request (already used by
-- content_service.go:155 via NotifyAdminsLocalized, admin-targeted only) were
-- found to have the EXACT same pre-existing gap while writing this migration's
-- own regression tests -- neither was ever added to chk_notif_type by any prior
-- migration, so every real call using them has been silently failing to persist
-- in exactly the same way as deposit/withdrawal. Included here since they are
-- the same class of bug this migration exists to fix, not new/invented types.
--
-- Preserves every type already allowed by 000038 verbatim; purely additive.
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS chk_notif_type;
ALTER TABLE notifications ADD CONSTRAINT chk_notif_type CHECK (type IN (
    'bid', 'win', 'payment', 'system', 'ad', 'general', 'new_auction',
    'transaction', 'report', 'auction_sold', 'new_message',
    'auction_ending_soon', 'auction_approved', 'auction_rejected',
    'banner_approved', 'banner_rejected', 'auction_pending',
    'auction_won', 'auction_ended', 'payment_received', 'auction_reported',
    'deposit_confirmed', 'deposit_rejected', 'withdrawal_processed',
    'deposit_submitted', 'withdrawal_submitted', 'bid_outbid', 'banner_request'
));
