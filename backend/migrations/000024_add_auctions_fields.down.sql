-- ============================================================
-- Rollback: Suppression des champs ajoutés à la table auctions
-- ============================================================

-- Supprimer les contraintes et indexes
DROP INDEX IF EXISTS idx_auctions_boosted;
DROP INDEX IF EXISTS idx_auctions_verified;
ALTER TABLE auctions DROP CONSTRAINT IF EXISTS chk_auction_condition;
DROP INDEX IF EXISTS idx_auctions_sub_category;

-- Supprimer les colonnes (dans l'ordre inverse de l'ajout)
ALTER TABLE auctions DROP COLUMN IF EXISTS video_url;
ALTER TABLE auctions DROP COLUMN IF EXISTS boosted_until;
ALTER TABLE auctions DROP COLUMN IF EXISTS is_verified;
ALTER TABLE auctions DROP COLUMN IF EXISTS brand;
ALTER TABLE auctions DROP COLUMN IF EXISTS condition;
ALTER TABLE auctions DROP COLUMN IF EXISTS sub_category_id;
