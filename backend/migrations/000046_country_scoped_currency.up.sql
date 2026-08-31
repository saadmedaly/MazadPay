-- ============================================================
-- Country-scoped currency support (V1): each account belongs to exactly one
-- market (its account_country_iso), and monetary records are denominated in
-- that market's currency. No FX conversion, no cross-market interaction in
-- this release -- see currencies/countries.currency_code/users.account_country_iso
-- below and the corresponding backend enforcement (auth/auction/bid services).
--
-- IMPORTANT: currency_code is a monetary-denomination detail only. Market
-- access/bidding authorization is decided by country equality
-- (account_country_iso / market_country_iso), NEVER by currency equality --
-- multiple countries share a currency (e.g. SN and CI both use XOF) but MUST
-- remain separate markets.
-- ============================================================

-- 1) currencies: ISO-4217 reference table. minor_units is the number of
-- decimal places conventionally used for the currency (e.g. TND=3, MRU=0,
-- MAD=2) -- required for correct rounding/formatting, not just cosmetic.
CREATE TABLE currencies (
    code         VARCHAR(3) PRIMARY KEY,
    minor_units  SMALLINT NOT NULL DEFAULT 2,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Data source: golang.org/x/text/currency (CLDR-derived), verified this
-- session against the 197 countries seeded by migration 000044. One manual,
-- documented override applied: MR -> MRU (see countries UPDATE below) because
-- the CLDR snapshot bundled in this module's release only contains the
-- legacy Mauritanian Ouguiya (MRO), not the New Ouguiya (MRU) that has been
-- Mauritania's actual legal tender since the 2018-01-01 ISO 4217 redenomination
-- (amendment 68). MRO must never be used for new MazadPay monetary records.
--
-- minor_units overrides applied for currencies with a non-default (non-2)
-- conventional decimal precision: 0 for MRU/XOF/XAF/CLP/DJF/GNF/ISK/JPY/KMF/
-- KRW/PYG/RWF/UGX/VND/VUV/XPF/BIF; 3 for BHD/IQD/JOD/KWD/LYD/OMR/TND.
INSERT INTO currencies (code, minor_units) VALUES
    ('XAF', 0), ('BAM', 2), ('XCD', 2), ('CRC', 2), ('DKK', 2), ('FJD', 2),
    ('ALL', 2), ('CDF', 2), ('GEL', 2), ('GTQ', 2), ('KGS', 2), ('MZN', 2),
    ('AED', 2), ('BGN', 2), ('ERN', 2), ('SZL', 2), ('PLN', 2), ('AMD', 2),
    ('LKR', 2), ('TJS', 2), ('BND', 2), ('EUR', 2), ('KMF', 0), ('MAD', 2),
    ('TWD', 2), ('BDT', 2), ('DOP', 2), ('GBP', 2), ('ISK', 0), ('JPY', 0),
    ('YER', 2), ('CVE', 2), ('UGX', 0), ('CZK', 2), ('LAK', 2), ('MYR', 2),
    ('PGK', 2), ('BHD', 3), ('BIF', 0), ('CHF', 2), ('JMD', 2), ('ZAR', 2),
    ('SOS', 2), ('VND', 0), ('BSD', 2), ('CLP', 0), ('COP', 2), ('USD', 2),
    ('GHS', 2), ('MVR', 2), ('RON', 2), ('AUD', 2), ('DZD', 2), ('IRR', 2),
    ('RSD', 2), ('SDG', 2), ('SGD', 2), ('THB', 2), ('TZS', 2), ('ETB', 2),
    ('MDL', 2), ('MNT', 2), ('RUB', 2), ('AFN', 2), ('AZN', 2), ('ILS', 2),
    ('JOD', 3), ('NAD', 2), ('UZS', 2), ('DJF', 0), ('SLL', 2), ('ZMW', 2),
    ('INR', 2), ('IQD', 3), ('MUR', 2), ('MWK', 2), ('SYP', 2), ('PAB', 2),
    ('BOB', 2), ('BRL', 2), ('GMD', 2), ('NGN', 2), ('SEK', 2), ('PHP', 2),
    ('GYD', 2), ('HTG', 2), ('BWP', 2), ('EGP', 2), ('NIO', 2), ('VEF', 2),
    ('PKR', 2), ('BBD', 2), ('BTN', 2), ('KRW', 0), ('MRU', 0), ('NPR', 2),
    ('TMT', 2), ('CNY', 2), ('KHR', 2), ('LRD', 2), ('PEN', 2), ('TOP', 2),
    ('UYU', 2), ('HUF', 2), ('MKD', 2), ('MMK', 2), ('NOK', 2), ('AOA', 2),
    ('OMR', 3), ('SRD', 2), ('TRY', 2), ('HRK', 2), ('KWD', 3), ('ARS', 2),
    ('KES', 2), ('KZT', 2), ('LBP', 2), ('KPW', 2), ('MGA', 2), ('NZD', 2),
    ('RWF', 0), ('SSP', 2), ('TND', 3), ('VUV', 0), ('WST', 2), ('STN', 2),
    ('TTD', 2), ('UAH', 2), ('GNF', 0), ('QAR', 2), ('SCR', 2), ('BYN', 2),
    ('BZD', 2), ('MXN', 2), ('SAR', 2), ('SBD', 2), ('XOF', 0), ('CAD', 2),
    ('CUP', 2), ('IDR', 2), ('LYD', 3), ('PYG', 0), ('HNL', 2);

-- 2) countries.currency_code: FK reference, nullable (the pre-existing 'TU'
-- row -- an invalid ISO-3166 code, NOT created by this migration, duplicate/
-- incorrect entry for Tunisia which is correctly seeded separately as 'TN' --
-- is deliberately left with currency_code NULL rather than guessed at).
ALTER TABLE countries ADD COLUMN currency_code VARCHAR(3) REFERENCES currencies(code);

