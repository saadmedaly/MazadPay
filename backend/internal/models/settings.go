package models

import (
	"time"

	"github.com/google/uuid"
)

// DefaultUserSettings mirrors the user_settings table's own column DEFAULTs
// (migration 000031) -- the canonical values GetUserSettings returns for a
// user who has never customized their settings (no row yet), so a missing
// row is a normal state, not an error (client feedback: Bug D).
func DefaultUserSettings(userID uuid.UUID) *UserSettings {
	return &UserSettings{
		UserID:             userID,
		Currency:           "MRU",
		Theme:              "auto",
		Language:           "ar",
		NotificationsEmail: true,
		NotificationsPush:  true,
		NotificationsSMS:   false,
		TwoFactorEnabled:   false,
	}
}

// UserSettingsUpdate carries the fields a caller may set via PUT
// /users/me/settings. Fields are pointers so BodyParser can distinguish
// "the client omitted this field" (nil) from "the client explicitly sent
// its zero value" (non-nil pointing at false/""), which ordinary non-pointer
// fields cannot: mobile's real PUT caller (SettingsPage._updateSetting)
// sends exactly one field per call (e.g. {"notifications_push": false}),
// so PUT is genuinely partial-update, not full-replacement (client
// feedback: Bug D hardening). The repository merges each nil pointer with
// the existing stored value (or the canonical default if no row exists yet)
// via a single atomic SQL UPSERT -- never a separate SELECT-then-INSERT,
// which would race under concurrent partial updates.
type UserSettingsUpdate struct {
	Currency           *string `json:"currency"`
	Theme              *string `json:"theme"`
	Language           *string `json:"language"`
	NotificationsEmail *bool   `json:"notifications_email"`
	NotificationsPush  *bool   `json:"notifications_push"`
	NotificationsSMS   *bool   `json:"notifications_sms"`
	TwoFactorEnabled   *bool   `json:"two_factor_enabled"`
}

type SystemSettings struct {
	ID        int        `db:"id" json:"id"`
	Key       string     `db:"key" json:"key"`
	Value     string     `db:"value" json:"value"`
	Type      string     `db:"type" json:"type"` // string, number, boolean, json
	UpdatedBy *uuid.UUID `db:"updated_by" json:"updated_by"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

// Default settings keys
const (
	SettingMaintenanceMode    = "maintenance_mode"
	SettingRegistrationOpen   = "registration_open"
	SettingMaxAuctionDuration = "max_auction_duration_hours"
	SettingDefaultInsurance   = "default_insurance_amount"
	SettingMinBidIncrement    = "min_bid_increment"
	SettingContactWhatsApp    = "contact_whatsapp"
	SettingContactEmail       = "contact_email"
	SettingTermsAr            = "terms_ar"
	SettingTermsFr            = "terms_fr"
	SettingTermsEn            = "terms_en"
)
