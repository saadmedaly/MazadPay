-- Migration: 000032_add_quantity.up.sql
-- Description: Add quantity field to auctions table for managing item quantities

-- ============================================================
-- Add quantity field to auctions
-- ============================================================

ALTER TABLE auctions ADD COLUMN quantity INT DEFAULT 1;
CREATE INDEX idx_auctions_quantity ON auctions(quantity);

-- Add comment
COMMENT ON COLUMN auctions.quantity IS 'Nombre d''items disponibles dans l''enchère (défaut: 1)';
