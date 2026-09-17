-- Customer Request #38: replace the permanent "one bid per user per
-- auction, ever" rule (client feedback #19, auction_bid_participants) with
-- "a user cannot place two CONSECUTIVE bids in a row" -- they may bid again
-- once someone else has bid in between. last_bidder_id tracks who placed
-- the CURRENT highest/latest bid, updated atomically alongside current_price
-- in the same optimistic-locked transaction PlaceBid already uses, so the
-- consecutive-bid check-and-set is race-safe the same way
-- TrySetWinnerAtomically/TryEndAuctionAtomically already are (WHERE-guarded
-- UPDATE ... RETURNING, never a prior read trusted across a race window).
ALTER TABLE auctions ADD COLUMN IF NOT EXISTS last_bidder_id UUID REFERENCES users(id);

-- Backfill: the current highest bidder (if any) on every existing auction
-- becomes its last_bidder_id, so an in-flight auction's consecutive-bid
-- guard is correct immediately after this migration runs, not just for
-- bids placed after it.
UPDATE auctions a
SET last_bidder_id = top.user_id
FROM (
    SELECT DISTINCT ON (auction_id) auction_id, user_id
    FROM bids
    ORDER BY auction_id, amount DESC, created_at DESC
) top
WHERE a.id = top.auction_id;

-- auction_bid_participants (migration 000050) enforced the OLD permanent
-- rule and has no other purpose (winner/finalization logic has never
-- referenced it -- see auction_service.go's FinalizeExpiredAuction, which
-- reads bids directly). Safe to drop entirely.
DROP TABLE IF EXISTS auction_bid_participants;
