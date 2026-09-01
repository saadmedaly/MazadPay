ALTER TABLE auctions DROP CONSTRAINT IF EXISTS chk_auctions_insurance_policy;
ALTER TABLE auctions DROP COLUMN IF EXISTS insurance_policy;

ALTER TABLE auction_requests DROP CONSTRAINT IF EXISTS chk_ar_insurance_policy;
ALTER TABLE auction_requests DROP COLUMN IF EXISTS insurance_policy;
