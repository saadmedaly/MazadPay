-- Fix super admin password hash to match PIN "0000"
-- New hash generated with bcrypt
UPDATE users 
SET password_hash = '$2a$10$D2F9d5PGxeoFqbI3FxC7bOTGjgbhpfmiDFukOEu4LuqXg1rmeMj8m'
WHERE phone = '22222222';