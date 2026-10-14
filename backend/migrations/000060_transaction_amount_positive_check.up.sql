-- Financial Audit (Phase 12, database defense-in-depth): transactions.amount
-- had NO CHECK constraint of any kind, unlike bids.amount (chk_bid_amount)
-- and wallet_holds.amount (chk_hold_amount), which both already enforce
-- amount > 0 at the schema level. Every money-moving code path already
-- rejects a zero/negative amount in Go (InitiateDeposit, RequestWithdraw,
-- AdminAddBalance, AdminDeductBalance, and TransactionRepository.
-- UpdateStatus's own "dernier filet de sécurité" guard) -- this migration
-- adds the matching DB-level constraint as defense-in-depth against any
-- future code path (or direct data correction) that might bypass those Go
-- checks, consistent with how bids/wallet_holds are already protected.
--
-- Confirmed via a read-only query against Validation before writing this
-- migration: zero existing rows violate this constraint
-- (SELECT count(*) FROM transactions WHERE type IN
-- ('deposit','withdraw','admin_credit','admin_debit') AND amount <= 0
-- returned 0), so this is safe to add without a backfill/cleanup step.
--
-- Scoped to only the four money-moving types this audit examined
-- (deposit/withdraw/admin_credit/admin_debit) rather than every possible
-- `type` value, since other transaction types (e.g. insurance-related
-- ledger rows, if any exist with different sign conventions) were not
-- individually audited for this constraint and are deliberately left
-- unconstrained rather than assumed safe.
ALTER TABLE transactions ADD CONSTRAINT chk_transaction_amount_positive
    CHECK (
        type NOT IN ('deposit', 'withdraw', 'admin_credit', 'admin_debit')
        OR amount > 0
    );
