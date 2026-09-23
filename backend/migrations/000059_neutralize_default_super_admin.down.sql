-- Reverts the SEC-01 neutralization -- only meaningful as a migration
-- rollback tool; re-activating a known-default-credential account is never
-- the correct real-world action. Restores is_active=TRUE for the seeded
-- account ONLY if it still carries the exact known default hash (i.e. this
-- migration's own up.sql is what deactivated it, not some unrelated admin
-- action taken afterward).
UPDATE users
SET is_active = TRUE
WHERE phone = '22222222'
  AND password_hash = '$2a$10$D2F9d5PGxeoFqbI3FxC7bOTGjgbhpfmiDFukOEu4LuqXg1rmeMj8m';
