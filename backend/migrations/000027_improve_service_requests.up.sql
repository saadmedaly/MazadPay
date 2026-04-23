-- ============================================================
-- Amélioration de la table service_requests pour livraison complète
-- ============================================================

-- Adresses détaillées
ALTER TABLE service_requests ADD COLUMN pickup_address TEXT;
ALTER TABLE service_requests ADD COLUMN delivery_address TEXT;

-- Contacts
ALTER TABLE service_requests ADD COLUMN pickup_contact_name VARCHAR(100);
ALTER TABLE service_requests ADD COLUMN pickup_contact_phone VARCHAR(20);
ALTER TABLE service_requests ADD COLUMN delivery_contact_name VARCHAR(100);
ALTER TABLE service_requests ADD COLUMN delivery_contact_phone VARCHAR(20);

-- Horaires
ALTER TABLE service_requests ADD COLUMN pickup_time TIMESTAMP WITH TIME ZONE;
ALTER TABLE service_requests ADD COLUMN delivery_time TIMESTAMP WITH TIME ZONE;

-- Détails des items
ALTER TABLE service_requests ADD COLUMN item_description TEXT;
ALTER TABLE service_requests ADD COLUMN item_images JSONB;

-- Métriques
ALTER TABLE service_requests ADD COLUMN weight DECIMAL(10,2);
ALTER TABLE service_requests ADD COLUMN distance DECIMAL(10,2);
ALTER TABLE service_requests ADD COLUMN duration INT;
