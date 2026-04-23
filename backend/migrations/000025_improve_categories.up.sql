-- ============================================================
-- Amélioration de la table categories pour sous-catégories
-- ============================================================

-- Ajouter is_active (catégorie active/inactive)
ALTER TABLE categories ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
CREATE INDEX idx_categories_active ON categories(is_active);

-- Ajouter image_url (image de la catégorie)
ALTER TABLE categories ADD COLUMN image_url TEXT;

-- Ajouter has_subcategories (flag pour optimiser les requêtes)
ALTER TABLE categories ADD COLUMN has_subcategories BOOLEAN DEFAULT FALSE;

-- Note: name_en est déjà ajouté dans la migration 000006_add_name_en_to_categories.up.sql
