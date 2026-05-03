package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, phone, pin, fullName, email, city, countryCode string) error
	Login(ctx context.Context, phone, pin string) (string, *models.User, error) // token, user, err
	SendOTP(ctx context.Context, phone, purpose, ip string) error
	VerifyOTP(ctx context.Context, phone, code, purpose string) error
	ResetPassword(ctx context.Context, phone, newPin string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPin, newPin string) error
	GenerateJWT(userID uuid.UUID, role string, isSuperAdmin bool) (string, error)
	ValidateJWT(tokenString string) (*JWTClaims, error)
	TrackPasswordReset(ctx context.Context, phone, ip string) error
}

type JWTClaims struct {
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	jwt.RegisteredClaims
}

type authService struct {
	userRepo       repository.UserRepository
	jwtSecret      string
	jwtExpiry      int
	env            string
	developmentOTP string // Code OTP de développement
	smsService     SMSService
	otpLength      int
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, jwtExpiry int, env string, devOTP string, sms SMSService, otpLength int) AuthService {
	// En production, devOTP doit être vide
	if env != "development" {
		devOTP = ""
	}
	if otpLength == 0 {
		otpLength = 4 // Default 4 digits
	}
	return &authService{
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		jwtExpiry:      jwtExpiry,
		env:            env,
		developmentOTP: devOTP,
		smsService:     sms,
		otpLength:      otpLength,
	}
}

// ValidCountryCodes liste les codes pays supportés
var ValidCountryCodes = map[string]string{
	"+222": "MR", // Mauritanie
	"+221": "SN", // Sénégal
	"+212": "MA", // Maroc
	"+216": "TN", // Tunisie
}

func (s *authService) Register(ctx context.Context, phone, pin, fullName, email, city, countryCode string) error {
	// Vérifier si le code pays est valide
	if countryCode == "" {
		countryCode = "+222" // Default: Mauritanie
	}
	if _, valid := ValidCountryCodes[countryCode]; !valid {
		return fmt.Errorf("code pays non supporté: %s (supportés: +222, +221, +212, +216)", countryCode)
	}

	// Vérifier si le numéro existe déjà
	existing, _ := s.userRepo.FindByPhone(ctx, phone)
	if existing != nil {
		return apperr.ErrDuplicatePhone
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		ID:           uuid.New(),
		Phone:        phone,
		PasswordHash: string(hash),
		FullName:     &fullName,
		Email:        &email,
		City:         &city,
		CountryCode:  &countryCode,
		LanguagePref: "ar",
		Role:         "user",
		IsVerified:   false,
	}

	return s.userRepo.Create(ctx, user)
}

func (s *authService) Login(ctx context.Context, phone, pin string) (string, *models.User, error) {
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return "", nil, apperr.ErrUnauthorized
	}

	// Vérifier si le compte est actif
	if !user.IsActive {
		return "", nil, apperr.ErrAccountBlocked
	}

	// Vérifier le blocage temporaire
	if user.BlockedUntil != nil && time.Now().Before(*user.BlockedUntil) {
		return "", nil, apperr.ErrAccountBlocked
	}

	if !user.IsVerified {
		return "", nil, apperr.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(pin)); err != nil {
		return "", nil, apperr.ErrInvalidPin
	}

	token, err := s.GenerateJWT(user.ID, user.Role, user.IsSuperAdmin)
	if err != nil {
		return "", nil, err
	}

	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log l'erreur mais ne bloque pas la connexion
		fmt.Println("Warning: failed to update last login for user", user.ID, ":", err)
	}
	return token, user, nil
}

func (s *authService) SendOTP(ctx context.Context, phone, purpose, ip string) error {
	code := GenerateOTP(s.otpLength)

	// Vérifier s'il existe déjà un OTP en cours de validité
	existingOTP, _ := s.userRepo.FindLatestOTP(ctx, phone, purpose)
	if existingOTP != nil && time.Now().Before(existingOTP.ExpiresAt) {
		// Un OTP valide existe déjà, retourner un message clair
		return fmt.Errorf("an OTP verification code is already active for this phone number. Please wait before requesting a new one")
	}

	// Save new OTP to database for verification
	otp := &models.OTPVerification{
		ID:          uuid.New(),
		Phone:       phone,
		TwilioSid:   code,
		Purpose:     purpose,
		Attempts:    0,
		MaxAttempts: 3,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		IPAddress:   &ip,
	}

	if err := s.userRepo.CreateOTP(ctx, otp); err != nil {
		return err
	}

	// Send SMS via Twilio
	if s.smsService != nil {
		if err := s.smsService.SendOTP(phone, code); err != nil {
			return err
		}
	}

	return nil
}

func (s *authService) VerifyOTP(ctx context.Context, phone, code, purpose string) error {
	otp, err := s.userRepo.FindLatestOTP(ctx, phone, purpose)
	if err != nil {
		return apperr.ErrOTPInvalid
	}

	if time.Now().After(otp.ExpiresAt) {
		return apperr.ErrOTPExpired
	}

	if otp.Attempts >= otp.MaxAttempts {
		return apperr.ErrOTPMaxAttempts
	}

	// Vérifier le code OTP
	// En développement, accepter le code configuré
	// En production, rejeter les codes de développement
	if s.env == "development" && s.developmentOTP != "" && code == s.developmentOTP {
		// OK - Mode développement avec code autorisé
	} else if code != otp.TwilioSid {
		// En production ou si code incorrect, incrémenter les tentatives
		if err := s.userRepo.IncrementOTPAttempts(ctx, otp.ID); err != nil {
			return err
		}
		return apperr.ErrOTPInvalid
	}

	if err := s.userRepo.MarkOTPVerified(ctx, otp.ID); err != nil {
		return err
	}

	if purpose == "register" {
		if err := s.userRepo.SetVerified(ctx, phone); err != nil {
			return err
		}
	}

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, phone, newPin string) error {
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return apperr.ErrNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePin(ctx, user.ID, string(hash))
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPin, newPin string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return apperr.ErrNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPin)); err != nil {
		return apperr.ErrInvalidPin
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePin(ctx, userID, string(hash))
}

func (s *authService) GenerateJWT(userID uuid.UUID, role string, isSuperAdmin bool) (string, error) {
	claims := JWTClaims{
		UserID:       userID.String(),
		Role:         role,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *authService) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// TrackPasswordReset enregistre une tentative de réinitialisation de mot de passe
// pour détecter les abus (brute force attempts)
func (s *authService) TrackPasswordReset(ctx context.Context, phone, ip string) error {
	return s.userRepo.LogPasswordResetAttempt(ctx, phone, ip)
}
