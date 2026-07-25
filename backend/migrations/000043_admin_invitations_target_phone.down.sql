-- Rollback de 000043 : supprime uniquement l'index et les deux colonnes ajoutés.
-- Ne touche jamais aux colonnes d'origine (id, token, created_by, expires_at,
-- used_at, created_at).

DROP INDEX IF EXISTS idx_admin_invitations_target_phone_hash;

ALTER TABLE admin_invitations
    DROP COLUMN IF EXISTS target_phone_masked,
    DROP COLUMN IF EXISTS target_phone_hash;
