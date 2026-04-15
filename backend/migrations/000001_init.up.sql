CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- 1. AUTHENTIFICATION & UTILISATEURS
-- ============================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(100),
    email VARCHAR(150),
    profile_pic_url TEXT,
    city VARCHAR(50),
    language_pref VARCHAR(5) DEFAULT 'ar',
    notifications_enabled BOOLEAN DEFAULT TRUE,
    terms_accepted_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    is_verified BOOLEAN DEFAULT FALSE,
    blocked_until TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_user_role CHECK (role IN ('user', 'admin', 'driver'))
);

CREATE TABLE otp_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) NOT NULL,
    termii_pin_id VARCHAR(100) NOT NULL,
    purpose VARCHAR(20) NOT NULL DEFAULT 'register',
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE,
    ip_address VARCHAR(45),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_otp_attempts CHECK (attempts <= max_attempts),
    CONSTRAINT chk_otp_purpose CHECK (purpose IN ('register', 'reset_password'))
);

CREATE INDEX idx_otp_phone ON otp_verifications(phone);
CREATE INDEX idx_otp_expires ON otp_verifications(expires_at);

-- ============================================================
-- 2. TAXONOMIE & LOCALISATION
-- ============================================================

CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name_ar VARCHAR(100) NOT NULL,
    name_fr VARCHAR(100) NOT NULL,
    parent_id INT REFERENCES categories(id) ON DELETE CASCADE,
    icon_name VARCHAR(50),
    display_order INT DEFAULT 0
);

CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    city_name VARCHAR(100) NOT NULL,
    area_name VARCHAR(100) NOT NULL
);

-- ============================================================
-- 3. SYSTÈME D'ENCHÈRES
-- ============================================================

CREATE TABLE auctions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    seller_id UUID REFERENCES users(id) NOT NULL,
    category_id INT REFERENCES categories(id) NOT NULL,
    location_id INT REFERENCES locations(id),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    start_price DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    current_price DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    min_increment DECIMAL(15, 2) NOT NULL DEFAULT 100.00,
    insurance_amount DECIMAL(15, 2) DEFAULT 0.00,
    reserve_price DECIMAL(15, 2) DEFAULT 0.00,
    start_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    lot_number VARCHAR(50) UNIQUE,
    views INT DEFAULT 0,
    bidder_count INT DEFAULT 0,
    winner_id UUID REFERENCES users(id),
    winning_bid_id UUID,
    payment_deadline TIMESTAMP WITH TIME ZONE,
    is_featured BOOLEAN DEFAULT FALSE,
    featured_until TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    phone_contact VARCHAR(20),
    item_details JSONB,
    buy_now_price DECIMAL(15, 2),
    version INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_auction_prices CHECK (current_price >= start_price),
    CONSTRAINT chk_auction_increment CHECK (min_increment > 0),
    CONSTRAINT chk_auction_status CHECK (status IN ('pending', 'active', 'ended', 'canceled', 'rejected'))
);

CREATE TABLE bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id UUID REFERENCES auctions(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    amount DECIMAL(15, 2) NOT NULL,
    previous_price DECIMAL(15, 2),
    is_winning BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_bid_amount CHECK (amount > 0)
);

ALTER TABLE auctions ADD CONSTRAINT fk_auction_winning_bid
    FOREIGN KEY (winning_bid_id) REFERENCES bids(id) DEFERRABLE INITIALLY DEFERRED;

CREATE INDEX idx_bids_auction ON bids(auction_id);
CREATE INDEX idx_bids_user ON bids(user_id);
CREATE INDEX idx_bids_top ON bids(auction_id, amount DESC);

CREATE TABLE auction_images (
    id SERIAL PRIMARY KEY,
    auction_id UUID REFERENCES auctions(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    media_type VARCHAR(10) DEFAULT 'image',
    display_order INT DEFAULT 0
);

CREATE TABLE user_favorites (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    auction_id UUID REFERENCES auctions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, auction_id)
);

-- ============================================================
-- 4. FINANCES & PORTEFEUILLE
-- ============================================================

CREATE TABLE wallets (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    balance DECIMAL(15, 2) DEFAULT 0.00,
    frozen_amount DECIMAL(15, 2) DEFAULT 0.00,
    version INT DEFAULT 1,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_wallet_balance CHECK (balance >= 0),
    CONSTRAINT chk_wallet_frozen CHECK (frozen_amount >= 0)
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    auction_id UUID REFERENCES auctions(id),
    type VARCHAR(20) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    gateway VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending',
    reference VARCHAR(100),
    receipt_url TEXT,
    admin_notes TEXT,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE wallet_holds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    auction_id UUID REFERENCES auctions(id),
    amount DECIMAL(15, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    transaction_id UUID REFERENCES transactions(id),
    released_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_hold_status CHECK (status IN ('active', 'released', 'captured')),
    CONSTRAINT chk_hold_amount CHECK (amount > 0)
);

CREATE INDEX idx_holds_auction ON wallet_holds(auction_id, status);
CREATE INDEX idx_holds_user ON wallet_holds(user_id);

ALTER TABLE transactions ADD COLUMN wallet_hold_id UUID REFERENCES wallet_holds(id);

-- ============================================================
-- 5. NOTIFICATIONS
-- ============================================================

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,
    title VARCHAR(200) NOT NULL,
    body TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    reference_id UUID,
    reference_type VARCHAR(50),
    data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_notif_type CHECK (type IN ('bid', 'win', 'payment', 'system', 'ad'))
);

