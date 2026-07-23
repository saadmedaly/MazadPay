-- Audit Logs Schema Improvement (Phase A) : ajoute des colonnes optionnelles pour
-- clarifier l'identité de l'acteur (admin/utilisateur/système), tracer IP/User-Agent,
-- et permettre un détail structuré (JSONB) — sans toucher aux colonnes existantes ni
-- casser les lectures/écritures actuelles (voir internal/repository/audit_repo.go).

ALTER TABLE audit_logs
    ADD COLUMN IF NOT EXISTS actor_id     UUID NULL,
    ADD COLUMN IF NOT EXISTS actor_type   VARCHAR(20) NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS ip_address   TEXT NULL,
    ADD COLUMN IF NOT EXISTS user_agent   TEXT NULL,
    ADD COLUMN IF NOT EXISTS details_json JSONB NULL,
    ADD COLUMN IF NOT EXISTS entity_key   TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor
    ON audit_logs (actor_type, actor_id);

CREATE INDEX IF NOT EXISTS idx_audit_logs_entity_key
    ON audit_logs (entity_type, entity_key)
    WHERE entity_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_audit_logs_action
    ON audit_logs (action);

-- Backfill sûr : uniquement pour les lignes existantes où actor_id n'est pas encore
-- renseigné. actor_type reste 'unknown' pour l'historique — on ne peut pas affirmer
-- rétroactivement si admin_id représentait un admin ou un utilisateur ordinaire (voir
-- rapport d'audit précédent : certaines écritures historiques utilisaient admin_id
-- pour l'ID d'un utilisateur non-admin, ex: receipt_uploaded). Ne jamais convertir
-- details (TEXT) en details_json ni extraire entity_key depuis l'historique — un
-- parsing rétroactif serait non fiable et pourrait produire des données trompeuses.
UPDATE audit_logs
SET actor_id = admin_id
WHERE actor_id IS NULL;
