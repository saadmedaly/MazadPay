-- Client feedback Note #5: admin's "Delete" button on the Auctions screen
-- silently failed (no visible effect, "لا تعمل") for any ENDED auction that
-- had ever received a bid, since every bid places a wallet_holds row
-- (insurance) and/or a transactions row against that auction_id, and
-- neither foreign key had an ON DELETE action -- Postgres defaults to
-- NO ACTION, so DELETE FROM auctions blocked with a foreign-key violation
-- the handler surfaced as a generic 500. A pristine active/pending auction
-- with zero bids/transactions never hit this, which is why the bug was
-- specific to ended (bid-having) auctions.
--
-- Fix: ON DELETE SET NULL for both, NOT CASCADE -- wallet_holds/transactions
-- rows are the project's existing financial/insurance audit trail
-- (wallet_repo.go's ReleaseHoldsForNonWinners only ever transitions a
-- hold's status, never deletes the row) and must survive the auction being
-- deleted, exactly like the ticket's "do not silently break ... wallet/
-- insurance, transactions" requirement. Both columns are already nullable
-- (no NOT NULL constraint), so this is a safe, non-destructive change.
--
-- auction_payments.auction_id is NOT NULL and would need a different
-- (non-SET-NULL) treatment if it were ever populated -- confirmed via
-- repo-wide search that no Go code anywhere inserts into this table (it is
-- dead schema), so it can never actually hold a row that would block a real
-- deletion. ON DELETE CASCADE added defensively so it can never become a
-- silent landmine if that table is ever wired up later; harmless today
-- since it is always empty.
ALTER TABLE transactions DROP CONSTRAINT transactions_auction_id_fkey;
ALTER TABLE transactions ADD CONSTRAINT transactions_auction_id_fkey
    FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE SET NULL;

ALTER TABLE wallet_holds DROP CONSTRAINT wallet_holds_auction_id_fkey;
ALTER TABLE wallet_holds ADD CONSTRAINT wallet_holds_auction_id_fkey
    FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE SET NULL;

ALTER TABLE auction_payments DROP CONSTRAINT auction_payments_auction_id_fkey;
ALTER TABLE auction_payments ADD CONSTRAINT auction_payments_auction_id_fkey
    FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE CASCADE;
