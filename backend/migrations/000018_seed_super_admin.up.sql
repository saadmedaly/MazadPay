-- ============================================================
-- SEED SUPER ADMIN USER
-- ============================================================

-- Insert or update super admin user with phone 22222222
-- This ensures the super admin has proper flags for JWT authentication
INSERT INTO users (
    id,
    phone,
    password_hash,
    full_name,
    role,
    is_super_admin,
    is_verified,
    is_active,
    language_pref,
    notifications_enabled,
    terms_accepted_at,
    created_at,
    updated_at
) VALUES (
    '11111111-1111-1111-1111-111111111111',
    '22222222',
    '$2a$10$D2F9d5PGxeoFqbI3FxC7bOTGjgbhpfmiDFukOEu4LuqXg1rmeMj8m',
    'Super Admin',
    'super_admin',
    TRUE,
    TRUE,
    TRUE,
    'ar',
    TRUE,
    NOW(),
    NOW(),
    NOW()
) ON CONFLICT (phone) DO UPDATE SET
    role = 'super_admin',
    is_super_admin = TRUE,
    is_active = TRUE,
    is_verified = TRUE;