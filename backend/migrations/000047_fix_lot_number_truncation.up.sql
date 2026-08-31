-- ============================================================
-- Fix generate_lot_number() truncation bug (migration 000005): the trigger used
-- LPAD(nextval('auctions_lot_number_seq')::TEXT, 2, '0'), and PostgreSQL's LPAD
-- TRUNCATES the input when it already exceeds the target width -- confirmed
-- directly: LPAD('100', 2, '0') = '10', LPAD('1000', 2, '0') = '10'. Once the
-- sequence passed 99, every newly generated lot_number silently collapsed to its
-- first two digits, guaranteeing a collision with an already-used value (verified:
-- 85 of the 100 possible 2-digit LOT-XX values were already taken in this
-- database before this fix).
--
-- Fix: pad to a MINIMUM width of 2, but let the width grow with the value's own
-- digit count so nothing is ever truncated:
--   GREATEST(2, LENGTH(v::TEXT))
-- 1 -> '01', 99 -> '99' (unchanged), 100 -> '100', 1000 -> '1000' (no longer
-- truncated). No sequence reset, no renumbering of existing rows, no new
-- sequence -- only the function body changes. auctions.lot_number is
-- VARCHAR(50), ample room for any realistic sequence value.
-- ============================================================

CREATE OR REPLACE FUNCTION generate_lot_number()
RETURNS TRIGGER AS $$
DECLARE
    next_val BIGINT;
BEGIN
    IF NEW.lot_number IS NULL OR NEW.lot_number = '' THEN
        next_val := nextval('auctions_lot_number_seq');
        NEW.lot_number := 'LOT-' || LPAD(next_val::TEXT, GREATEST(2, LENGTH(next_val::TEXT)), '0');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
