-- Bankily B-PAY automatic wallet top-up (TEST phase).
--
-- Purely additive: one new table plus one partial unique index on
-- transactions that only ever matches rows written by the Bankily path
-- (gateway = 'bankily_bpay'). Manual deposits, admin credits/debits and every
-- existing row are untouched.
--
-- Money-safety invariants enforced by the database itself (not just code):
--   * operation_id is unique             -> one Bankily operation per attempt
--   * bankily_transaction_id is unique   -> one Bankily transaction never maps to two attempts
--   * (user_id, idempotency_key) unique  -> a retried/double-tapped client submission reuses its attempt
--   * wallet_transaction_id is unique    -> one attempt credits at most one ledger row
--   * status = SUCCESS <=> wallet_transaction_id IS NOT NULL
--   * transactions.reference unique for gateway = 'bankily_bpay'
--     ('bankily:<operation_id>')        -> one operation credits the wallet at most once
-- The B-PAY passcode is never stored: there is deliberately no column for it.

CREATE TABLE IF NOT EXISTS bankily_payments (
    id                      UUID PRIMARY KEY,
    user_id                 UUID NOT NULL REFERENCES users(id),
    operation_id            VARCHAR(64) NOT NULL,
    idempotency_key         VARCHAR(100),
    amount                  DECIMAL(15,2) NOT NULL,
    currency_code           VARCHAR(3) NOT NULL DEFAULT 'MRU',
    client_phone            VARCHAR(20) NOT NULL,
    language                VARCHAR(2) NOT NULL,
    status                  VARCHAR(30) NOT NULL DEFAULT 'INITIATED',
    bankily_transaction_id  VARCHAR(100),
    bankily_error_code      VARCHAR(20),
    bankily_error_message   TEXT,
    bankily_status          VARCHAR(10),
    failure_reason          VARCHAR(100),
    wallet_transaction_id   UUID REFERENCES transactions(id),
    check_attempts          INTEGER NOT NULL DEFAULT 0,
    needs_review            BOOLEAN NOT NULL DEFAULT FALSE,
    last_checked_at         TIMESTAMPTZ,
    next_check_at           TIMESTAMPTZ,
    submitted_at            TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_bankily_payments_status CHECK (
        status IN ('INITIATED', 'SUBMITTED', 'PENDING_VERIFICATION', 'SUCCESS', 'FAILED')
    ),
    CONSTRAINT chk_bankily_payments_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_bankily_payments_success_has_credit CHECK (
        (status = 'SUCCESS') = (wallet_transaction_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_bankily_payments_operation_id
    ON bankily_payments (operation_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_bankily_payments_bankily_tx_id
    ON bankily_payments (bankily_transaction_id)
    WHERE bankily_transaction_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_bankily_payments_user_idempotency_key
    ON bankily_payments (user_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_bankily_payments_wallet_tx
    ON bankily_payments (wallet_transaction_id)
    WHERE wallet_transaction_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_bankily_payments_user_created
    ON bankily_payments (user_id, created_at DESC);

-- Reconciliation worker scan: only unresolved rows, ordered by due time.
CREATE INDEX IF NOT EXISTS idx_bankily_payments_due
    ON bankily_payments (next_check_at)
    WHERE status IN ('SUBMITTED', 'PENDING_VERIFICATION');

CREATE INDEX IF NOT EXISTS idx_bankily_payments_status_created
    ON bankily_payments (status, created_at DESC);

-- Ledger-level guarantee: a Bankily operation can produce at most one wallet
-- credit row, whatever code path (inline response, reconciliation worker,
-- user re-check, admin re-check) gets there first.
CREATE UNIQUE INDEX IF NOT EXISTS uq_transactions_bankily_reference
    ON transactions (reference)
    WHERE gateway = 'bankily_bpay';
