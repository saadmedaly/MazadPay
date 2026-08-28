package handlers

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	apperr "github.com/mazadpay/backend/internal/errors"
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

// RegisterRequest est le contrat v2 (/v2/api/auth/register) — STRICT, à utiliser
// exclusivement par la nouvelle version de l'app. country_iso est requis dès la
// validation ici : un appelant ne peut jamais retomber sur le comportement legacy (4
// pays, PIN 4 chiffres) en omettant simplement le champ.
type RegisterRequest struct {
	Phone string `json:"phone"       validate:"required,min=8,max=20"`
	// Pin (Password Strength Phase 1) : mot de passe d'au moins 8 caractères. 72 = limite
	// intrinsèque de bcrypt (tronque silencieusement au-delà, donc on refuse
	// explicitement plutôt que de tronquer sans le dire).
	Pin      string `json:"pin"         validate:"required,min=8,max=72"`
	FullName string `json:"full_name"   validate:"required,min=2,max=100"`
	Email    string `json:"email"       validate:"omitempty,email"`
	City     string `json:"city"        validate:"omitempty,max=100"`
	// CountryISO (International Auth Phase 2) : code région ISO-2 REQUIS (ex: "MR",
	// "TN", "US", "CA") — un indicatif d'appel seul (ex: "+1") ne suffit pas à identifier
	// un pays, plusieurs pays le partagent. Validation réelle du numéro pour cette région
	// précise via libphonenumber dans AuthService.Register (NormalizeE164 +
	// IsValidNumberForRegion). `required,len=2` : impossible de contourner la politique
	// v2 en omettant ce champ (API Versioning Phase 1 — voir RegisterRequestLegacy pour
	// le contrat v1 séparé, qui n'a PAS ce champ du tout).
	CountryISO string `json:"country_iso" validate:"required,len=2"`
}

// RegisterRequestLegacy est le contrat v1 (/v1/api/auth/register) — reproduit EXACTEMENT
// le contrat publié sur Google Play avant cette fonctionnalité (voir git history,
// baseline avant "International Auth"). NE JAMAIS AJOUTER de nouveau champ optionnel qui
// élargirait silencieusement ce contrat — toute nouvelle capacité va dans RegisterRequest
// (v2). country_code reste un indicatif d'appel (pas ISO-2), restreint aux 4 pays
// historiques (validé dans AuthService.RegisterLegacy). v1 est TEMPORAIRE : à retirer une
// fois l'ancienne version de l'app suffisamment désinstallée/mise à jour (voir
// AuthService, commentaire d'interface).
type RegisterRequestLegacy struct {
	Phone       string `json:"phone"       validate:"required,min=8,max=20"`
	Pin         string `json:"pin"         validate:"required,len=4,numeric"`
	FullName    string `json:"full_name"   validate:"required,min=2,max=100"`
	Email       string `json:"email"       validate:"omitempty,email"`
	City        string `json:"city"        validate:"omitempty,max=100"`
	CountryCode string `json:"country_code" validate:"omitempty,oneof=+222 +221 +212 +216"`
}

