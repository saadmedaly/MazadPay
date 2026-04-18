-- Create system_settings table for global configuration
CREATE TABLE IF NOT EXISTS system_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'string',
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Insert default settings
INSERT INTO system_settings (key, value, type) VALUES 
    ('maintenance_mode', 'false', 'boolean'),
    ('registration_open', 'true', 'boolean'),
    ('max_auction_duration_hours', '72', 'number'),
    ('default_insurance_amount', '1000', 'number'),
    ('min_bid_increment', '100', 'number'),
    ('contact_whatsapp', '47601175', 'string'),
    ('contact_email', 'mazadpay@gmail.com', 'string')
ON CONFLICT (key) DO NOTHING;