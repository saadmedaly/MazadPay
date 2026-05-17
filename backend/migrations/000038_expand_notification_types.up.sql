-- Expand notification types constraint to include all missing types
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS chk_notif_type;
ALTER TABLE notifications ADD CONSTRAINT chk_notif_type CHECK (type IN (
    'bid', 'win', 'payment', 'system', 'ad', 'general', 'new_auction', 
    'transaction', 'report', 'auction_sold', 'new_message', 
    'auction_ending_soon', 'auction_approved', 'auction_rejected', 
    'banner_approved', 'banner_rejected', 'auction_pending', 
    'auction_won', 'auction_ended', 'payment_received', 'auction_reported'
));
