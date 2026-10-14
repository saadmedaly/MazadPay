-- MAZADPAY -- admin wallet controls: admins need to fully manage a user's
-- wallet, not just credit it. Adding balance (admin_credit) and now
-- deducting balance (admin_debit) both reuse the existing balance column and
-- transaction-ledger pattern -- no schema change needed for those.
--
-- "Disable/freeze the wallet" is a genuinely new concept, orthogonal to the
-- existing frozen_amount column: frozen_amount tracks a QUANTITY temporarily
-- earmarked for a specific pending withdrawal or bid-insurance hold (see
-- FreezeForWithdraw/DebitFreezeBalance), while is_disabled is a STATUS flag
-- meaning "this wallet, including its currently-unfrozen balance, must
-- reject all new spend attempts (bidding, withdrawal requests) regardless of
-- amount." The two concepts coexist independently: a disabled wallet can
-- still have a nonzero frozen_amount from a withdrawal that was already
-- in-flight before it was disabled.
--
-- disabled_reason/disabled_by/disabled_at give the admin UI enough to show
-- who disabled it and why, without solely relying on the separate audit_log
-- table for that immediate-glance context.
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS is_disabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS disabled_reason TEXT;
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS disabled_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE wallets ADD COLUMN IF NOT EXISTS disabled_at TIMESTAMPTZ;

-- Same double-click/retry idempotency guard as uq_admin_credit_reference
-- (migration 000054), mirrored for the new admin_debit ledger type.
CREATE UNIQUE INDEX IF NOT EXISTS uq_admin_debit_reference
    ON transactions (user_id, reference)
    WHERE type = 'admin_debit' AND reference IS NOT NULL;
