ALTER TABLE auction_requests DROP COLUMN IF EXISTS subscription_fee;

ALTER TABLE categories DROP CONSTRAINT IF EXISTS chk_categories_fee_tier;
ALTER TABLE categories DROP COLUMN IF EXISTS fee_tier;
