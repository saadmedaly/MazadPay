-- Security fix (SEC-01): migrations 000018/000019 seeded a super_admin
-- account (phone 22222222) with a hardcoded, publicly-known bcrypt hash
-- corresponding to PIN "0000". That hash is visible in the migration files
-- themselves (000018_seed_super_admin.up.sql, 000019_fix_super_admin_password.up.sql),
-- so any deployment that never rotated this account's credential still has
-- a live, publicly-derivable super-admin login.
--
-- This migration does NOT touch 000018/000019 (deployed migrations are
-- treated as immutable) and does NOT delete the account (other rows may
-- reference it by user_id, and deleting a user is a much larger, unrelated
-- operation). Instead it deactivates login for that account ONLY if it
-- still carries the EXACT known default hash -- i.e. only if nobody has
-- ever rotated its credential. If the password_hash has changed (meaning
-- someone already set a real credential), this is a strict no-op: the
-- WHERE clause simply won't match, and the account is left completely
-- untouched.
--
-- After this runs, the account can be safely re-enabled with a real
-- credential via `cmd/bootstrap_super_admin` (reads SUPER_ADMIN_PHONE /
-- SUPER_ADMIN_PIN from required env vars, no defaults, never logs the PIN).
UPDATE users
SET is_active = FALSE
WHERE phone = '22222222'
  AND password_hash = '$2a$10$D2F9d5PGxeoFqbI3FxC7bOTGjgbhpfmiDFukOEu4LuqXg1rmeMj8m';
