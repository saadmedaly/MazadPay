-- ============================================================
-- Rollback: Suppression des améliorations de la table categories
-- ============================================================

-- Supprimer les indexes
DROP INDEX IF EXISTS idx_categories_active;

-- Supprimer les colonnes (dans l'ordre inverse de l'ajout)
ALTER TABLE categories DROP COLUMN IF EXISTS has_subcategories;
ALTER TABLE categories DROP COLUMN IF EXISTS image_url;
ALTER TABLE categories DROP COLUMN IF EXISTS is_active;
