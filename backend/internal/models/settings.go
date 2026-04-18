package models

import (
	"time"

	"github.com/google/uuid"
)

type SystemSettings struct {
	ID        int       `db:"id" json:"id"`
	Key      string    `db:"key" json:"key"`
	Value   string    `db:"value" json:"value"`
	Type    string    `db:"type" json:"type"` // string, number, boolean, json
	UpdatedBy *uuid.UUID `db:"updated_by" json:"updated_by"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Default settings keys
const (
	SettingMaintenanceMode = "maintenance_mode"
	SettingRegistrationOpen = "registration_open"
	SettingMaxAuctionDuration = "max_auction_duration_hours"
	SettingDefaultInsurance = "default_insurance_amount"
	SettingMinBidIncrement = "min_bid_increment"
	SettingContactWhatsApp = "contact_whatsapp"
	SettingContactEmail = "contact_email"
	SettingTermsAr = "terms_ar"
	SettingTermsFr = "terms_fr"
	SettingTermsEn = "terms_en"
)