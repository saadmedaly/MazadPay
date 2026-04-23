-- Migration: 000033_add_quantity_requests.down.sql
-- Description: Remove quantity field from auction_requests table

-- ============================================================
-- Remove quantity field from auction_requests
-- ============================================================

DROP INDEX IF EXISTS idx_auction_requests_quantity;
ALTER TABLE auction_requests DROP COLUMN IF EXISTS quantity;
