-- ============================================================
-- Revert: Suivi des tentatives de réinitialisation de mot de passe
-- ============================================================

DROP FUNCTION IF EXISTS cleanup_old_password_reset_attempts();
DROP VIEW IF EXISTS password_reset_suspicious_activity;
DROP TABLE IF EXISTS password_reset_attempts;
