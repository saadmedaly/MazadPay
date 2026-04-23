-- ============================================================
-- Amélioration de la table notifications pour meilleure UX
-- ============================================================

-- Ajouter priority (priorité de la notification)
ALTER TABLE notifications ADD COLUMN priority VARCHAR(10) DEFAULT 'normal';
ALTER TABLE notifications ADD CONSTRAINT chk_notification_priority 
    CHECK (priority IN ('low', 'normal', 'high', 'urgent'));

-- Ajouter action_url (URL de deep link pour l'action)
ALTER TABLE notifications ADD COLUMN action_url TEXT;

-- Ajouter action_label (label du bouton d'action)
ALTER TABLE notifications ADD COLUMN action_label VARCHAR(50);

-- Ajouter expires_at (date d'expiration)
ALTER TABLE notifications ADD COLUMN expires_at TIMESTAMP WITH TIME ZONE;

-- Ajouter image_url (image attachée à la notification)
ALTER TABLE notifications ADD COLUMN image_url TEXT;

-- Créer un index pour optimiser les requêtes par priorité
CREATE INDEX idx_notifications_priority ON notifications(priority, is_read);
