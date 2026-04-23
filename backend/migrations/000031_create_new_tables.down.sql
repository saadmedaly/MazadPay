-- ============================================================
-- Rollback: Suppression des nouvelles tables créées
-- ============================================================

-- Supprimer les contraintes CHECK
ALTER TABLE user_settings DROP CONSTRAINT IF EXISTS chk_theme;
ALTER TABLE auction_boosts DROP CONSTRAINT IF EXISTS chk_boost_status;
ALTER TABLE auction_boosts DROP CONSTRAINT IF EXISTS chk_boost_type;

-- Supprimer les indexes
DROP INDEX IF EXISTS idx_auto_bids_active;
DROP INDEX IF EXISTS idx_auto_bids_auction;
DROP INDEX IF EXISTS idx_auto_bids_user;
DROP INDEX IF EXISTS idx_boosts_active;
DROP INDEX IF EXISTS idx_boosts_auction;
DROP INDEX IF EXISTS idx_drivers_available;
DROP INDEX IF EXISTS idx_drivers_user;
DROP INDEX IF EXISTS idx_payment_methods_country;
DROP INDEX IF EXISTS idx_payment_methods_code;
DROP INDEX IF EXISTS idx_car_details_auction;

-- Supprimer les tables (dans l'ordre inverse de la création)
DROP TABLE IF EXISTS bid_auto_bids;
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS auction_boosts;
DROP TABLE IF EXISTS delivery_drivers;
DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS auction_car_details;
