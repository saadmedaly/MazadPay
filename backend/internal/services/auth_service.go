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
	Register(ctx context.Context, phone, pin, fullName, email, city string) error
	Login(ctx context.Context, phone, pin string) (string, *models.User, error) // token, user, err
	SendOTP(ctx context.Context, phone, purpose, ip string) error
	VerifyOTP(ctx context.Context, phone, code, purpose string) error
	ResetPassword(ctx context.Context, phone, newPin string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPin, newPin string) error
	GenerateJWT(userID uuid.UUID, role string) (string, error)
}

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type authService struct {
	userRepo       repository.UserRepository
	jwtSecret      string
	jwtExpiry      int
	env            string
	developmentOTP string // Code OTP de développement
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string, jwtExpiry int, env string, devOTP string) AuthService {
	// En production, devOTP doit être vide
	if env != "development" {
		devOTP = ""
	}
	return &authService{userRepo: userRepo, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry, env: env, developmentOTP: devOTP}
}

func (s *authService) Register(ctx context.Context, phone, pin, fullName, email, city string) error {
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

	// Vérifier le blocage temporaire (pourrait être ajouté à models.User via STEP3)
	// if user.BlockedUntil != nil && time.Now().Before(*user.BlockedUntil) {
	//     return "", nil, apperr.ErrAccountBlocked
	// }

	if !user.IsVerified {
		return "", nil, apperr.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(pin)); err != nil {
		return "", nil, apperr.ErrInvalidPin
	}

	token, err := s.GenerateJWT(user.ID, user.Role)
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
	// STUB : En développement, le code OTP est "123456".
	// À remplacer par l'intégration Termii dans une étape dédiée.

	otp := &models.OTPVerification{
		ID:          uuid.New(),
		Phone:       phone,
		TermiiPinID: fmt.Sprintf("stub-%s-%d", phone, time.Now().Unix()), // sera remplacé par Termii
		Purpose:     purpose,
		Attempts:    0,
		MaxAttempts: 3,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		IPAddress:   &ip,
	}

	return s.userRepo.CreateOTP(ctx, otp)
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
	} else if code != otp.TermiiPinID {
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

func (s *authService) GenerateJWT(userID uuid.UUID, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID.String(),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
