ALTER TABLE auctions RENAME COLUMN title TO title_ar;
ALTER TABLE auctions RENAME COLUMN description TO description_ar;

ALTER TABLE auctions ADD COLUMN title_fr VARCHAR(200);
ALTER TABLE auctions ADD COLUMN title_en VARCHAR(200);
ALTER TABLE auctions ADD COLUMN description_fr TEXT;
ALTER TABLE auctions ADD COLUMN description_en TEXT;
