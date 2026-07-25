-- Admin Authorization Phase 1C-A : fondation de schéma pour lier une invitation admin
-- à un numéro de téléphone cible spécifique (n'active aucune vérification pour
-- l'instant — voir Phase 1C-B). Jamais le numéro complet en clair : seul un hash
-- (comparaison) et une version masquée (affichage/audit) sont stockés.

ALTER TABLE admin_invitations
    ADD COLUMN IF NOT EXISTS target_phone_hash TEXT NULL,
    ADD COLUMN IF NOT EXISTS target_phone_masked TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_admin_invitations_target_phone_hash
    ON admin_invitations (target_phone_hash)
    WHERE target_phone_hash IS NOT NULL;
