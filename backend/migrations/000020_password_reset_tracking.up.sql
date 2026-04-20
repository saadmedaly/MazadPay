-- ============================================================
-- Suivi des tentatives de réinitialisation de mot de passe
-- ============================================================

CREATE TABLE password_reset_attempts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    success BOOLEAN DEFAULT FALSE,
    reason VARCHAR(100), -- Motif en cas d'échec
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_password_reset_phone_time ON password_reset_attempts(phone, created_at DESC);
CREATE INDEX idx_password_reset_ip_time ON password_reset_attempts(ip_address, created_at DESC);

-- Vue pour détecter les abus (plus de 5 tentatives en 1 heure)
CREATE VIEW password_reset_suspicious_activity AS
SELECT
    phone,
    COUNT(*) as attempt_count,
    MAX(created_at) as last_attempt,
    array_agg(DISTINCT ip_address) as ip_addresses
FROM password_reset_attempts
WHERE created_at > NOW() - INTERVAL '1 hour'
    AND success = FALSE
GROUP BY phone
HAVING COUNT(*) >= 5;

-- Fonction pour nettoyer les anciennes tentatives (> 30 jours)
CREATE OR REPLACE FUNCTION cleanup_old_password_reset_attempts()
RETURNS void AS $$
BEGIN
    DELETE FROM password_reset_attempts
    WHERE created_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;
