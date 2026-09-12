-- Category fee tier (client feedback #4): distinguishes the 100 MRU
-- "ordinary" subscription fee from the 500 MRU "premium" fee (cars, real
-- estate). Two states only, DEFAULT 'standard' -- every existing category
-- (and every new one that omits this field) stays at the ordinary fee
-- unless an admin explicitly marks it 'premium'. Never derive the fee from
-- name_ar/name_fr/name_en, id, or icon_name -- those are not stable and
-- proved not to correspond consistently to category meaning across
-- environments (e.g. category id=1 is "Phones" in some environments, not
-- "Real Estate" as the original seed order implied).
ALTER TABLE categories ADD COLUMN fee_tier VARCHAR(20) NOT NULL DEFAULT 'standard';
ALTER TABLE categories ADD CONSTRAINT chk_categories_fee_tier
    CHECK (fee_tier IN ('standard', 'premium'));

-- auction_requests.subscription_fee: stamped from the request's category at
-- creation time (same "stamped once, immutable" pattern as currency_code/
-- market_country_iso, migration 000046) -- so a request's fee stays correct
-- even if the category's fee_tier is changed by an admin afterward.
ALTER TABLE auction_requests ADD COLUMN subscription_fee DECIMAL(15, 2) NOT NULL DEFAULT 100;
