package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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

	// Joined Fields (Metadata)
	CategoryNameAr *string `db:"category_name_ar" json:"category"`
	CityNameAr     *string `db:"city_name_ar"     json:"city"`
}

type AuctionImage struct {
	ID           int       `db:"id"           json:"id"`
	AuctionID    uuid.UUID `db:"auction_id"   json:"auction_id"`
	URL          string    `db:"url"          json:"url"`
	MediaType    string    `db:"media_type"   json:"media_type"`
	DisplayOrder int       `db:"display_order" json:"display_order"`
}

type Category struct {
	ID           int     `db:"id"            json:"id"`
	NameAr       string  `db:"name_ar"       json:"name_ar"`
	NameFr       string  `db:"name_fr"       json:"name_fr"`
	NameEn       string  `db:"name_en"       json:"name_en"`
	ParentID     *int    `db:"parent_id"     json:"parent_id"`
	IconName     *string `db:"icon_name"     json:"icon_name"`
	DisplayOrder int     `db:"display_order" json:"display_order"`
}

type Location struct {
	ID         int    `db:"id"           json:"id"`
	CityNameAr string `db:"city_name_ar" json:"city_name_ar"`
	CityNameFr string `db:"city_name_fr" json:"city_name_fr"`
	AreaNameAr string `db:"area_name_ar" json:"area_name_ar"`
	AreaNameFr string `db:"area_name_fr" json:"area_name_fr"`
}