-- ============================================================
-- 6. SERVICES COMPLÉMENTAIRES (Livraison/Transport)
-- ============================================================

CREATE TABLE service_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    service_type VARCHAR(50) NOT NULL,
    pickup_location TEXT,
    delivery_location TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    tracking_number VARCHAR(20),
    estimated_price DECIMAL(15, 2),
    actual_price DECIMAL(15, 2),
    notes TEXT,
    driver_id UUID REFERENCES users(id),
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE delivery_timeline (
    id SERIAL PRIMARY KEY,
    request_id UUID REFERENCES service_requests(id) ON DELETE CASCADE,
    step_name VARCHAR(100) NOT NULL,
    description TEXT,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 7. CONTENU 
-- ============================================================

CREATE TABLE faq_items (
    id SERIAL PRIMARY KEY,
    question_ar TEXT NOT NULL,
    question_fr TEXT,
    answer_ar TEXT NOT NULL,
    answer_fr TEXT,
    display_order INT DEFAULT 0
);

CREATE TABLE banners (
    id SERIAL PRIMARY KEY,
    title_ar VARCHAR(200),
    title_fr VARCHAR(200),
    image_url TEXT NOT NULL,
    target_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    starts_at TIMESTAMP WITH TIME ZONE,
    ends_at TIMESTAMP WITH TIME ZONE,
    display_order INT DEFAULT 0
);

CREATE TABLE app_ratings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    rating INT CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_rating_user UNIQUE (user_id)
);

CREATE TABLE tutorials (
    id SERIAL PRIMARY KEY,
    title_ar VARCHAR(200) NOT NULL,
    title_fr VARCHAR(200),
    video_url TEXT NOT NULL,
    thumbnail_url TEXT,
    category VARCHAR(50), 
    display_order INT DEFAULT 0
);

CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id UUID REFERENCES auctions(id) ON DELETE CASCADE,
    reporter_id UUID REFERENCES users(id),
    reason TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    admin_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE kyc_verifications (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    id_card_front_url TEXT,
    id_card_back_url TEXT,
    nni_number VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending',
    admin_notes TEXT,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 8. NOUVELLES TABLES
-- ============================================================

CREATE TABLE push_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    fcm_token TEXT NOT NULL,
    device_id VARCHAR(100),
    platform VARCHAR(10) CHECK (platform IN ('android', 'ios')),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_fcm_token UNIQUE (fcm_token)
);
CREATE INDEX idx_push_tokens_user ON push_tokens(user_id);

CREATE TABLE auction_payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id UUID REFERENCES auctions(id) NOT NULL,
    winner_id UUID REFERENCES users(id) NOT NULL,
    transaction_id UUID REFERENCES transactions(id),
    amount DECIMAL(15, 2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    paid_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_auction_payment UNIQUE (auction_id),
    CONSTRAINT chk_payment_status CHECK (status IN ('pending', 'completed', 'failed', 'overdue'))
);
CREATE INDEX idx_auction_payments_winner ON auction_payments(winner_id);
CREATE INDEX idx_auction_payments_deadline ON auction_payments(deadline) WHERE status = 'pending';

CREATE TABLE blocked_phones (
    phone VARCHAR(20) PRIMARY KEY,
    reason TEXT,
    blocked_by UUID REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    blocked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 9. INDEX DE PERFORMANCE
-- ============================================================

CREATE INDEX idx_auctions_status ON auctions(status);
CREATE INDEX idx_auctions_seller ON auctions(seller_id);
CREATE INDEX idx_auctions_category ON auctions(category_id);
CREATE INDEX idx_auctions_end_time ON auctions(end_time);
CREATE INDEX idx_transactions_user ON transactions(user_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);
CREATE INDEX idx_favorites_user ON user_favorites(user_id);
CREATE INDEX idx_auctions_winner ON auctions(winner_id);
CREATE INDEX idx_auctions_featured ON auctions(is_featured) WHERE is_featured = TRUE;
CREATE INDEX idx_notif_reference ON notifications(reference_id) WHERE reference_id IS NOT NULL;
CREATE INDEX idx_banners_active ON banners(is_active, starts_at, ends_at);

-- ============================================================
-- 10. TRIGGERS AUTOMATIQUES
-- ============================================================

CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER trg_users_updated BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

CREATE TRIGGER trg_wallets_updated BEFORE UPDATE ON wallets
    FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
