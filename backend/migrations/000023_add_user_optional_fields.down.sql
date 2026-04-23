-- ============================================================
-- Rollback: Suppression des champs optionnels de la table users
-- ============================================================

-- Supprimer les contraintes CHECK
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_gender;
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_kyc_status;

-- Supprimer les indexes
DROP INDEX IF EXISTS idx_users_profile_completed;
DROP INDEX IF EXISTS idx_users_kyc_status;
DROP INDEX IF EXISTS idx_users_country_code;

-- Supprimer les colonnes (dans l'ordre inverse de l'ajout)
ALTER TABLE users DROP COLUMN IF EXISTS kyc_status;
ALTER TABLE users DROP COLUMN IF EXISTS profile_completed;
ALTER TABLE users DROP COLUMN IF EXISTS gender;
ALTER TABLE users DROP COLUMN IF EXISTS postal_code;
ALTER TABLE users DROP COLUMN IF EXISTS address;
ALTER TABLE users DROP COLUMN IF EXISTS date_of_birth;
ALTER TABLE users DROP COLUMN IF EXISTS country_code;
