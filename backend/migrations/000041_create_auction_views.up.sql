CREATE TABLE IF NOT EXISTS auction_views (
    auction_id UUID NOT NULL REFERENCES auctions(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (auction_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_auction_views_auction_id ON auction_views (auction_id);
CREATE INDEX IF NOT EXISTS idx_auction_views_user_id    ON auction_views (user_id);
