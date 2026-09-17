-- Customer Request #35: admin-initiated wallet balance credit. Unlike
-- deposit/withdraw, an admin_credit ledger row is created already
-- 'completed' in a single step (no pending-then-approve transition), so
-- there is no existing status-guard to protect against a double-click/retry
-- creating two credits. A partial unique index on (user_id, reference) --
-- reference stamped as 'admin_credit:<anchor_transaction_id>' -- makes a
-- second attempt for the same anchor transaction fail at the DB level
-- instead of relying on a read-then-write race window.
CREATE UNIQUE INDEX IF NOT EXISTS uq_admin_credit_reference
    ON transactions (user_id, reference)
    WHERE type = 'admin_credit' AND reference IS NOT NULL;
