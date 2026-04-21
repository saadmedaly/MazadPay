-- Update notification types constraint to include additional types
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS chk_notif_type;
ALTER TABLE notifications ADD CONSTRAINT chk_notif_type CHECK (type IN ('bid', 'win', 'payment', 'system', 'ad', 'general', 'new_auction', 'transaction', 'report', 'auction_sold'));