type LoginRequest struct {
	Phone string `json:"phone" validate:"required"`
	// CountryISO est optionnel ici (contrairement à l'inscription) : un client non
	// encore mis à jour peut omettre ce champ, auquel cas Login retombe sur la
	// correspondance exacte historique de 'phone' (pont de compatibilité — voir
	// AuthService.Login). Un client à jour doit toujours l'envoyer.
	CountryISO string `json:"country_iso" validate:"omitempty,len=2"`
	// Pin : volontairement pas de contrainte de longueur ici (juste "required") —
	// Login doit accepter aussi bien un ancien PIN à 4 chiffres (comptes existants,
	// jamais migrés de force) qu'un nouveau mot de passe de 8+ caractères (comptes créés
	// depuis Password Strength Phase 1). bcrypt.CompareHashAndPassword ne se soucie pas
	// de la longueur d'entrée, donc rien n'empêche cette souplesse ici.
	Pin string `json:"pin"   validate:"required"`
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

// ResetPasswordRequest est le contrat v2 (/v2/api/auth/reset-password) — impose la
// politique de mot de passe fort même sur un compte historique qui se réinitialise :
// une fois réinitialisé via v2, il n'a plus de PIN faible à 4 chiffres.
type ResetPasswordRequest struct {
	Phone  string `json:"phone"   validate:"required"`
	Code   string `json:"code"    validate:"required,min=4,max=6,numeric"`
	NewPin string `json:"new_pin" validate:"required,min=8,max=72"`
}

// ResetPasswordRequestLegacy est le contrat v1 (/v1/api/auth/reset-password) —
// reproduit EXACTEMENT le contrat publié : new_pin reste un PIN 4 chiffres numériques.
// Le service sous-jacent (AuthService.ResetPassword) est PARTAGÉ entre v1 et v2 — sa
// logique ne dépend pas de la longueur du mot de passe, seule cette validation DTO en
// amont diffère (voir AuthService, commentaire d'interface).
type ResetPasswordRequestLegacy struct {
	Phone  string `json:"phone"   validate:"required"`
	Code   string `json:"code"    validate:"required,min=4,max=6,numeric"`
	NewPin string `json:"new_pin" validate:"required,len=4,numeric"`
}

type ChangePasswordRequest struct {
	// OldPin : pas de contrainte de longueur — doit accepter un ancien PIN à 4 chiffres
	// (voir LoginRequest.Pin, même raisonnement) puisqu'on vérifie seulement l'identité
	// avant d'appliquer NewPin.
	OldPin string `json:"old_pin" validate:"required"`
	// NewPin : même politique 8+ caractères que ResetPasswordRequest.NewPin — changer
	// son mot de passe doit toujours produire un mot de passe fort, jamais un retour à
	// un PIN à 4 chiffres.
	NewPin string `json:"new_pin" validate:"required,min=8,max=72"`
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

	// ip := c.IP()
	if err := h.service.Register(c.Context(), req.Phone, req.Pin, req.FullName, req.Email, req.City, req.CountryISO); err != nil {
		return MapError(c, h.logger, err)
	}

	ip := c.IP()
	// Automatically send OTP after registration.
	// If SMS/WhatsApp service is unavailable, registration still succeeds.
	// User will be marked verified automatically when SMS is disabled.
	if err := h.service.SendOTP(c.Context(), req.Phone, "register", ip); err != nil {
		h.logger.Warn("[Register] OTP sending failed — registration still succeeds (SMS may be disabled)",
			zap.String("phone", req.Phone),
			zap.Error(err),
		)
	}

	return OK(c, fiber.Map{"message": "Registration successful. Please verify your phone with the code sent via WhatsApp."})
}

// RegisterLegacy (API Versioning Phase 1, /v1/api/auth/register) reproduit EXACTEMENT le
// comportement publié avant cette fonctionnalité (4 pays fermés, PIN 4 chiffres) — voir
// AuthService.RegisterLegacy et RegisterRequestLegacy pour le détail des garanties. Ne
// jamais faire converger cette fonction avec Register (v2) : la version déjà publiée sur
// Google Play dépend de ce comportement exact restant inchangé indéfiniment tant qu'elle
// est encore installée.
func (h *AuthHandler) RegisterLegacy(c *fiber.Ctx) error {
	var req RegisterRequestLegacy
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}

	if err := services.ValidatePINStrength(req.Pin); err != nil {
		return BadRequest(c, "PIN is too weak. Avoid repeating digits (1111) or sequences (1234)")
	}

	if err := h.service.RegisterLegacy(c.Context(), req.Phone, req.Pin, req.FullName, req.Email, req.City, req.CountryCode); err != nil {
		return MapError(c, h.logger, err)
	}

	ip := c.IP()
	if err := h.service.SendOTP(c.Context(), req.Phone, "register", ip); err != nil {
		h.logger.Warn("[RegisterLegacy] OTP sending failed — registration still succeeds (SMS may be disabled)",
			zap.String("phone", req.Phone),
			zap.Error(err),
		)
	}

	return OK(c, fiber.Map{"message": "Registration successful. Please verify your phone with the code sent via WhatsApp."})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}

	token, user, err := h.service.Login(c.Context(), req.Phone, req.CountryISO, req.Pin)
	if err != nil {
		// Réponse identique que le PIN soit incorrect ou que le numéro ne soit pas
		// enregistré (OTP Security Phase 1B-1 — durcissement anti-énumération) :
		// auparavant, invalid_pin/phone_not_registered + can_reset_password/
		// can_register permettaient de déterminer si un numéro possède un compte.
		// Aucun code mobile/web ne dépend de ces champs (vérifié avant ce changement),
		// donc aucune adaptation client n'est requise pour ce correctif.
		if errors.Is(err, apperr.ErrInvalidPin) || errors.Is(err, apperr.ErrUserNotFound) || errors.Is(err, apperr.ErrPhoneUnavailable) {
			return Fail(c, fiber.StatusUnauthorized, "invalid_credentials", "Invalid phone or PIN")
		}
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
	return h.resetPasswordCommon(c, req.Phone, req.Code, req.NewPin)
}

