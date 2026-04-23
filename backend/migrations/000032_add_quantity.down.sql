-- Migration: 000032_add_quantity.down.sql
-- Description: Remove quantity field from auctions table

-- ============================================================
-- Remove quantity field from auctions
-- ============================================================

DROP INDEX IF EXISTS idx_auctions_quantity;
ALTER TABLE auctions DROP COLUMN IF EXISTS quantity;
