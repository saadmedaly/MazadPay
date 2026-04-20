-- Remove is_super_admin column
ALTER TABLE users DROP COLUMN IF EXISTS is_super_admin;

-- Restore original constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_user_role;
ALTER TABLE users ADD CONSTRAINT chk_user_role CHECK (role::text = ANY (ARRAY['user'::character varying, 'admin'::character varying, 'driver'::character varying]::text[]));