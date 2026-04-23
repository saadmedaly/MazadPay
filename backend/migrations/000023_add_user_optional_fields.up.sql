-- ============================================================
-- Ajout de champs optionnels à la table users
-- Support multi-pays, livraison, KYC et personnalisation
-- ============================================================

-- Ajouter country_code (optionnel, default +222 pour Mauritanie)
ALTER TABLE users ADD COLUMN country_code VARCHAR(5) DEFAULT '+222';

-- Ajouter date_of_birth (optionnel pour KYC)
ALTER TABLE users ADD COLUMN date_of_birth DATE;

-- Ajouter address (optionnel pour livraison)
ALTER TABLE users ADD COLUMN address TEXT;

-- Ajouter postal_code (optionnel pour livraison)
ALTER TABLE users ADD COLUMN postal_code VARCHAR(20);

-- Ajouter gender (optionnel pour personnalisation)
ALTER TABLE users ADD COLUMN gender VARCHAR(10);

-- Ajouter profile_completed (flag pour savoir si le profil est complet)
ALTER TABLE users ADD COLUMN profile_completed BOOLEAN DEFAULT FALSE;

-- Ajouter kyc_status (statut de vérification KYC)
ALTER TABLE users ADD COLUMN kyc_status VARCHAR(20) DEFAULT 'none';

-- Créer des indexes pour optimiser les requêtes
CREATE INDEX idx_users_country_code ON users(country_code);
CREATE INDEX idx_users_kyc_status ON users(kyc_status);
CREATE INDEX idx_users_profile_completed ON users(profile_completed);

-- Ajouter des contraintes CHECK pour kyc_status
ALTER TABLE users ADD CONSTRAINT chk_kyc_status 
    CHECK (kyc_status IN ('none', 'pending', 'approved', 'rejected'));

-- Ajouter des contraintes CHECK pour gender
ALTER TABLE users ADD CONSTRAINT chk_gender 
    CHECK (gender IN ('male', 'female', 'other') OR gender IS NULL);
