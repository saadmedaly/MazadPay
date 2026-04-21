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
}

// MaskPhone retourne le numéro masqué (####4709)
func (u *User) MaskPhone() string {
	if len(u.Phone) < 4 {
		return "####"
	}
	return "####" + u.Phone[len(u.Phone)-4:]
}

type OTPVerification struct {
	ID          uuid.UUID  `db:"id"`
	Phone       string     `db:"phone"`
	TwilioSid   string     `db:"twilio_sid"`
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

