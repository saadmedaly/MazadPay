ALTER TABLE locations RENAME COLUMN city_name TO city_name_ar;
ALTER TABLE locations RENAME COLUMN area_name TO area_name_ar;
ALTER TABLE locations ADD COLUMN city_name_fr VARCHAR(100);
ALTER TABLE locations ADD COLUMN area_name_fr VARCHAR(100);

-- Initialiser les valeurs FR avec les valeurs AR pour éviter les NULL
UPDATE locations SET city_name_fr = city_name_ar, area_name_fr = area_name_ar;

ALTER TABLE locations ALTER COLUMN city_name_fr SET NOT NULL;
ALTER TABLE locations ALTER COLUMN city_name_fr SET DEFAULT '';
ALTER TABLE locations ALTER COLUMN area_name_fr SET DEFAULT '';
