-- Migration to add App Ratings, App Complaints, and Sponsors features

-- 1. Update app_ratings table
ALTER TABLE app_ratings ADD COLUMN IF NOT EXISTS title VARCHAR(255);
ALTER TABLE app_ratings ADD COLUMN IF NOT EXISTS auction_id UUID REFERENCES auctions(id) ON DELETE CASCADE;
-- Remove unique constraint on user_id if we want users to rate different things (app vs auction)
-- but for app rating, usually one per user. Let's keep it unique per (user_id, auction_id)
ALTER TABLE app_ratings DROP CONSTRAINT IF EXISTS uq_rating_user;
ALTER TABLE app_ratings ADD CONSTRAINT uq_rating_user_auction UNIQUE (user_id, auction_id);

-- 2. Update reports table
ALTER TABLE reports ADD COLUMN IF NOT EXISTS type VARCHAR(20) DEFAULT 'auction';
ALTER TABLE reports ALTER COLUMN auction_id DROP NOT NULL;
-- Add constraint for type
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_report_type') THEN
        ALTER TABLE reports ADD CONSTRAINT chk_report_type CHECK (type IN ('auction', 'app'));
    END IF;
END $$;

-- 3. Create sponsors table
CREATE TABLE IF NOT EXISTS sponsors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    image_url TEXT NOT NULL,
    phone VARCHAR(20),
    link_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sponsors_active ON sponsors(is_active);
