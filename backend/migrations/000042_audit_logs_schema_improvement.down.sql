-- Rollback de 000042 : supprime uniquement les index et colonnes ajoutés par cette
-- migration. Ne touche jamais aux colonnes d'origine (id, admin_id, action,
-- entity_type, entity_id, details, created_at).

DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_entity_key;
DROP INDEX IF EXISTS idx_audit_logs_actor;

ALTER TABLE audit_logs
    DROP COLUMN IF EXISTS entity_key,
    DROP COLUMN IF EXISTS details_json,
    DROP COLUMN IF EXISTS user_agent,
    DROP COLUMN IF EXISTS ip_address,
    DROP COLUMN IF EXISTS actor_type,
    DROP COLUMN IF EXISTS actor_id;
