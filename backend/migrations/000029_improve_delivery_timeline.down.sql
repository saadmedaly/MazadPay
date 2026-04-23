-- ============================================================
-- Rollback: Suppression des améliorations de la table delivery_timeline
-- ============================================================

-- Supprimer l'index
DROP INDEX IF EXISTS idx_delivery_timeline_order;

-- Supprimer la contrainte
ALTER TABLE delivery_timeline DROP CONSTRAINT IF EXISTS chk_performed_by;

-- Supprimer les colonnes (dans l'ordre inverse de l'ajout)
ALTER TABLE delivery_timeline DROP COLUMN IF EXISTS performed_by;
ALTER TABLE delivery_timeline DROP COLUMN IF EXISTS location_lng;
ALTER TABLE delivery_timeline DROP COLUMN IF EXISTS location_lat;
ALTER TABLE delivery_timeline DROP COLUMN IF EXISTS icon;
ALTER TABLE delivery_timeline DROP COLUMN IF EXISTS step_order;
