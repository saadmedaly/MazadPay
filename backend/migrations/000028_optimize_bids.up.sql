-- ============================================================
-- Optimisation de la table bids pour l'historique des enchères
-- ============================================================

-- Ajouter bidder_name (pour éviter les JOINs)
ALTER TABLE bids ADD COLUMN bidder_name VARCHAR(100);

-- Ajouter bidder_phone (pour éviter les JOINs)
ALTER TABLE bids ADD COLUMN bidder_phone VARCHAR(20);

-- Ajouter is_anonymous (option pour enchérisseur anonyme)
ALTER TABLE bids ADD COLUMN is_anonymous BOOLEAN DEFAULT FALSE;

-- Créer des indexes pour optimiser les requêtes d'historique
CREATE INDEX idx_bids_auction_amount ON bids(auction_id, amount DESC);
