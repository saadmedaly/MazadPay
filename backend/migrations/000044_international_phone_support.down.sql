-- ============================================================
-- Down migration for 000044_international_phone_support
-- ============================================================

-- Reverse users additions.
DROP INDEX IF EXISTS idx_users_phone_e164_unique;
ALTER TABLE users DROP COLUMN IF EXISTS phone_country_iso;
ALTER TABLE users DROP COLUMN IF EXISTS phone_e164;

-- Reverse countries additions.
ALTER TABLE countries DROP COLUMN IF EXISTS phone_max_length;
ALTER TABLE countries DROP COLUMN IF EXISTS phone_min_length;

-- Intentionally NOT deleting the seeded country rows: this is broad, inert reference
-- data (country names/flags/dial codes) that other rows may already reference by the
-- time this down-migration would ever run in practice, and removing ~195 rows on a
-- rollback is unsafe/unnecessary busywork for a down-migration whose only real job is
-- undoing schema changes. If a full rollback of the seed data is genuinely needed,
-- do it as a deliberate, reviewed follow-up, not automatically here.
