-- ============================================================
-- Rollback: Suppression des optimisations de la table bids
-- ============================================================

-- Supprimer l'index
DROP INDEX IF EXISTS idx_bids_auction_amount;

-- Supprimer les colonnes (dans l'ordre inverse de l'ajout)
ALTER TABLE bids DROP COLUMN IF EXISTS is_anonymous;
ALTER TABLE bids DROP COLUMN IF EXISTS bidder_phone;
ALTER TABLE bids DROP COLUMN IF EXISTS bidder_name;
