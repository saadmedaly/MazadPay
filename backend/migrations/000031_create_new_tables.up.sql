-- ============================================================
-- Création des nouvelles tables pour fonctionnalités avancées
-- ============================================================

-- Table auction_car_details (pour les détails véhicule structurés)
CREATE TABLE auction_car_details (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id UUID REFERENCES auctions(id) ON DELETE CASCADE,
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    year INT,
    mileage INT,
    fuel_type VARCHAR(50),
    transmission VARCHAR(50),
    color VARCHAR(50),
    engine_size VARCHAR(20),
    vin VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_car_details_auction ON auction_car_details(auction_id);

-- Table payment_methods (pour gérer les méthodes de paiement disponibles)
CREATE TABLE payment_methods (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,
    name_ar VARCHAR(100) NOT NULL,
    name_fr VARCHAR(100) NOT NULL,
    name_en VARCHAR(100),
    logo_url TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    country_id INT REFERENCES countries(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_payment_methods_code ON payment_methods(code);
CREATE INDEX idx_payment_methods_country ON payment_methods(country_id);

-- Table delivery_drivers (pour les infos chauffeurs détaillées)
CREATE TABLE delivery_drivers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    vehicle_type VARCHAR(50),
    vehicle_plate VARCHAR(20),
    vehicle_color VARCHAR(50),
    license_number VARCHAR(50),
    rating DECIMAL(3,2),
    total_deliveries INT DEFAULT 0,
    is_available BOOLEAN DEFAULT TRUE,
    current_location_lat DECIMAL(10,8),
    current_location_lng DECIMAL(11,8),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_drivers_user ON delivery_drivers(user_id);
CREATE INDEX idx_drivers_available ON delivery_drivers(is_available);

-- Table auction_boosts (pour gérer les annonces boostées/featured)
CREATE TABLE auction_boosts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auction_id UUID REFERENCES auctions(id) ON DELETE CASCADE,
    boost_type VARCHAR(20) NOT NULL,
    start_at TIMESTAMP WITH TIME ZONE NOT NULL,
    end_at TIMESTAMP WITH TIME ZONE NOT NULL,
    amount DECIMAL(15,2),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_boosts_auction ON auction_boosts(auction_id);
CREATE INDEX idx_boosts_active ON auction_boosts(status, end_at);

-- Table user_settings (pour les préférences utilisateur détaillées)
CREATE TABLE user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    currency VARCHAR(3) DEFAULT 'MRU',
    theme VARCHAR(10) DEFAULT 'auto',
    language VARCHAR(5) DEFAULT 'ar',
    notifications_email BOOLEAN DEFAULT TRUE,
    notifications_push BOOLEAN DEFAULT TRUE,
    notifications_sms BOOLEAN DEFAULT FALSE,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table bid_auto_bids (pour les enchères automatiques/proxy bidding)
CREATE TABLE bid_auto_bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    auction_id UUID REFERENCES auctions(id) ON DELETE CASCADE,
    max_amount DECIMAL(15,2) NOT NULL,
    current_bid_amount DECIMAL(15,2),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_auto_bids_user ON bid_auto_bids(user_id);
CREATE INDEX idx_auto_bids_auction ON bid_auto_bids(auction_id);
CREATE INDEX idx_auto_bids_active ON bid_auto_bids(is_active);

-- Ajouter contraintes CHECK
ALTER TABLE auction_boosts ADD CONSTRAINT chk_boost_type 
    CHECK (boost_type IN ('featured', 'urgent', 'top'));
ALTER TABLE auction_boosts ADD CONSTRAINT chk_boost_status 
    CHECK (status IN ('active', 'completed', 'cancelled'));
ALTER TABLE user_settings ADD CONSTRAINT chk_theme 
    CHECK (theme IN ('light', 'dark', 'auto'));
