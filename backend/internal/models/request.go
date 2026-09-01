package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AuctionRequest struct {
	ID              uuid.UUID       `db:"id"                json:"id"`
	UserID          uuid.UUID       `db:"user_id"           json:"user_id" validate:"required"`
	CategoryID      int             `db:"category_id"       json:"category_id" validate:"required"`
	LocationID      *int            `db:"location_id"       json:"location_id"`
	TitleAr         string          `db:"title_ar"          json:"title_ar" validate:"required"`
	TitleFr         *string         `db:"title_fr"          json:"title_fr"`
	TitleEn         *string         `db:"title_en"          json:"title_en"`
	DescriptionAr   *string         `db:"description_ar"    json:"description_ar" validate:"required,min=10,max=5000"`
	DescriptionFr   *string         `db:"description_fr"    json:"description_fr" validate:"omitempty,min=10,max=5000"`
	DescriptionEn   *string         `db:"description_en"    json:"description_en" validate:"omitempty,min=10,max=5000"`
	StartPrice      decimal.Decimal `db:"start_price"       json:"start_price" validate:"required,gt=0"`
	MinIncrement    decimal.Decimal `db:"min_increment"     json:"min_increment" validate:"required,gt=0"`
	InsuranceAmount decimal.Decimal `db:"insurance_amount"  json:"insurance_amount" validate:"gte=0"`
	ReservePrice    *decimal.Decimal `db:"reserve_price"     json:"reserve_price"`
	BuyNowPrice     *decimal.Decimal `db:"buy_now_price"     json:"buy_now_price"`
	StartDate       time.Time       `db:"start_date"        json:"start_date" validate:"required"`
	EndDate         time.Time       `db:"end_date"          json:"end_date" validate:"required"`
	Images          JSONB           `db:"images"             json:"images"`
	Status          string          `db:"status"             json:"status"`
	AdminNotes      *string         `db:"admin_notes"        json:"admin_notes"`
	ReviewedBy      *uuid.UUID      `db:"reviewed_by"        json:"reviewed_by"`
	ReviewedAt      *time.Time      `db:"reviewed_at"        json:"reviewed_at"`
	CreatedAt       time.Time       `db:"created_at"         json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"         json:"updated_at"`
	Quantity        int             `db:"quantity"          json:"quantity"` // Nombre d'items (défaut: 1)
	// MarketCountryISO/CurrencyCode (migration 000046) are always server-derived
	// from the authenticated requester's account market, never client-supplied
	// -- assigned by the handler before validation, same trust pattern as
	// UserID (see request_handler.go CreateAuctionRequest). No validate tag:
	// requiring it here would force a client to guess a value it never
	// legitimately controls.
	MarketCountryISO *string `db:"market_country_iso" json:"market_country_iso,omitempty"`
	CurrencyCode     *string `db:"currency_code"      json:"currency_code,omitempty"`

	// InsurancePolicy (migration 000048): distinguishes an explicit admin
	// decision "no insurance required" (InsurancePolicyNotRequired) from the
	// accidental/default insurance_amount = 0 state that caused the V03
	// incident (bid_service.go). Two states only, DB DEFAULT 'required' --
	// every existing/legacy row and every new row that omits this field stays
	// under today's exact protection unless an admin explicitly flips it. Use
	// InsuranceRequired(), never this raw field, for any policy decision.
	InsurancePolicy string `db:"insurance_policy" json:"insurance_policy" validate:"omitempty,oneof=required not_required"`

	// Relations
	User *User `db:"-" json:"user,omitempty"`
}

// InsurancePolicyRequired/InsurancePolicyNotRequired are the only two valid
// values of InsurancePolicy (also enforced by the DB CHECK constraint,
// migration 000048).
const (
	InsurancePolicyRequired    = "required"
	InsurancePolicyNotRequired = "not_required"
)

// InsuranceRequired reports whether this request currently requires a
// positive insurance_amount. Defaults safely to true (required) for any
// empty/unexpected value, matching the DB column's DEFAULT 'required' --
// callers must use this, never compare InsurancePolicy directly, so an
// unrecognized value never accidentally reads as "no insurance needed".
func (r *AuctionRequest) InsuranceRequired() bool {
	return r.InsurancePolicy != InsurancePolicyNotRequired
}

// EffectiveMarketCountryISO/EffectiveCurrencyCode fall back to
// DefaultAccountCountryISO/DefaultCurrencyCode for requests predating
// migration 000046 (MarketCountryISO/CurrencyCode NULL). Callers must use
// these, never the raw fields directly.
func (r *AuctionRequest) EffectiveMarketCountryISO() string {
	if r.MarketCountryISO != nil && *r.MarketCountryISO != "" {
		return *r.MarketCountryISO
	}
	return DefaultAccountCountryISO
}

func (r *AuctionRequest) EffectiveCurrencyCode() string {
	if r.CurrencyCode != nil && *r.CurrencyCode != "" {
		return *r.CurrencyCode
	}
	return DefaultCurrencyCode
}

// MarshalJSON ensures currency_code/market_country_iso are ALWAYS present
// when *AuctionRequest is serialized directly (e.g. handlers.OK/PaginatedOK),
// using the Effective* fallback for legacy pre-migration-000046 rows instead
// of omitting the key.
func (r AuctionRequest) MarshalJSON() ([]byte, error) {
	type alias AuctionRequest
	return json.Marshal(struct {
		alias
		MarketCountryISO string `json:"market_country_iso"`
		CurrencyCode     string `json:"currency_code"`
	}{
		alias:            alias(r),
		MarketCountryISO: r.EffectiveMarketCountryISO(),
		CurrencyCode:     r.EffectiveCurrencyCode(),
	})
}

type BannerRequest struct {
	ID         uuid.UUID  `db:"id"         json:"id"`
	UserID     uuid.UUID  `db:"user_id"    json:"user_id" validate:"required"`
	TitleAr    string     `db:"title_ar"   json:"title_ar" validate:"required"`
	TitleFr    *string    `db:"title_fr"   json:"title_fr"`
	TitleEn    *string    `db:"title_en"   json:"title_en"`
	ImageURL      string     `db:"image_url"      json:"image_url" validate:"required,url"`
	TargetURL     *string    `db:"target_url"     json:"target_url" validate:"omitempty,url"`
	DescriptionAr *string    `db:"description_ar" json:"description_ar"`
	DescriptionFr *string    `db:"description_fr" json:"description_fr"`
	DescriptionEn *string    `db:"description_en" json:"description_en"`
	StartsAt      time.Time  `db:"starts_at"      json:"starts_at" validate:"required"`
	EndsAt        time.Time  `db:"ends_at"        json:"ends_at" validate:"required"`
	Status     string     `db:"status"     json:"status"`
	AdminNotes *string    `db:"admin_notes" json:"admin_notes"`
	ReviewedBy *uuid.UUID `db:"reviewed_by" json:"reviewed_by"`
	ReviewedAt *time.Time `db:"reviewed_at" json:"reviewed_at"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`

	// Relations
	User *User `db:"-" json:"user,omitempty"`
}
