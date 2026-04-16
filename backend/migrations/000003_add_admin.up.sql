-- ============================================================
-- ADD ADMIN USER
-- ============================================================

-- Insert default admin user
-- phone: +22200000000
-- pin: 0000 (hashed with bcrypt)
INSERT INTO users (
    phone,
    password_hash,
    full_name,
    role,
    is_verified,
    is_active,
    language_pref,
    notifications_enabled,
    terms_accepted_at,
    last_login_at
) VALUES (
    '+22200000000',
    '$2a$10$PfQ/6v4o1SnL8jfO2r9JAewX6jQkBtKCW5PHmjG0JNLUn2sv6yJxK',
    'Admin User',
    'admin',
    TRUE,
    TRUE,
    'ar',
    TRUE,
    NOW(),
    NOW()
) ON CONFLICT (phone) DO NOTHING;
