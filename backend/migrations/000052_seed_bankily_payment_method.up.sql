-- Customer Request #33: Bankily manual-transfer deposit flow. The
-- payment_methods table (000031) and its full deposit/receipt/admin-review
-- pipeline (wallet_handler.go, wallet_service.go, admin_service.go) are
-- already generic across payment methods and required no code changes --
-- the only missing piece is that no migration ever seeded any row into
-- payment_methods, so no method (including Bankily) was ever selectable in
-- the mobile deposit form. country_id is left NULL (global): GetPaymentMethods
-- (wallet_repo.go) returns all is_active=true rows unfiltered by country.
INSERT INTO payment_methods (code, name_ar, name_fr, name_en, is_active)
VALUES ('bankily', 'بنكيلي', 'Bankily', 'Bankily', true)
ON CONFLICT (code) DO NOTHING;
