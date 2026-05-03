package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONB) Scan(src interface{}) error {
	if src == nil {
		*j = nil
		return nil
	}
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into JSONB", src)
	}
	return json.Unmarshal(b, j)
}

type Auction struct {
	ID              uuid.UUID        `db:"id"               json:"id"`
	SellerID        uuid.UUID        `db:"seller_id"        json:"seller_id"`
	CategoryID      int              `db:"category_id"      json:"category_id"`
	SubCategoryID   *int             `db:"sub_category_id"  json:"sub_category_id"`
	LocationID      *int             `db:"location_id"      json:"location_id"`
	TitleAr         string           `db:"title_ar"         json:"title_ar"`
	TitleFr         *string          `db:"title_fr"         json:"title_fr"`
	TitleEn         *string          `db:"title_en"         json:"title_en"`
	DescriptionAr   *string          `db:"description_ar"   json:"description_ar"`
	DescriptionFr   *string          `db:"description_fr"   json:"description_fr"`
	DescriptionEn   *string          `db:"description_en"   json:"description_en"`
	StartPrice      decimal.Decimal  `db:"start_price"      json:"start_price"`
	CurrentPrice    decimal.Decimal  `db:"current_price"    json:"current_price"`
	MinIncrement    decimal.Decimal  `db:"min_increment"    json:"min_increment"`
	InsuranceAmount decimal.Decimal  `db:"insurance_amount" json:"insurance_amount"`
	ReservePrice    decimal.Decimal  `db:"reserve_price"    json:"reserve_price"`
	StartTime       time.Time        `db:"start_time"       json:"start_time"`
	EndTime         time.Time        `db:"end_time"         json:"end_time"`
	Status          string           `db:"status"           json:"status"`
	LotNumber       *string          `db:"lot_number"       json:"lot_number"`
	Views           int              `db:"views"            json:"views"`
	BidderCount     int              `db:"bidder_count"     json:"bidder_count"`
	WinnerID        *uuid.UUID       `db:"winner_id"        json:"winner_id"`
	WinningBidID    *uuid.UUID       `db:"winning_bid_id"   json:"winning_bid_id"`
	PaymentDeadline *time.Time       `db:"payment_deadline" json:"payment_deadline"`
	IsFeatured      bool             `db:"is_featured"      json:"is_featured"`
	FeaturedUntil   *time.Time       `db:"featured_until"   json:"featured_until"`
	RejectionReason *string          `db:"rejection_reason" json:"rejection_reason"`
	PhoneContact    *string          `db:"phone_contact"    json:"-"` // Masqué avant envoi
	ItemDetails     JSONB            `db:"item_details"     json:"item_details"`
	BuyNowPrice     *decimal.Decimal `db:"buy_now_price"    json:"buy_now_price"`
	Version         int              `db:"version"          json:"version"`
	CreatedAt       time.Time        `db:"created_at"       json:"created_at"`
	// New fields from migration 000024
	Condition    *string    `db:"condition"         json:"condition"`
	Brand        *string    `db:"brand"             json:"brand"`
	IsVerified   bool       `db:"is_verified"      json:"is_verified"`
	BoostedUntil *time.Time `db:"boosted_until"    json:"boosted_until"`
	VideoURL     *string    `db:"video_url"         json:"video_url"`
	Quantity     int        `db:"quantity"         json:"quantity"` // Nombre d'items disponibles

	// Joined Fields (Metadata)
	CategoryNameAr *string `db:"category_name_ar" json:"category"`
	CityNameAr     *string `db:"city_name_ar"     json:"city"`
	ImageURLs      *string `db:"image_urls"       json:"image_urls"` // Comma-separated image URLs
}

// GetImagesArray returns image URLs as a slice of strings
func (a *Auction) GetImagesArray() []string {
	if a.ImageURLs == nil || *a.ImageURLs == "" {
		return []string{}
	}
	return strings.Split(*a.ImageURLs, ",")
}

type AuctionImage struct {
	ID           int       `db:"id"           json:"id"`
	AuctionID    uuid.UUID `db:"auction_id"   json:"auction_id"`
	URL          string    `db:"url"          json:"url"`
	MediaType    string    `db:"media_type"   json:"media_type"`
	DisplayOrder int       `db:"display_order" json:"display_order"`
}

type Category struct {
	ID           int     `db:"id"               json:"id"`
	NameAr       string  `db:"name_ar"          json:"name_ar"`
	NameFr       string  `db:"name_fr"          json:"name_fr"`
	NameEn       string  `db:"name_en"          json:"name_en"`
	ParentID     *int    `db:"parent_id"        json:"parent_id"`
	IconName     *string `db:"icon_name"        json:"icon_name"`
	DisplayOrder int     `db:"display_order"    json:"display_order"`
	// New fields from migration 000025
	IsActive         bool    `db:"is_active"         json:"is_active"`
	ImageURL         *string `db:"image_url"         json:"image_url"`
	HasSubcategories bool    `db:"has_subcategories" json:"has_subcategories"`
}

