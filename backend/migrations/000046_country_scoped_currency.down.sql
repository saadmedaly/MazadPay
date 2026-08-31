-- ============================================================
-- Down migration for 000046_country_scoped_currency
-- ============================================================

ALTER TABLE transactions DROP COLUMN IF EXISTS currency_code;
ALTER TABLE wallets DROP COLUMN IF EXISTS currency_code;

ALTER TABLE auctions DROP COLUMN IF EXISTS currency_code;
ALTER TABLE auctions DROP COLUMN IF EXISTS market_country_iso;

ALTER TABLE auction_requests DROP COLUMN IF EXISTS currency_code;
ALTER TABLE auction_requests DROP COLUMN IF EXISTS market_country_iso;

ALTER TABLE users DROP COLUMN IF EXISTS account_country_iso;

ALTER TABLE countries DROP COLUMN IF EXISTS currency_code;

DROP TABLE IF EXISTS currencies;
