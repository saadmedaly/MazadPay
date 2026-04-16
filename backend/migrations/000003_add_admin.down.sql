-- ============================================================
-- DROP ADMIN USER
-- ============================================================

-- Remove admin user
DELETE FROM users WHERE phone = '+22200000000' AND role = 'admin';
