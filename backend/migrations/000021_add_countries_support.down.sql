-- ============================================================
-- Revert: Support Countries (Pays)
-- ============================================================

-- Supprimer la fonction
DROP FUNCTION IF EXISTS get_countries_with_locations();

-- Supprimer l'index
DROP INDEX IF EXISTS idx_countries_code;

-- Supprimer country_id de locations
ALTER TABLE locations DROP COLUMN IF EXISTS country_id;

-- Supprimer la table countries
DROP TABLE IF EXISTS countries;
