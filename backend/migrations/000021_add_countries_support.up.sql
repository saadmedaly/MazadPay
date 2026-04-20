-- ============================================================
-- Support Countries (Pays) - Hiérarchie Countries > Locations (Cities)
-- ============================================================

-- Créer table countries
CREATE TABLE countries (
    id SERIAL PRIMARY KEY,
    code VARCHAR(2) UNIQUE NOT NULL,
    name_ar VARCHAR(100) NOT NULL,
    name_fr VARCHAR(100) NOT NULL,
    name_en VARCHAR(100),
    flag_emoji VARCHAR(10),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Ajouter country_id à locations avec contrainte
ALTER TABLE locations ADD COLUMN country_id INT REFERENCES countries(id);

-- Créer index pour la performance
CREATE INDEX idx_locations_country_id ON locations(country_id);
CREATE INDEX idx_countries_code ON countries(code);

-- Insérer données initiales (Mauritanie et Sénégal)
INSERT INTO countries (code, name_ar, name_fr, name_en, flag_emoji, is_active) VALUES
('MR', 'موريتانيا', 'Mauritanie', 'Mauritania', '🇲🇷', TRUE),
('SN', 'السنغال', 'Sénégal', 'Senegal', '🇸🇳', TRUE)
ON CONFLICT (code) DO NOTHING;

-- Créer une fonction pour obtenir les pays avec villes
CREATE OR REPLACE FUNCTION get_countries_with_locations()
RETURNS TABLE (
    country_id INT,
    country_code VARCHAR(2),
    country_name_ar VARCHAR(100),
    country_name_fr VARCHAR(100),
    country_name_en VARCHAR(100),
    flag_emoji VARCHAR(10),
    locations_count INT
) AS $$
SELECT 
    c.id,
    c.code,
    c.name_ar,
    c.name_fr,
    c.name_en,
    c.flag_emoji,
    COALESCE(COUNT(l.id), 0)::INT
FROM countries c
LEFT JOIN locations l ON c.id = l.country_id
WHERE c.is_active = TRUE
GROUP BY c.id, c.code, c.name_ar, c.name_fr, c.name_en, c.flag_emoji
ORDER BY c.created_at;
$$ LANGUAGE SQL;
