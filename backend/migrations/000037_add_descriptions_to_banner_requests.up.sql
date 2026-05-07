-- Add description columns to banner_requests table
ALTER TABLE banner_requests ADD COLUMN description_ar TEXT;
ALTER TABLE banner_requests ADD COLUMN description_fr TEXT;
ALTER TABLE banner_requests ADD COLUMN description_en TEXT;