// ResetPasswordLegacy (API Versioning Phase 1, /v1/api/auth/reset-password) accepte un
// PIN 4 chiffres (contrat publié) au lieu du mot de passe 8-72 caractères de v2. Le
// service sous-jacent est identique — voir AuthService.ResetPassword et
// ResetPasswordRequestLegacy.
func (h *AuthHandler) ResetPasswordLegacy(c *fiber.Ctx) error {
	var req ResetPasswordRequestLegacy
	if err := c.BodyParser(&req); err != nil {
		return BadRequest(c, "Invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return BadRequest(c, err.Error())
	}
	return h.resetPasswordCommon(c, req.Phone, req.Code, req.NewPin)
}

// resetPasswordCommon factorise la logique partagée par ResetPassword (v2) et
// ResetPasswordLegacy (v1) : seule la validation de la longueur/format de newPin diffère
// entre les deux DTO appelants, tout le reste (tracking, message neutre anti-énumération,
// gestion des erreurs) est strictement identique.
func (h *AuthHandler) resetPasswordCommon(c *fiber.Ctx, phone, code, newPin string) error {
	// Validate new PIN strength
	if err := services.ValidatePINStrength(newPin); err != nil {
		return BadRequest(c, "New PIN is too weak. Avoid repeating digits (1111) or sequences (1234)")
	}

	// Track password reset attempt
	ip := c.IP()
	if err := h.service.TrackPasswordReset(c.Context(), phone, ip); err != nil {
		h.logger.Warn("failed to track password reset attempt", zap.Error(err))
		// Don't return error, continue with reset
	}

	// Message neutre unique, identique que le compte existe ou non (OTP Security Phase
	// 1B-2 — durcissement anti-énumération). Un texte différent selon le cas ("Password
	// reset successfully" vs un message conditionnel) permettait de deviner l'existence
	// du compte même à statut HTTP 200 identique — corrigé ici en unifiant aussi le
	// texte, pas seulement le code de statut. Ne jamais confirmer/infirmer l'existence
	// du compte par le contenu de la réponse.
	const neutralResetMessage = "If this phone is registered, the PIN has been reset successfully."

	// La vérification OTP ("reset_password") est désormais effectuée à l'intérieur de
	// ResetPassword elle-même (OTP Security Phase 1C) — le handler ne fait plus que
	// relayer code, il n'est plus seul responsable d'imposer la preuve OTP.
	if err := h.service.ResetPassword(c.Context(), phone, code, newPin); err != nil {
		// apperr.ErrNotFound (numéro sans compte, après un OTP pourtant valide) reçoit
		// la même réponse 200 + même message que le succès réel — voir neutralResetMessage
		// ci-dessus. Toute autre erreur (OTP invalide/expiré/MaxAttempts, PIN faible,
		// etc.) garde son traitement normal via MapError, inchangé.
		if errors.Is(err, apperr.ErrNotFound) {
			return OK(c, fiber.Map{"message": neutralResetMessage})
		}
		return MapError(c, h.logger, err)
	}

	return OK(c, fiber.Map{"message": neutralResetMessage})
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

	if h.rdb == nil {
		// Sans Redis, aucune blacklist n'est possible : le token reste valide malgré
		// la demande de logout. Le signaler explicitement plutôt que de répondre
		// succès (Session Security Phase 1 — ne jamais faire croire à un logout
		// effectif quand il ne l'est pas).
		h.logger.Error("Logout failed: Redis unavailable, cannot blacklist token")
		return c.Status(503).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "service_unavailable", "message": "تعذر تسجيل الخروج حالياً، حاول لاحقاً"},
		})
	}

	// TTL du blacklist aligné sur la durée de vie réelle restante du token (Session
	// Security Phase 1) : auparavant fixé à 24h alors que JWT_EXPIRY_HOURS peut aller
	// jusqu'à 72h, laissant une fenêtre où le token redevenait utilisable après
	// expiration prématurée de l'entrée Redis mais avant l'expiration réelle du JWT.
	ttl := 24 * time.Hour
	claims, parseErr := h.service.ValidateJWT(tokenStr)
	if parseErr == nil && claims.ExpiresAt != nil {
		remaining := time.Until(claims.ExpiresAt.Time)
		if remaining <= 0 {
			// Déjà expiré naturellement : rien à blacklister, le middleware JWT
			// rejettera ce token via la vérification exp standard.
			return OK(c, fiber.Map{"message": "Logged out successfully"})
		}
		ttl = remaining
	}
	// Si le token est invalide/non parseable (signature altérée, format inattendu),
	// on garde le TTL par défaut de 24h par prudence plutôt que de ne pas blacklister
	// du tout — un token illisible ici ne doit pas empêcher un logout best-effort.

	if err := h.rdb.Set(c.Context(), fmt.Sprintf("blacklist:%s", tokenStr), "1", ttl).Err(); err != nil {
		h.logger.Error("Logout failed: could not write blacklist entry", zap.Error(err))
		return c.Status(503).JSON(fiber.Map{
			"success": false,
			"error":   fiber.Map{"code": "service_unavailable", "message": "تعذر تسجيل الخروج حالياً، حاول لاحقاً"},
		})
	}

	return OK(c, fiber.Map{"message": "Logged out successfully"})
}
