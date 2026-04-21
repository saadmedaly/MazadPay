-- Create auction_requests table for users requesting to create auctions
CREATE TABLE auction_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INT NOT NULL REFERENCES categories(id),
    location_id INT REFERENCES locations(id),
    title_ar VARCHAR(200) NOT NULL,
    title_fr VARCHAR(200),
    title_en VARCHAR(200),
    description_ar TEXT,
    description_fr TEXT,
    description_en TEXT,
    start_price DECIMAL(12,2) NOT NULL,
    min_increment DECIMAL(12,2) NOT NULL,
    insurance_amount DECIMAL(12,2) NOT NULL,
    reserve_price DECIMAL(12,2),
    buy_now_price DECIMAL(12,2),
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    images JSONB,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    admin_notes TEXT,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create banner_requests table for users requesting to add banner ads
CREATE TABLE banner_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title_ar VARCHAR(200) NOT NULL,
    title_fr VARCHAR(200),
    title_en VARCHAR(200),
    image_url TEXT NOT NULL,
    target_url TEXT,
    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ends_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    admin_notes TEXT,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_auction_requests_user_id ON auction_requests(user_id);
CREATE INDEX idx_auction_requests_status ON auction_requests(status);
CREATE INDEX idx_auction_requests_created_at ON auction_requests(created_at DESC);

CREATE INDEX idx_banner_requests_user_id ON banner_requests(user_id);
CREATE INDEX idx_banner_requests_status ON banner_requests(status);
CREATE INDEX idx_banner_requests_created_at ON banner_requests(created_at DESC);
