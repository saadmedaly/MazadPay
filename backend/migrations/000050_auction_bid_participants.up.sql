-- Customer feedback #19: a user may successfully place at most one bid per
-- auction, ever -- even after being outbid. This does NOT rewrite or delete
-- any historical bid: bids stays exactly as-is (bid history, winning_bid_id,
-- and every existing row are untouched). Instead, a dedicated guard table
-- records "this user has already consumed their one participation on this
-- auction" and is the sole, DB-level-atomic source of truth for the rule --
-- see BidService.PlaceBid, which claims a row here in the SAME transaction
-- as the bid itself, so a rolled-back bid never consumes eligibility and a
-- true concurrent double-claim is rejected by the UNIQUE constraint below,
-- not by any race-prone application-level check.
CREATE TABLE auction_bid_participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id UUID NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    -- first_bid_id: the bid that claimed this slot, kept only for
    -- traceability/debugging -- never read by the eligibility check itself
    -- (existence of the guard row alone is authoritative).
    first_bid_id UUID REFERENCES bids(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_auction_bid_participants_auction_user UNIQUE (auction_id, user_id)
);

CREATE INDEX idx_auction_bid_participants_user ON auction_bid_participants(user_id);

-- Backfill: every user who has EVER placed a bid on an auction historically
-- is considered to have already consumed their one participation on it --
-- no historical bids row is deleted, updated, or deduplicated. Picks the
-- earliest bid per (auction_id, user_id) pair as first_bid_id purely for
-- traceability.
INSERT INTO auction_bid_participants (auction_id, user_id, first_bid_id, created_at)
SELECT DISTINCT ON (b.auction_id, b.user_id)
    b.auction_id, b.user_id, b.id, b.created_at
FROM bids b
WHERE b.auction_id IS NOT NULL AND b.user_id IS NOT NULL
ORDER BY b.auction_id, b.user_id, b.created_at ASC, b.id ASC;
