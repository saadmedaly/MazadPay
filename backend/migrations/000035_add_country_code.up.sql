-- ============================================================
-- Ajout du country_code (indicatif téléphonique) à countries
-- ============================================================

-- Ajouter la colonne sans contrainte UNIQUE initialement
ALTER TABLE countries ADD COLUMN country_code VARCHAR(5);

-- Mettre à jour les pays existants avec les indicatifs
UPDATE countries SET country_code = '+222' WHERE code = 'MR';
UPDATE countries SET country_code = '+221' WHERE code = 'SN';
UPDATE countries SET country_code = '+212' WHERE code = 'MA';
UPDATE countries SET country_code = '+216' WHERE code = 'TN';

-- Créer index pour la performance
CREATE INDEX idx_countries_phone_code ON countries(country_code);

-- Ajouter contrainte UNIQUE partielle (seulement pour les valeurs non-NULL)
CREATE UNIQUE INDEX idx_countries_code_unique ON countries(country_code) WHERE country_code IS NOT NULL;
