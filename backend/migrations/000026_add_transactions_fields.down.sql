-- ============================================================
-- Rollback: Suppression des champs ajoutés à la table transactions
-- ============================================================

-- Supprimer l'index
DROP INDEX IF EXISTS idx_transactions_payment_method;

-- Supprimer les colonnes (dans l'ordre inverse de l'ajout)
ALTER TABLE transactions DROP COLUMN IF EXISTS failure_reason;
ALTER TABLE transactions DROP COLUMN IF EXISTS description;
ALTER TABLE transactions DROP COLUMN IF EXISTS net_amount;
ALTER TABLE transactions DROP COLUMN IF EXISTS fee_amount;
ALTER TABLE transactions DROP COLUMN IF EXISTS payment_method;
ALTER TABLE transactions DROP COLUMN IF EXISTS receipt_image_temp;