UPDATE countries SET currency_code = 'EUR' WHERE code = 'AD';
UPDATE countries SET currency_code = 'AED' WHERE code = 'AE';
UPDATE countries SET currency_code = 'AFN' WHERE code = 'AF';
UPDATE countries SET currency_code = 'XCD' WHERE code = 'AG';
UPDATE countries SET currency_code = 'ALL' WHERE code = 'AL';
UPDATE countries SET currency_code = 'AMD' WHERE code = 'AM';
UPDATE countries SET currency_code = 'AOA' WHERE code = 'AO';
UPDATE countries SET currency_code = 'ARS' WHERE code = 'AR';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'AT';
UPDATE countries SET currency_code = 'AUD' WHERE code = 'AU';
UPDATE countries SET currency_code = 'AZN' WHERE code = 'AZ';
UPDATE countries SET currency_code = 'BAM' WHERE code = 'BA';
UPDATE countries SET currency_code = 'BBD' WHERE code = 'BB';
UPDATE countries SET currency_code = 'BDT' WHERE code = 'BD';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'BE';
UPDATE countries SET currency_code = 'XOF' WHERE code = 'BF';
UPDATE countries SET currency_code = 'BGN' WHERE code = 'BG';
UPDATE countries SET currency_code = 'BHD' WHERE code = 'BH';
UPDATE countries SET currency_code = 'BIF' WHERE code = 'BI';
UPDATE countries SET currency_code = 'XOF' WHERE code = 'BJ';
UPDATE countries SET currency_code = 'BND' WHERE code = 'BN';
UPDATE countries SET currency_code = 'BOB' WHERE code = 'BO';
UPDATE countries SET currency_code = 'BRL' WHERE code = 'BR';
UPDATE countries SET currency_code = 'BSD' WHERE code = 'BS';
UPDATE countries SET currency_code = 'BTN' WHERE code = 'BT';
UPDATE countries SET currency_code = 'BWP' WHERE code = 'BW';
UPDATE countries SET currency_code = 'BYN' WHERE code = 'BY';
UPDATE countries SET currency_code = 'BZD' WHERE code = 'BZ';
UPDATE countries SET currency_code = 'CAD' WHERE code = 'CA';
UPDATE countries SET currency_code = 'CDF' WHERE code = 'CD';
UPDATE countries SET currency_code = 'XAF' WHERE code = 'CF';
UPDATE countries SET currency_code = 'XAF' WHERE code = 'CG';
UPDATE countries SET currency_code = 'CHF' WHERE code = 'CH';
UPDATE countries SET currency_code = 'XOF' WHERE code = 'CI';
UPDATE countries SET currency_code = 'CLP' WHERE code = 'CL';
UPDATE countries SET currency_code = 'XAF' WHERE code = 'CM';
UPDATE countries SET currency_code = 'CNY' WHERE code = 'CN';
UPDATE countries SET currency_code = 'COP' WHERE code = 'CO';
UPDATE countries SET currency_code = 'CRC' WHERE code = 'CR';
UPDATE countries SET currency_code = 'CUP' WHERE code = 'CU';
UPDATE countries SET currency_code = 'CVE' WHERE code = 'CV';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'CY';
UPDATE countries SET currency_code = 'CZK' WHERE code = 'CZ';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'DE';
UPDATE countries SET currency_code = 'DJF' WHERE code = 'DJ';
UPDATE countries SET currency_code = 'DKK' WHERE code = 'DK';
UPDATE countries SET currency_code = 'XCD' WHERE code = 'DM';
UPDATE countries SET currency_code = 'DOP' WHERE code = 'DO';
UPDATE countries SET currency_code = 'DZD' WHERE code = 'DZ';
UPDATE countries SET currency_code = 'USD' WHERE code = 'EC';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'EE';
UPDATE countries SET currency_code = 'EGP' WHERE code = 'EG';
UPDATE countries SET currency_code = 'ERN' WHERE code = 'ER';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'ES';
UPDATE countries SET currency_code = 'ETB' WHERE code = 'ET';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'FI';
UPDATE countries SET currency_code = 'FJD' WHERE code = 'FJ';
UPDATE countries SET currency_code = 'USD' WHERE code = 'FM';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'FR';
UPDATE countries SET currency_code = 'XAF' WHERE code = 'GA';
UPDATE countries SET currency_code = 'GBP' WHERE code = 'GB';
UPDATE countries SET currency_code = 'XCD' WHERE code = 'GD';
UPDATE countries SET currency_code = 'GEL' WHERE code = 'GE';
UPDATE countries SET currency_code = 'GHS' WHERE code = 'GH';
UPDATE countries SET currency_code = 'GMD' WHERE code = 'GM';
UPDATE countries SET currency_code = 'GNF' WHERE code = 'GN';
UPDATE countries SET currency_code = 'XAF' WHERE code = 'GQ';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'GR';
UPDATE countries SET currency_code = 'GTQ' WHERE code = 'GT';
UPDATE countries SET currency_code = 'XOF' WHERE code = 'GW';
UPDATE countries SET currency_code = 'GYD' WHERE code = 'GY';
UPDATE countries SET currency_code = 'HNL' WHERE code = 'HN';
UPDATE countries SET currency_code = 'HRK' WHERE code = 'HR';
UPDATE countries SET currency_code = 'HTG' WHERE code = 'HT';
UPDATE countries SET currency_code = 'HUF' WHERE code = 'HU';
UPDATE countries SET currency_code = 'IDR' WHERE code = 'ID';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'IE';
UPDATE countries SET currency_code = 'ILS' WHERE code = 'IL';
UPDATE countries SET currency_code = 'INR' WHERE code = 'IN';
UPDATE countries SET currency_code = 'IQD' WHERE code = 'IQ';
UPDATE countries SET currency_code = 'IRR' WHERE code = 'IR';
UPDATE countries SET currency_code = 'ISK' WHERE code = 'IS';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'IT';
UPDATE countries SET currency_code = 'JMD' WHERE code = 'JM';
UPDATE countries SET currency_code = 'JOD' WHERE code = 'JO';
UPDATE countries SET currency_code = 'JPY' WHERE code = 'JP';
UPDATE countries SET currency_code = 'KES' WHERE code = 'KE';
UPDATE countries SET currency_code = 'KGS' WHERE code = 'KG';
UPDATE countries SET currency_code = 'KHR' WHERE code = 'KH';
UPDATE countries SET currency_code = 'AUD' WHERE code = 'KI';
UPDATE countries SET currency_code = 'KMF' WHERE code = 'KM';
UPDATE countries SET currency_code = 'XCD' WHERE code = 'KN';
UPDATE countries SET currency_code = 'KPW' WHERE code = 'KP';
UPDATE countries SET currency_code = 'KRW' WHERE code = 'KR';
UPDATE countries SET currency_code = 'KWD' WHERE code = 'KW';
UPDATE countries SET currency_code = 'KZT' WHERE code = 'KZ';
UPDATE countries SET currency_code = 'LAK' WHERE code = 'LA';
UPDATE countries SET currency_code = 'LBP' WHERE code = 'LB';
UPDATE countries SET currency_code = 'XCD' WHERE code = 'LC';
UPDATE countries SET currency_code = 'CHF' WHERE code = 'LI';
UPDATE countries SET currency_code = 'LKR' WHERE code = 'LK';
UPDATE countries SET currency_code = 'LRD' WHERE code = 'LR';
UPDATE countries SET currency_code = 'ZAR' WHERE code = 'LS';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'LT';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'LU';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'LV';
UPDATE countries SET currency_code = 'LYD' WHERE code = 'LY';
UPDATE countries SET currency_code = 'MAD' WHERE code = 'MA';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'MC';
UPDATE countries SET currency_code = 'MDL' WHERE code = 'MD';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'ME';
UPDATE countries SET currency_code = 'MGA' WHERE code = 'MG';
UPDATE countries SET currency_code = 'USD' WHERE code = 'MH';
UPDATE countries SET currency_code = 'MKD' WHERE code = 'MK';
UPDATE countries SET currency_code = 'XOF' WHERE code = 'ML';
UPDATE countries SET currency_code = 'MMK' WHERE code = 'MM';
UPDATE countries SET currency_code = 'MNT' WHERE code = 'MN';
-- Explicit current-currency override: see documented rationale above (2018
-- ISO 4217 redenomination, CLDR snapshot staleness). MRO must never be used.
UPDATE countries SET currency_code = 'MRU' WHERE code = 'MR';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'MT';
UPDATE countries SET currency_code = 'MUR' WHERE code = 'MU';
UPDATE countries SET currency_code = 'MVR' WHERE code = 'MV';
UPDATE countries SET currency_code = 'MWK' WHERE code = 'MW';
UPDATE countries SET currency_code = 'MXN' WHERE code = 'MX';
UPDATE countries SET currency_code = 'MYR' WHERE code = 'MY';
UPDATE countries SET currency_code = 'MZN' WHERE code = 'MZ';
UPDATE countries SET currency_code = 'NAD' WHERE code = 'NA';
UPDATE countries SET currency_code = 'XOF' WHERE code = 'NE';
UPDATE countries SET currency_code = 'NGN' WHERE code = 'NG';
UPDATE countries SET currency_code = 'NIO' WHERE code = 'NI';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'NL';
UPDATE countries SET currency_code = 'NOK' WHERE code = 'NO';
UPDATE countries SET currency_code = 'NPR' WHERE code = 'NP';
UPDATE countries SET currency_code = 'AUD' WHERE code = 'NR';
UPDATE countries SET currency_code = 'NZD' WHERE code = 'NZ';
UPDATE countries SET currency_code = 'OMR' WHERE code = 'OM';
UPDATE countries SET currency_code = 'PAB' WHERE code = 'PA';
UPDATE countries SET currency_code = 'PEN' WHERE code = 'PE';
UPDATE countries SET currency_code = 'PGK' WHERE code = 'PG';
UPDATE countries SET currency_code = 'PHP' WHERE code = 'PH';
UPDATE countries SET currency_code = 'PKR' WHERE code = 'PK';
UPDATE countries SET currency_code = 'PLN' WHERE code = 'PL';
UPDATE countries SET currency_code = 'ILS' WHERE code = 'PS';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'PT';
UPDATE countries SET currency_code = 'USD' WHERE code = 'PW';
UPDATE countries SET currency_code = 'PYG' WHERE code = 'PY';
UPDATE countries SET currency_code = 'QAR' WHERE code = 'QA';
UPDATE countries SET currency_code = 'RON' WHERE code = 'RO';
UPDATE countries SET currency_code = 'RSD' WHERE code = 'RS';
UPDATE countries SET currency_code = 'RUB' WHERE code = 'RU';
UPDATE countries SET currency_code = 'RWF' WHERE code = 'RW';
UPDATE countries SET currency_code = 'SAR' WHERE code = 'SA';
UPDATE countries SET currency_code = 'SBD' WHERE code = 'SB';
UPDATE countries SET currency_code = 'SCR' WHERE code = 'SC';
UPDATE countries SET currency_code = 'SDG' WHERE code = 'SD';
UPDATE countries SET currency_code = 'SEK' WHERE code = 'SE';
UPDATE countries SET currency_code = 'SGD' WHERE code = 'SG';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'SI';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'SK';
UPDATE countries SET currency_code = 'SLL' WHERE code = 'SL';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'SM';
UPDATE countries SET currency_code = 'XOF' WHERE code = 'SN';
UPDATE countries SET currency_code = 'SOS' WHERE code = 'SO';
UPDATE countries SET currency_code = 'SRD' WHERE code = 'SR';
UPDATE countries SET currency_code = 'SSP' WHERE code = 'SS';
UPDATE countries SET currency_code = 'STN' WHERE code = 'ST';
UPDATE countries SET currency_code = 'USD' WHERE code = 'SV';
UPDATE countries SET currency_code = 'SYP' WHERE code = 'SY';
UPDATE countries SET currency_code = 'SZL' WHERE code = 'SZ';
UPDATE countries SET currency_code = 'XAF' WHERE code = 'TD';
UPDATE countries SET currency_code = 'XOF' WHERE code = 'TG';
UPDATE countries SET currency_code = 'THB' WHERE code = 'TH';
UPDATE countries SET currency_code = 'TJS' WHERE code = 'TJ';
UPDATE countries SET currency_code = 'USD' WHERE code = 'TL';
UPDATE countries SET currency_code = 'TMT' WHERE code = 'TM';
UPDATE countries SET currency_code = 'TND' WHERE code = 'TN';
UPDATE countries SET currency_code = 'TOP' WHERE code = 'TO';
UPDATE countries SET currency_code = 'TRY' WHERE code = 'TR';
UPDATE countries SET currency_code = 'TTD' WHERE code = 'TT';
UPDATE countries SET currency_code = 'AUD' WHERE code = 'TV';
UPDATE countries SET currency_code = 'TWD' WHERE code = 'TW';
UPDATE countries SET currency_code = 'TZS' WHERE code = 'TZ';
UPDATE countries SET currency_code = 'UAH' WHERE code = 'UA';
UPDATE countries SET currency_code = 'UGX' WHERE code = 'UG';
UPDATE countries SET currency_code = 'USD' WHERE code = 'US';
UPDATE countries SET currency_code = 'UYU' WHERE code = 'UY';
UPDATE countries SET currency_code = 'UZS' WHERE code = 'UZ';
UPDATE countries SET currency_code = 'EUR' WHERE code = 'VA';
UPDATE countries SET currency_code = 'XCD' WHERE code = 'VC';
UPDATE countries SET currency_code = 'VEF' WHERE code = 'VE';
UPDATE countries SET currency_code = 'VND' WHERE code = 'VN';
UPDATE countries SET currency_code = 'VUV' WHERE code = 'VU';
UPDATE countries SET currency_code = 'WST' WHERE code = 'WS';
UPDATE countries SET currency_code = 'YER' WHERE code = 'YE';
UPDATE countries SET currency_code = 'ZAR' WHERE code = 'ZA';
UPDATE countries SET currency_code = 'ZMW' WHERE code = 'ZM';
UPDATE countries SET currency_code = 'USD' WHERE code = 'ZW';
-- 'TU' (pre-existing invalid/duplicate row, not created by this migration,
-- not the correct Tunisia code -- see 'TN' above) deliberately left NULL.

