ALTER TABLE auctions DROP COLUMN IF EXISTS last_bidder_id;

CREATE TABLE IF NOT EXISTS auction_bid_participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id UUID NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    first_bid_id UUID REFERENCES bids(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_auction_bid_participants_auction_user UNIQUE (auction_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_auction_bid_participants_user ON auction_bid_participants(user_id);

INSERT INTO auction_bid_participants (auction_id, user_id, first_bid_id, created_at)
SELECT DISTINCT ON (b.auction_id, b.user_id)
    b.auction_id, b.user_id, b.id, b.created_at
FROM bids b
WHERE b.auction_id IS NOT NULL AND b.user_id IS NOT NULL
ORDER BY b.auction_id, b.user_id, b.created_at ASC, b.id ASC;
