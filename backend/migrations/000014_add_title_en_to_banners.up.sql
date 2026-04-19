-- Add title_en column to banners table
ALTER TABLE banners ADD COLUMN IF NOT EXISTS title_en VARCHAR(255);