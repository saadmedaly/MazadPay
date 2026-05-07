-- Migration to revert App Ratings, App Complaints, and Sponsors features

-- 3. Drop sponsors table
DROP TABLE IF EXISTS sponsors;

-- 2. Revert reports table
ALTER TABLE reports DROP CONSTRAINT IF EXISTS chk_report_type;
ALTER TABLE reports DROP COLUMN IF EXISTS type;
-- Note: Making auction_id NOT NULL again might fail if there are 'app' reports.
-- So we only revert if user explicitly wants to.

-- 1. Revert app_ratings table
ALTER TABLE app_ratings DROP CONSTRAINT IF EXISTS uq_rating_user_auction;
ALTER TABLE app_ratings ADD CONSTRAINT uq_rating_user UNIQUE (user_id);
ALTER TABLE app_ratings DROP COLUMN IF EXISTS auction_id;
ALTER TABLE app_ratings DROP COLUMN IF EXISTS title;
