-- Revert to previous notification types constraint
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS chk_notif_type;
ALTER TABLE notifications ADD CONSTRAINT chk_notif_type CHECK (type IN (
    'bid', 'win', 'payment', 'system', 'ad', 'general', 'new_auction', 
    'transaction', 'report', 'auction_sold'
));
