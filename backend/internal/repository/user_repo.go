package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	FindByPhone(ctx context.Context, phone string) (*models.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	SetVerified(ctx context.Context, phone string) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdateProfile(ctx context.Context, id uuid.UUID, fullName, email, city string) error
	UpdateProfileExtended(ctx context.Context, id uuid.UUID, fullName, email, city, countryCode, address, postalCode, dateOfBirth, gender string) error
	UpdateProfilePic(ctx context.Context, id uuid.UUID, url string) error
	UpdatePin(ctx context.Context, id uuid.UUID, hash string) error
	UpdateLanguage(ctx context.Context, id uuid.UUID, lang string) error
	UpdateNotificationSettings(ctx context.Context, id uuid.UUID, enabled bool) error
	SetBlockedUntil(ctx context.Context, phone string, until time.Time) error

	// OTP
	CreateOTP(ctx context.Context, otp *models.OTPVerification) error
	FindLatestOTP(ctx context.Context, phone, purpose string) (*models.OTPVerification, error)
	IncrementOTPAttempts(ctx context.Context, id uuid.UUID) error
	MarkOTPVerified(ctx context.Context, id uuid.UUID) error
	CleanExpiredOTPs(ctx context.Context) error

	// Password Reset
	LogPasswordResetAttempt(ctx context.Context, phone, ip string) error
	SeedDefaultSuperAdmin(ctx context.Context, phone, pin, fullName, email string) error

	// Admin
	ListPaginated(ctx context.Context, page, perPage int, query string) ([]models.User, int, error)
	Count(ctx context.Context, query string) (int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetStats(ctx context.Context) (int, int, error) // Total, Verified
	PromoteToAdmin(ctx context.Context, id uuid.UUID, fullName, email, hash string) error
	FindAllAdmins(ctx context.Context) ([]models.User, error)

	// New methods for extended functionality
	UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status string) error
	GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error)
	UpdateUserSettings(ctx context.Context, userID uuid.UUID, settings interface{}) error
}

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE phone = $1 AND is_active = true`, phone)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return &u, nil
}

func (r *userRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return &u, nil
}

func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO users (id, phone, password_hash, full_name, email, language_pref, role, is_verified)
		VALUES (:id, :phone, :password_hash, :full_name, :email, :language_pref, :role, :is_verified)
	`, user)
	return err
}

func (r *userRepo) SetVerified(ctx context.Context, phone string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET is_verified = true WHERE phone = $1`, phone)
	return err
}

func (r *userRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET last_login_at = now() WHERE id = $1`, id)
	return err
}

func (r *userRepo) CreateOTP(ctx context.Context, otp *models.OTPVerification) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO otp_verifications 
			(id, phone, twilio_sid, purpose, attempts, max_attempts, expires_at, ip_address)
		VALUES 
			(:id, :phone, :twilio_sid, :purpose, :attempts, :max_attempts, :expires_at, :ip_address)
	`, otp)
	return err
}

func (r *userRepo) FindLatestOTP(ctx context.Context, phone, purpose string) (*models.OTPVerification, error) {
	var otp models.OTPVerification
	err := r.db.GetContext(ctx, &otp, `
		SELECT * FROM otp_verifications 
		WHERE phone = $1 AND purpose = $2 AND verified_at IS NULL
		ORDER BY created_at DESC 
		LIMIT 1
	`, phone, purpose)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return &otp, nil
}

func (r *userRepo) IncrementOTPAttempts(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE otp_verifications SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

func (r *userRepo) MarkOTPVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE otp_verifications SET verified_at = now() WHERE id = $1`, id)
	return err
}

func (r *userRepo) CleanExpiredOTPs(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM otp_verifications WHERE expires_at < now()`)
	return err
}

func (r *userRepo) UpdatePin(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2`, hash, id)
	return err
}

func (r *userRepo) SetBlockedUntil(ctx context.Context, phone string, until time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET blocked_until = $1 WHERE phone = $2`, until, phone)
	return err
}

func (r *userRepo) UpdateProfile(ctx context.Context, id uuid.UUID, fullName, email, city string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET full_name=$1, email=$2, city=$3 WHERE id=$4`,
		fullName, email, city, id)
	return err
}

func (r *userRepo) UpdateProfilePic(ctx context.Context, id uuid.UUID, url string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET profile_pic_url=$1 WHERE id=$2`, url, id)
	return err
}

func (r *userRepo) UpdateLanguage(ctx context.Context, id uuid.UUID, lang string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET language_pref=$1 WHERE id=$2`, lang, id)
	return err
}

