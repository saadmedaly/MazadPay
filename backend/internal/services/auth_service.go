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
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Verrouillage automatique après échecs de PIN successifs (Auth/OTP Rate Limiting
// Hardening) : 6 échecs en 15 minutes déclenchent un blocage temporaire de 15 minutes.
// Valeurs fixes (pas de variable d'env dédiée) car aucune n'existait déjà pour ce cas
// précis — cohérent avec la demande de ne pas introduire de config supplémentaire non
// nécessaire.
const (
	loginFailWindow    = 15 * time.Minute
	loginMaxFailures   = 6
	loginLockoutPeriod = 15 * time.Minute
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
	smsEnabled     bool // true only when SMS credentials are configured
	otpLength      int
	otpTTL         time.Duration
	rdb            *redis.Client
	logger         *zap.Logger
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, jwtExpiry int, env string, devOTP string, sms SMSService, otpLength int, otpTTLMinutes int, rdb *redis.Client, logger *zap.Logger) AuthService {
	// En production, devOTP doit être vide
	if env != "development" {
		devOTP = ""
	}
	if otpLength == 0 {
		otpLength = 4 // Default 4 digits
	}
	if otpTTLMinutes <= 0 {
		otpTTLMinutes = 5
	}
	// Detect if SMS credentials are actually configured
	smsEnabled := sms != nil && sms.IsConfigured()
	return &authService{
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		jwtExpiry:      jwtExpiry,
		env:            env,
		developmentOTP: devOTP,
		smsService:     sms,
		smsEnabled:     smsEnabled,
		otpLength:      otpLength,
		otpTTL:         time.Duration(otpTTLMinutes) * time.Minute,
		rdb:            rdb,
		logger:         logger,
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

	// Auto-verify user when SMS is not configured so login works immediately.
	// When Wablas credentials are set, user must verify via OTP first.
	isVerified := !s.smsEnabled

	user := &models.User{
		ID:           uuid.New(),
		Phone:        phone,
		PasswordHash: string(hash),
		FullName:     stringPtr(fullName),
		Email:        stringPtr(email),
		City:         stringPtr(city),
		CountryCode:  stringPtr(countryCode),
		LanguagePref: "ar",
		Role:         "user",
		IsActive:     true,
		IsVerified:   isVerified,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.userRepo.Create(ctx, user)
}

// loginFailKey retourne la clé Redis du compteur d'échecs de PIN pour un téléphone
// donné — distincte des clés de rate limiting par route (voir middleware/ratelimit.go).
func loginFailKey(phone string) string {
	return fmt.Sprintf("login_fail:%s", phone)
}

func (s *authService) Login(ctx context.Context, phone, pin string) (string, *models.User, error) {
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return "", nil, apperr.ErrUserNotFound
	}

	// Vérifier si le compte est actif
	if !user.IsActive {
		return "", nil, apperr.ErrAccountBlocked
	}

	// Vérifier le blocage temporaire (déjà actif, posé soit manuellement par un admin,
	// soit automatiquement ci-dessous après des échecs de PIN répétés)
	if user.BlockedUntil != nil && time.Now().Before(*user.BlockedUntil) {
		return "", nil, apperr.ErrAccountBlocked
	}

	if !user.IsVerified {
		return "", nil, apperr.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(pin)); err != nil {
		s.recordFailedLogin(ctx, phone)
		return "", nil, apperr.ErrInvalidPin
	}

	// PIN correct : réinitialiser le compteur d'échecs
	if s.rdb != nil {
		s.rdb.Del(ctx, loginFailKey(phone))
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

// recordFailedLogin incrémente le compteur d'échecs de PIN pour ce téléphone et pose
// un blocage temporaire (via BlockedUntil, déjà utilisé par le blocage manuel admin)
// une fois le seuil atteint — verrouillage automatique déclenché uniquement par un
// échec de PIN réel, jamais par le rate limiting de route (Auth/OTP Rate Limiting
// Hardening).
func (s *authService) recordFailedLogin(ctx context.Context, phone string) {
	if s.rdb == nil {
		return
	}
	key := loginFailKey(phone)
	count, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		if s.logger != nil {
			s.logger.Error("recordFailedLogin: Redis error", zap.Error(err))
		}
		return
	}
	if count == 1 {
		s.rdb.Expire(ctx, key, loginFailWindow)
	}
	if count >= loginMaxFailures {
		until := time.Now().Add(loginLockoutPeriod)
		if err := s.userRepo.SetBlockedUntil(ctx, phone, until); err != nil {
			if s.logger != nil {
				s.logger.Error("recordFailedLogin: failed to set BlockedUntil", zap.Error(err))
			}
			return
		}
		s.rdb.Del(ctx, key)
		if s.logger != nil {
			s.logger.Warn("Account auto-locked after repeated failed PIN attempts",
				zap.String("phone", phone),
				zap.Int64("attempts", count),
			)
		}
	}
}

func (s *authService) SendOTP(ctx context.Context, phone, purpose, ip string) error {
	code := GenerateOTP(s.otpLength)

	// Vérifier s'il existe déjà un OTP en cours de validité
	existingOTP, _ := s.userRepo.FindLatestOTP(ctx, phone, purpose)
	if existingOTP != nil && time.Now().Before(existingOTP.ExpiresAt) {
		// Un OTP valide existe déjà, retourner un message clair
		return apperr.ErrOTPRateLimited
	}

	// Save new OTP to database for verification
	otp := &models.OTPVerification{
		ID:          uuid.New(),
		Phone:       phone,
		Code:   code,
		Purpose:     purpose,
		Attempts:    0,
		MaxAttempts: 3,
		ExpiresAt:   time.Now().Add(s.otpTTL),
		IPAddress:   &ip,
	}

	if err := s.userRepo.CreateOTP(ctx, otp); err != nil {
		return err
	}

	// Send OTP via WhatsApp (Wablas)
	if s.smsService != nil {
		if err := s.smsService.SendOTP(phone, code); err != nil {
			if s.env == "development" {
				fmt.Printf("DEBUG: WhatsApp OTP sending failed in development (%v). Continuing because development OTP is allowed.\n", err)
				return nil
			}
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
	} else if code != otp.Code {
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

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
