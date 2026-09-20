ALTER TABLE transactions DROP CONSTRAINT transactions_auction_id_fkey;
ALTER TABLE transactions ADD CONSTRAINT transactions_auction_id_fkey
    FOREIGN KEY (auction_id) REFERENCES auctions(id);

ALTER TABLE wallet_holds DROP CONSTRAINT wallet_holds_auction_id_fkey;
ALTER TABLE wallet_holds ADD CONSTRAINT wallet_holds_auction_id_fkey
    FOREIGN KEY (auction_id) REFERENCES auctions(id);

ALTER TABLE auction_payments DROP CONSTRAINT auction_payments_auction_id_fkey;
ALTER TABLE auction_payments ADD CONSTRAINT auction_payments_auction_id_fkey
    FOREIGN KEY (auction_id) REFERENCES auctions(id);
