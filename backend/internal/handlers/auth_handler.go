package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

type AuthHandler struct {
	service  services.AuthService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewAuthHandler(service services.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}


type RegisterRequest struct {
	Phone string `json:"phone" validate:"required,min=8,max=20"`
	Pin   string `json:"pin"   validate:"required,len=4,numeric"`
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
	Code    string `json:"code"    validate:"required,len=6"`
	Purpose string `json:"purpose" validate:"required,oneof=register reset_password"`
}

type ResetPasswordRequest struct {
	Phone  string `json:"phone"   validate:"required"`
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

	if err := h.service.Register(c.Context(), req.Phone, req.Pin); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Registration successful. OTP will be sent."})
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
			"id":       user.ID,
			"phone":    user.MaskPhone(),
			"role":     user.Role,
			"language": user.LanguagePref,
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

	userIDStr := GetUserID(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return BadRequest(c, "Invalid user ID in token")
	}

	if err := h.service.ChangePassword(c.Context(), userID, req.OldPin, req.NewPin); err != nil {
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": "Password changed successfully"})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Avec JWT stateless, le logout côté serveur peut être fait via blacklist dans Redis.
	// Pour cette étape, on renvoie juste un succès, le client supprime son token.
	return OK(c, fiber.Map{"message": "Logged out successfully"})
}