-- 3) users.account_country_iso: the canonical, explicit MazadPay account
-- market -- distinct from phone_country_iso (phone-number-region metadata
-- only) and from the legacy, overloaded country_code column (historically a
-- dial code, e.g. '+222', for pre-v2 users). Nullable initially: existing
-- Production-era users have no explicit selection recorded; the application
-- layer treats NULL as an effective 'MR' market at runtime (see Register/
-- RegisterLegacy/PlaceBid), so no backfill write is required by this
-- migration or before this release.
ALTER TABLE users ADD COLUMN account_country_iso CHAR(2) REFERENCES countries(code);

-- 4) auction_requests / auctions: market_country_iso + currency_code, stamped
-- once at creation time from the requester's/seller's account market and
-- never re-derived later (an auction's currency must not change retroactively
-- if the seller's own account market were ever edited afterward). Nullable
-- initially for the same backward-compatibility reason as users above --
-- existing rows (currently 0 in Production, confirmed this session) have no
-- explicit market; runtime fallback treats NULL as 'MR'/'MRU'.
ALTER TABLE auction_requests ADD COLUMN market_country_iso CHAR(2) REFERENCES countries(code);
ALTER TABLE auction_requests ADD COLUMN currency_code VARCHAR(3) REFERENCES currencies(code);

ALTER TABLE auctions ADD COLUMN market_country_iso CHAR(2) REFERENCES countries(code);
ALTER TABLE auctions ADD COLUMN currency_code VARCHAR(3) REFERENCES currencies(code);

-- 5) wallets.currency_code: a wallet balance must have an unambiguous
-- denomination. wallets has no immutable parent relationship that determines
-- currency (it is keyed only by user_id, and a user's own account market is
-- itself mutable metadata on the users row, not a foreign-key-enforced
-- immutable link) -- per the audit's own rule, this means wallets MUST carry
-- an explicit currency_code rather than relying on a join to users at read
-- time. One wallet per user in this V1 (no multi-currency wallets); its
-- currency is set once, from the user's account market, at wallet-creation
-- time.
ALTER TABLE wallets ADD COLUMN currency_code VARCHAR(3) REFERENCES currencies(code);

-- 6) transactions.currency_code: same reasoning as wallets -- a transaction
-- is a standalone historical financial record (deposit/withdrawal/etc.) that
-- must remain correctly denominated even if the user's account market or
-- wallet state changes later. Stamped at transaction-creation time.
ALTER TABLE transactions ADD COLUMN currency_code VARCHAR(3) REFERENCES currencies(code);
