package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/mazadpay/backend/internal/middleware"
	"github.com/mazadpay/backend/internal/services"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type AuthHandler struct {
	service  services.AuthService
	rdb      *redis.Client
	logger   *zap.Logger
	validate *validator.Validate
}

func NewAuthHandler(service services.AuthService, logger *zap.Logger, rdb *redis.Client) *AuthHandler {
	return &AuthHandler{
		service:  service,
		rdb:      rdb,
		logger:   logger,
		validate: validator.New(),
	}
}

type RegisterRequest struct {
	Phone       string `json:"phone"       validate:"required,min=8,max=20,numeric"`
	Pin         string `json:"pin"         validate:"required,len=4,numeric"`
	FullName    string `json:"full_name"   validate:"required,min=2,max=100"`
	Email       string `json:"email"       validate:"required,email"`
	City        string `json:"city"        validate:"omitempty,max=100"`
	CountryCode string `json:"country_code" validate:"omitempty,oneof=+222 +221 +212 +216"`
}

type LoginRequest struct {
	Phone string `json:"phone" validate:"required"`
	Pin   string `json:"pin"   validate:"required,len=4"`
}

type SendOTPRequest struct {
	Phone   string `json:"phone"   validate:"required"`
	Purpose string `json:"purpose" validate:"required,oneof=register reset_password"`
}

type VerifyOTPRequest struct {
	Phone   string `json:"phone"   validate:"required"`
	Code    string `json:"code"    validate:"required,min=4,max=6,numeric"`
	Purpose string `json:"purpose" validate:"required,oneof=register reset_password"`
}

type ResetPasswordRequest struct {
	Phone  string `json:"phone"   validate:"required"`
	Code   string `json:"code"    validate:"required,min=4,max=6,numeric"`
	NewPin string `json:"new_pin" validate:"required,len=4,numeric"`
}

type ChangePasswordRequest struct {
	OldPin string `json:"old_pin" validate:"required,len=4"`
	NewPin string `json:"new_pin" validate:"required,len=4,numeric"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	// Validate PIN strength
	if err := services.ValidatePINStrength(req.Pin); err != nil {
		return BadRequest(c, "PIN is too weak. Avoid repeating digits (1111) or sequences (1234)")
	}

	ip := c.IP()
	if err := h.service.Register(c.Context(), req.Phone, req.Pin, req.FullName, req.Email, req.City, req.CountryCode); err != nil {
		return MapError(c, h.logger, err)
	}

	// Automatically send OTP after registration
	if err := h.service.SendOTP(c.Context(), req.Phone, "register", ip); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Registration successful. OTP has been sent via SMS."})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	token, user, err := h.service.Login(c.Context(), req.Phone, req.Pin)
	if err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":             user.ID,
			"phone":          user.MaskPhone(),
			"role":           user.Role,
			"language":       user.LanguagePref,
			"is_super_admin": user.IsSuperAdmin,
		},
	})
}

func (h *AuthHandler) SendOTP(c *fiber.Ctx) error {
	var req SendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	ip := c.IP()
	if err := h.service.SendOTP(c.Context(), req.Phone, req.Purpose, ip); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "OTP sent successfully"})
}

func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	if err := h.service.VerifyOTP(c.Context(), req.Phone, req.Code, req.Purpose); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "OTP verified successfully"})
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	// Validate new PIN strength
	if err := services.ValidatePINStrength(req.NewPin); err != nil {
		return BadRequest(c, "New PIN is too weak. Avoid repeating digits (1111) or sequences (1234)")
	}

	// Verify OTP code
	if err := h.service.VerifyOTP(c.Context(), req.Phone, req.Code, "reset_password"); err != nil {
		return MapError(c, h.logger, err)
	}

	// Track password reset attempt
	ip := c.IP()
	if err := h.service.TrackPasswordReset(c.Context(), req.Phone, ip); err != nil {
		h.logger.Warn("failed to track password reset attempt", zap.Error(err))
		// Don't return error, continue with reset
	}

	if err := h.service.ResetPassword(c.Context(), req.Phone, req.NewPin); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Password reset successfully"})
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	// Validate new PIN strength
	if err := services.ValidatePINStrength(req.NewPin); err != nil {
		return BadRequest(c, "New PIN is too weak. Avoid repeating digits (1111) or sequences (1234)")
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		return BadRequest(c, "Invalid user ID in token")
	}

	if err := h.service.ChangePassword(c.Context(), userID, req.OldPin, req.NewPin); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Password changed successfully"})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Récupérer le token de l'entête Authorization
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return OK(c, fiber.Map{"message": "Logged out successfully (no token)"})
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == "" {
		return OK(c, fiber.Map{"message": "Logged out successfully (empty token)"})
	}

	// Ajouter le token à la blacklist Redis (expirer après 24h par sécurité)
	if h.rdb != nil {
		err := h.rdb.Set(c.Context(), fmt.Sprintf("blacklist:%s", tokenStr), "1", 24*time.Hour).Err()
		if err != nil {
			h.logger.Error("Failed to blacklist token", zap.Error(err))
		}
	}

	return OK(c, fiber.Map{"message": "Logged out successfully"})
}
