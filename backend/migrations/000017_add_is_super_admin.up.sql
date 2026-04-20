-- Add is_super_admin column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_super_admin BOOLEAN DEFAULT FALSE;

-- Update constraint to allow super_admin role
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_user_role;
ALTER TABLE users ADD CONSTRAINT chk_user_role CHECK (role::text = ANY (ARRAY['user'::character varying, 'admin'::character varying, 'super_admin'::character varying, 'driver'::character varying]::text[]));
