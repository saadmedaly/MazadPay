-- Migration: 000033_add_quantity_requests.up.sql
-- Description: Add quantity field to auction_requests table

-- ============================================================
-- Add quantity field to auction_requests
-- ============================================================

ALTER TABLE auction_requests ADD COLUMN quantity INT DEFAULT 1;
CREATE INDEX idx_auction_requests_quantity ON auction_requests(quantity);

-- Add comment
COMMENT ON COLUMN auction_requests.quantity IS 'Nombre d''items disponibles dans la demande d''enchère (défaut: 1)';
