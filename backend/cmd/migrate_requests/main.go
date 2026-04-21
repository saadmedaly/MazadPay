package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "mazadpay"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "mazadpay_secret"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "mazadpay"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sql := `
-- Create auction_requests table for users requesting to create auctions
CREATE TABLE IF NOT EXISTS auction_requests (
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
CREATE TABLE IF NOT EXISTS banner_requests (
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
CREATE INDEX IF NOT EXISTS idx_auction_requests_user_id ON auction_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_auction_requests_status ON auction_requests(status);
CREATE INDEX IF NOT EXISTS idx_auction_requests_created_at ON auction_requests(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_banner_requests_user_id ON banner_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_banner_requests_status ON banner_requests(status);
CREATE INDEX IF NOT EXISTS idx_banner_requests_created_at ON banner_requests(created_at DESC);
`

	_, err = db.Exec(sql)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Requests tables migration applied successfully!")
}