type Location struct {
	ID         int      `db:"id"            json:"id"`
	CountryID  *int     `db:"country_id"    json:"country_id"`
	CityNameAr string   `db:"city_name_ar"  json:"city_name_ar"`
	CityNameFr string   `db:"city_name_fr"  json:"city_name_fr"`
	AreaNameAr string   `db:"area_name_ar"  json:"area_name_ar"`
	AreaNameFr string   `db:"area_name_fr"  json:"area_name_fr"`
	Country    *Country `db:"-"            json:"country,omitempty"`
}

type Country struct {
	ID          int       `db:"id"           json:"id"`
	Code        string    `db:"code"         json:"code"`
	CountryCode *string   `db:"country_code" json:"country_code,omitempty"`
	NameAr      string    `db:"name_ar"      json:"name_ar"`
	NameFr      string    `db:"name_fr"      json:"name_fr"`
	NameEn      string    `db:"name_en"      json:"name_en"`
	FlagEmoji   string    `db:"flag_emoji"   json:"flag_emoji"`
	IsActive    bool      `db:"is_active"    json:"is_active"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}

// New models from migration 000031

type AuctionCarDetails struct {
	ID           uuid.UUID `db:"id"           json:"id"`
	AuctionID    uuid.UUID `db:"auction_id"   json:"auction_id"`
	Manufacturer *string   `db:"manufacturer" json:"manufacturer"`
	Model        *string   `db:"model"        json:"model"`
	Year         *int      `db:"year"         json:"year"`
	Mileage      *int      `db:"mileage"      json:"mileage"`
	FuelType     *string   `db:"fuel_type"    json:"fuel_type"`
	Transmission *string   `db:"transmission" json:"transmission"`
	Color        *string   `db:"color"        json:"color"`
	EngineSize   *string   `db:"engine_size"  json:"engine_size"`
	VIN          *string   `db:"vin"          json:"vin"`
	CreatedAt    time.Time `db:"created_at"   json:"created_at"`
}

type PaymentMethod struct {
	ID        int       `db:"id"         json:"id"`
	Code      string    `db:"code"       json:"code"`
	NameAr    string    `db:"name_ar"    json:"name_ar"`
	NameFr    string    `db:"name_fr"    json:"name_fr"`
	NameEn    *string   `db:"name_en"    json:"name_en"`
	LogoURL   *string   `db:"logo_url"   json:"logo_url"`
	IsActive  bool      `db:"is_active"  json:"is_active"`
	CountryID *int      `db:"country_id" json:"country_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type DeliveryDriver struct {
	ID                 uuid.UUID  `db:"id"                   json:"id"`
	UserID             *uuid.UUID `db:"user_id"              json:"user_id"`
	VehicleType        *string    `db:"vehicle_type"         json:"vehicle_type"`
	VehiclePlate       *string    `db:"vehicle_plate"        json:"vehicle_plate"`
	VehicleColor       *string    `db:"vehicle_color"        json:"vehicle_color"`
	LicenseNumber      *string    `db:"license_number"       json:"license_number"`
	Rating             *float64   `db:"rating"               json:"rating"`
	TotalDeliveries    int        `db:"total_deliveries"     json:"total_deliveries"`
	IsAvailable        bool       `db:"is_available"         json:"is_available"`
	CurrentLocationLat *float64   `db:"current_location_lat" json:"current_location_lat"`
	CurrentLocationLng *float64   `db:"current_location_lng" json:"current_location_lng"`
	CreatedAt          time.Time  `db:"created_at"           json:"created_at"`
}

type AuctionBoost struct {
	ID        uuid.UUID        `db:"id"         json:"id"`
	AuctionID uuid.UUID        `db:"auction_id" json:"auction_id"`
	BoostType string           `db:"boost_type" json:"boost_type"`
	StartAt   time.Time        `db:"start_at"   json:"start_at"`
	EndAt     time.Time        `db:"end_at"     json:"end_at"`
	Amount    *decimal.Decimal `db:"amount"     json:"amount"`
	Status    string           `db:"status"     json:"status"`
	CreatedAt time.Time        `db:"created_at" json:"created_at"`
}

type UserSettings struct {
	UserID             uuid.UUID `db:"user_id"                  json:"user_id"`
	Currency           string    `db:"currency"                  json:"currency"`
	Theme              string    `db:"theme"                     json:"theme"`
	Language           string    `db:"language"                 json:"language"`
	NotificationsEmail bool      `db:"notifications_email"     json:"notifications_email"`
	NotificationsPush  bool      `db:"notifications_push"      json:"notifications_push"`
	NotificationsSMS   bool      `db:"notifications_sms"       json:"notifications_sms"`
	TwoFactorEnabled   bool      `db:"two_factor_enabled"      json:"two_factor_enabled"`
	CreatedAt          time.Time `db:"created_at"               json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"               json:"updated_at"`
}

type BidAutoBid struct {
	ID               uuid.UUID        `db:"id"               json:"id"`
	UserID           uuid.UUID        `db:"user_id"         json:"user_id"`
	AuctionID        uuid.UUID        `db:"auction_id"      json:"auction_id"`
	MaxAmount        decimal.Decimal  `db:"max_amount"      json:"max_amount"`
	CurrentBidAmount *decimal.Decimal `db:"current_bid_amount" json:"current_bid_amount"`
	IsActive         bool             `db:"is_active"       json:"is_active"`
	CreatedAt        time.Time        `db:"created_at"      json:"created_at"`
}
