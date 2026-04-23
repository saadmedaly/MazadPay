-- ============================================================
-- Rollback: Suppression des améliorations de la table service_requests
-- ============================================================

-- Supprimer les colonnes (dans l'ordre inverse de l'ajout)
ALTER TABLE service_requests DROP COLUMN IF EXISTS duration;
ALTER TABLE service_requests DROP COLUMN IF EXISTS distance;
ALTER TABLE service_requests DROP COLUMN IF EXISTS weight;
ALTER TABLE service_requests DROP COLUMN IF EXISTS item_images;
ALTER TABLE service_requests DROP COLUMN IF EXISTS item_description;
ALTER TABLE service_requests DROP COLUMN IF EXISTS delivery_time;
ALTER TABLE service_requests DROP COLUMN IF EXISTS pickup_time;
ALTER TABLE service_requests DROP COLUMN IF EXISTS delivery_contact_phone;
ALTER TABLE service_requests DROP COLUMN IF EXISTS delivery_contact_name;
ALTER TABLE service_requests DROP COLUMN IF EXISTS pickup_contact_phone;
ALTER TABLE service_requests DROP COLUMN IF EXISTS pickup_contact_name;
ALTER TABLE service_requests DROP COLUMN IF EXISTS delivery_address;
ALTER TABLE service_requests DROP COLUMN IF EXISTS pickup_address;
