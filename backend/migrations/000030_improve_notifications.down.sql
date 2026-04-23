-- ============================================================
-- Rollback: Suppression des améliorations de la table notifications
-- ============================================================

-- Supprimer l'index
DROP INDEX IF EXISTS idx_notifications_priority;

-- Supprimer la contrainte
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS chk_notification_priority;

-- Supprimer les colonnes (dans l'ordre inverse de l'ajout)
ALTER TABLE notifications DROP COLUMN IF EXISTS image_url;
ALTER TABLE notifications DROP COLUMN IF EXISTS expires_at;
ALTER TABLE notifications DROP COLUMN IF EXISTS action_label;
ALTER TABLE notifications DROP COLUMN IF EXISTS action_url;
ALTER TABLE notifications DROP COLUMN IF EXISTS priority;