func (r *userRepo) UpdateNotificationSettings(ctx context.Context, id uuid.UUID, enabled bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET notifications_enabled=$1 WHERE id=$2`, enabled, id)
	return err
}

func (r *userRepo) ListPaginated(ctx context.Context, page, perPage int, query string) ([]models.User, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if query != "" {
		where += " AND (phone ILIKE $1 OR full_name ILIKE $1 OR email ILIKE $1)"
		args = append(args, "%"+query+"%")
	}

	count, err := r.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	q := fmt.Sprintf("SELECT * FROM users %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		where, len(args)+1, len(args)+2)

	listArgs := append(args, perPage, offset)
	users := make([]models.User, 0)
	err = r.db.SelectContext(ctx, &users, q, listArgs...)
	return users, count, err
}

func (r *userRepo) Count(ctx context.Context, query string) (int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if query != "" {
		where += " AND (phone ILIKE $1 OR full_name ILIKE $1 OR email ILIKE $1)"
		args = append(args, "%"+query+"%")
	}

	var count int
	err := r.db.GetContext(ctx, &count, fmt.Sprintf("SELECT COUNT(*) FROM users %s", where), args...)
	return count, err
}

func (r *userRepo) UpdateStatus(ctx context.Context, id uuid.UUID, isActive bool) error {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET is_active = $1 WHERE id = $2", isActive, id)
	return err
}

func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM bids WHERE user_id = $1", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM transactions WHERE user_id = $1", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM notifications WHERE user_id = $1", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM push_tokens WHERE user_id = $1", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM wallet_holds WHERE user_id = $1", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM app_ratings WHERE user_id = $1", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM kyc_verifications WHERE user_id = $1", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM reports WHERE reporter_id = $1", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM system_settings WHERE updated_by = $1", id)
	if err != nil {
		return err
	}
	// Delete auctions where seller_id = id
	_, err = tx.ExecContext(ctx, "DELETE FROM auctions WHERE seller_id = $1", id)
	if err != nil {
		return err
	}
	// Delete wallet
	_, err = tx.ExecContext(ctx, "DELETE FROM wallets WHERE user_id = $1", id)
	if err != nil {
		return err
	}
	// Finally delete user
	_, err = tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *userRepo) GetStats(ctx context.Context) (int, int, error) {
	var total, verified int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM users")
	if err != nil {
		return 0, 0, err
	}
	err = r.db.GetContext(ctx, &verified, "SELECT COUNT(*) FROM users WHERE is_verified = true")
	return total, verified, err
}

func (r *userRepo) PromoteToAdmin(ctx context.Context, id uuid.UUID, fullName, email, hash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET role = 'admin', is_verified = true, full_name = $1, email = $2, password_hash = $3 WHERE id = $4`,
		fullName, email, hash, id)
	return err
}

func (r *userRepo) FindAllAdmins(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := r.db.SelectContext(ctx, &users, "SELECT * FROM users WHERE role = 'admin' AND is_active = true")
	return users, err
}

func (r *userRepo) LogPasswordResetAttempt(ctx context.Context, phone, ip string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO password_reset_attempts (id, phone, ip_address, success, created_at)
		VALUES ($1, $2, $3, true, now())
	`, uuid.New(), phone, ip)
	return err
}

// SeedDefaultSuperAdmin crée un super admin par défaut s'il n'existe pas
func (r *userRepo) SeedDefaultSuperAdmin(ctx context.Context, phone, pin, fullName, email string) error {
	if phone == "" || pin == "" {
		return nil // Skip si pas configuré
	}

	// Vérifier si l'utilisateur existe déjà
	existing, _ := r.FindByPhone(ctx, phone)
	if existing != nil {
		// Mettre à jour en super admin si nécessaire
		if !existing.IsSuperAdmin {
			_, err := r.db.ExecContext(ctx, `
				UPDATE users SET is_super_admin = true WHERE phone = $1
			`, phone)
			return err
		}
		return nil // Déjà existe et est super admin
	}

	// Créer le super admin
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO users (
			id, phone, password_hash, full_name, email, language_pref, 
			role, is_verified, is_active, is_super_admin, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, uuid.New(), phone, string(hash), fullName, email, "ar", "admin", true, true, true, now, now)

	return err
}

// New methods implementations for extended user functionality

func (r *userRepo) UpdateProfileExtended(ctx context.Context, id uuid.UUID, fullName, email, city, countryCode, address, postalCode, dateOfBirth, gender string) error {
	query := `
		UPDATE users SET
			full_name = $1,
			email = $2,
			city = $3,
			country_code = $4,
			address = $5,
			postal_code = $6,
			date_of_birth = $7,
			gender = $8,
			updated_at = now()
		WHERE id = $9
	`
	var dob *time.Time
	if dateOfBirth != "" {
		// Try different date formats
		parsedDob, err := time.Parse("2006-01-02", dateOfBirth)
		if err != nil {
			// Try RFC3339 format (ISO format from frontend)
			parsedDob, err = time.Parse(time.RFC3339, dateOfBirth)
			if err != nil {
				// Try with just date part
				parsedDob, err = time.Parse("2006-01-02T15:04:05Z", dateOfBirth)
			}
		}
		if err == nil {
			dob = &parsedDob
		}
	}

	_, err := r.db.ExecContext(ctx, query,
		nullString(fullName),
		nullString(email),
		nullString(city),
		nullString(countryCode),
		nullString(address),
		nullString(postalCode),
		dob,
		nullString(gender),
		id,
	)
	return err
}

func (r *userRepo) UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET kyc_status = $1, updated_at = now() WHERE id = $2",
		status, userID,
	)
	return err
}

func (r *userRepo) GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	var settings models.UserSettings
	err := r.db.GetContext(ctx, &settings,
		"SELECT * FROM user_settings WHERE user_id = $1", userID,
	)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return &settings, nil
}

func (r *userRepo) UpdateUserSettings(ctx context.Context, userID uuid.UUID, settings interface{}) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE user_settings SET updated_at = now() WHERE user_id = $1`,
		userID,
	)
	return err
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
