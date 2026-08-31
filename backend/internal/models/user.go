package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                   uuid.UUID  `db:"id"                    json:"id"`
	Phone                string     `db:"phone"                 json:"phone"`
	PasswordHash         string     `db:"password_hash"         json:"-"`
	FullName             *string    `db:"full_name"             json:"full_name"`
	Email                *string    `db:"email"                 json:"email"`
	ProfilePicURL        *string    `db:"profile_pic_url"       json:"profile_pic_url"`
	City                 *string    `db:"city"                  json:"city"`
	LanguagePref         string     `db:"language_pref"         json:"language_pref"`
	NotificationsEnabled bool       `db:"notifications_enabled" json:"notifications_enabled"`
	TermsAcceptedAt      *time.Time `db:"terms_accepted_at"     json:"terms_accepted_at"`
	IsActive             bool       `db:"is_active"             json:"is_active"`
	Role                 string     `db:"role"                  json:"role"`
	IsSuperAdmin         bool       `db:"is_super_admin"        json:"-"`
	IsVerified           bool       `db:"is_verified"           json:"is_verified"`
	BlockedUntil         *time.Time `db:"blocked_until"         json:"-"`
	LastLoginAt          *time.Time `db:"last_login_at"         json:"last_login_at"`
	CreatedAt            time.Time  `db:"created_at"            json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"            json:"updated_at"`
	// New fields from migration 000023
	CountryCode          *string    `db:"country_code"          json:"country_code"`
	DateOfBirth          *time.Time `db:"date_of_birth"         json:"date_of_birth"`
	Address              *string    `db:"address"               json:"address"`
	PostalCode           *string    `db:"postal_code"           json:"postal_code"`
	Gender               *string    `db:"gender"                json:"gender"`
	ProfileCompleted     bool       `db:"profile_completed"     json:"profile_completed"`
	KycStatus            *string    `db:"kyc_status"            json:"kyc_status"`
	// New fields from migration 000044: canonical E.164 phone + detected ISO region,
	// alongside the legacy Phone field (kept populated for backward compatibility).
	PhoneE164       *string `db:"phone_e164"        json:"phone_e164"`
	PhoneCountryISO *string `db:"phone_country_iso" json:"phone_country_iso"`
	// AccountCountryISO (migration 000046) is the canonical, explicit MazadPay
	// account market -- distinct from PhoneCountryISO (phone-number-region
	// metadata only) and from the legacy, overloaded CountryCode column
	// (historically a dial code). Nil for legacy users; EffectiveAccountCountryISO
	// below is the fallback-aware accessor callers should use instead of reading
	// this field directly.
	AccountCountryISO *string `db:"account_country_iso" json:"account_country_iso"`
}

// DefaultAccountCountryISO is the runtime fallback market for any user whose
// AccountCountryISO is NULL (all pre-international-auth legacy accounts) --
// MazadPay's only market before this feature. No Production backfill write is
// required for this: the fallback is applied at read time everywhere a user's
// effective market/currency is needed.
const DefaultAccountCountryISO = "MR"

// DefaultCurrencyCode is the runtime fallback currency for legacy users/
// auctions with no explicit currency_code (NULL) -- MazadPay's only currency
// (Mauritanian Ouguiya, current ISO-4217 code) before this feature.
const DefaultCurrencyCode = "MRU"

// EffectiveAccountCountryISO returns the user's account market, falling back
// to DefaultAccountCountryISO for legacy users with no explicit selection.
// Callers (auction/bid/wallet logic) must use this, never AccountCountryISO
// directly, so the NULL-legacy case is never silently mishandled.
func (u *User) EffectiveAccountCountryISO() string {
	if u.AccountCountryISO != nil && *u.AccountCountryISO != "" {
		return *u.AccountCountryISO
	}
	return DefaultAccountCountryISO
}

 func (u *User) MaskPhone() string {
	if len(u.Phone) < 4 {
		return "####"
	}
	return "####" + u.Phone[len(u.Phone)-4:]
}

type OTPVerification struct {
	ID          uuid.UUID  `db:"id"`
	Phone       string     `db:"phone"`
	Code        string     `db:"otp_code"`
	Purpose     string     `db:"purpose"`
	Attempts    int        `db:"attempts"`
	MaxAttempts int        `db:"max_attempts"`
	ExpiresAt   time.Time  `db:"expires_at"`
	VerifiedAt  *time.Time `db:"verified_at"`
	IPAddress   *string    `db:"ip_address"`
	CreatedAt   time.Time  `db:"created_at"`
}

type UserFavorite struct {
	UserID    uuid.UUID `db:"user_id"    json:"user_id"`
	AuctionID uuid.UUID `db:"auction_id" json:"auction_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

