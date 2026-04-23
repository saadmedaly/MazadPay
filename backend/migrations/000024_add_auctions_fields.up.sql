-- ============================================================
-- Ajout de champs manquants à la table auctions
-- Support sous-catégories, état, marque, vérification, boost, vidéo
-- ============================================================

-- Ajouter sub_category_id (FK vers categories pour sous-catégories)
ALTER TABLE auctions ADD COLUMN sub_category_id INT REFERENCES categories(id);
CREATE INDEX idx_auctions_sub_category ON auctions(sub_category_id);

-- Ajouter condition (état de l'item: neuf, utilisé, etc.)
ALTER TABLE auctions ADD COLUMN condition VARCHAR(20) DEFAULT 'used';
ALTER TABLE auctions ADD CONSTRAINT chk_auction_condition 
    CHECK (condition IN ('new', 'used', 'refurbished', 'damaged'));

-- Ajouter brand (marque)
ALTER TABLE auctions ADD COLUMN brand VARCHAR(100);

-- Ajouter is_verified (vérification par admin)
ALTER TABLE auctions ADD COLUMN is_verified BOOLEAN DEFAULT FALSE;
CREATE INDEX idx_auctions_verified ON auctions(is_verified);

-- Ajouter boosted_until (pour les annonces boostées)
ALTER TABLE auctions ADD COLUMN boosted_until TIMESTAMP WITH TIME ZONE;
CREATE INDEX idx_auctions_boosted ON auctions(boosted_until) WHERE boosted_until IS NOT NULL;

-- Ajouter video_url (URL vidéo de l'annonce)
ALTER TABLE auctions ADD COLUMN video_url TEXT;
