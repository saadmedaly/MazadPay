DROP INDEX IF EXISTS idx_countries_code_unique;
DROP INDEX IF EXISTS idx_countries_phone_code;
ALTER TABLE countries DROP COLUMN IF EXISTS country_code;
