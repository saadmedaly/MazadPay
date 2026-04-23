-- ============================================================
-- Amélioration de la table delivery_timeline pour tracking GPS
-- ============================================================

-- Ajouter step_order (ordre du step pour tri)
ALTER TABLE delivery_timeline ADD COLUMN step_order INT;

-- Ajouter icon (icône pour l'affichage mobile)
ALTER TABLE delivery_timeline ADD COLUMN icon VARCHAR(50);

-- Ajouter coordonnées GPS pour tracking
ALTER TABLE delivery_timeline ADD COLUMN location_lat DECIMAL(10,8);
ALTER TABLE delivery_timeline ADD COLUMN location_lng DECIMAL(11,8);

-- Ajouter performed_by (qui a effectué l'action)
ALTER TABLE delivery_timeline ADD COLUMN performed_by VARCHAR(20);
ALTER TABLE delivery_timeline ADD CONSTRAINT chk_performed_by 
    CHECK (performed_by IN ('driver', 'system', 'user', 'admin') OR performed_by IS NULL);

-- Créer un index pour optimiser le tri par ordre
CREATE INDEX idx_delivery_timeline_order ON delivery_timeline(request_id, step_order);
