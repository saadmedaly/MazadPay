ALTER TABLE auctions DROP COLUMN title_fr;
ALTER TABLE auctions DROP COLUMN title_en;
ALTER TABLE auctions DROP COLUMN description_fr;
ALTER TABLE auctions DROP COLUMN description_en;

ALTER TABLE auctions RENAME COLUMN title_ar TO title;
ALTER TABLE auctions RENAME COLUMN description_ar TO description;
