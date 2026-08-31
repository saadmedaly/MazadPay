-- ============================================================
-- Down migration for 000047_fix_lot_number_truncation.
--
-- WARNING: this restores the ORIGINAL, BUGGY function body from migration
-- 000005 (fixed 2-digit LPAD truncation). Rolling back REINTRODUCES the >99
-- collision defect -- any auction created after rollback, once the sequence
-- exceeds 99, will again collide with an existing 2-digit LOT-XX lot_number.
-- This down migration exists only for symmetry/rollback-testing; it is not a
-- semantically safe restoration once the sequence has advanced past 99 in a
-- real environment.
-- ============================================================

CREATE OR REPLACE FUNCTION generate_lot_number()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.lot_number IS NULL OR NEW.lot_number = '' THEN
        NEW.lot_number := 'LOT-' || LPAD(nextval('auctions_lot_number_seq')::TEXT, 2, '0');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
