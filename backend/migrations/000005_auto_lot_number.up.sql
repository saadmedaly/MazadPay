CREATE SEQUENCE IF NOT EXISTS auctions_lot_number_seq START 1;

CREATE OR REPLACE FUNCTION generate_lot_number()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.lot_number IS NULL OR NEW.lot_number = '' THEN
        NEW.lot_number := 'LOT-' || LPAD(nextval('auctions_lot_number_seq')::TEXT, 2, '0');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_generate_lot_number ON auctions;
CREATE TRIGGER trg_generate_lot_number
BEFORE INSERT ON auctions
FOR EACH ROW
EXECUTE FUNCTION generate_lot_number();
