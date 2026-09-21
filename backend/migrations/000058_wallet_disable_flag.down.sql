DROP INDEX IF EXISTS uq_admin_debit_reference;
ALTER TABLE wallets DROP COLUMN IF EXISTS disabled_at;
ALTER TABLE wallets DROP COLUMN IF EXISTS disabled_by;
ALTER TABLE wallets DROP COLUMN IF EXISTS disabled_reason;
ALTER TABLE wallets DROP COLUMN IF EXISTS is_disabled;
