package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func OK(c *fiber.Ctx, data interface{}) error {

	return c.JSON(fiber.Map{"success": true, "data": data})
}

func PaginatedOK(c *fiber.Ctx, data interface{}, meta interface{}) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
		"meta":    meta,
	})
}

func Created(c *fiber.Ctx, data interface{}) error {

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": data})
}

func Fail(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   fiber.Map{"code": code, "message": message},
	})
}

func BadRequest(c *fiber.Ctx, message string) error {
	return Fail(c, 400, "bad_request", message)
}

func Unauthorized(c *fiber.Ctx, message ...string) error {
	msg := "Authentication required"
	if len(message) > 0 {
		msg = message[0]
	}
	return Fail(c, 401, "unauthorized", msg)
}

func Forbidden(c *fiber.Ctx, message ...string) error {
	msg := "Access denied"
	if len(message) > 0 {
		msg = message[0]
	}
	return Fail(c, 403, "forbidden", msg)
}

func NotFound(c *fiber.Ctx, resource ...string) error {
	msg := "Resource"
	if len(resource) > 0 {
		msg = resource[0]
	}
	return Fail(c, 404, "not_found", msg+" not found")
}

func InternalError(c *fiber.Ctx, message ...string) error {
	msg := "An internal error occurred"
	if len(message) > 0 {
		msg = message[0]
	}
	return Fail(c, 500, "server_error", msg)
}

// MapError convertit les erreurs métier en réponses HTTP appropriées et logue le problème.
func MapError(c *fiber.Ctx, logger *zap.Logger, err error) error {
	if err == nil {
		return nil
	}

	// Logging de base pour toutes les erreurs traitées par MapError
	logFields := []zap.Field{
		zap.String("path", c.Path()),
		zap.String("method", c.Method()),
		zap.Error(err),
	}

	switch err.Error() {
	case "resource_not_found":
		logger.Warn("Resource not found", logFields...)
		return NotFound(c, "Resource")
	case "unauthorized":
		logger.Warn("Unauthorized access attempt", logFields...)
		return Unauthorized(c)
	case "forbidden":
		logger.Warn("Forbidden access attempt", logFields...)
		return Forbidden(c)
	case "bid_conflict":
		logger.Warn("Bid conflict occurred", logFields...)
		return Fail(c, 409, "bid_conflict", "Bid conflict, please retry")
	case "insufficient_balance":
		logger.Warn("Insufficient balance for operation", logFields...)
		return Fail(c, 422, "insufficient_balance", "Insufficient wallet balance")
	case "auction_not_active", "auction_ended":
		logger.Warn("Action on ended auction", logFields...)
		return Fail(c, 422, "auction_ended", "This auction is no longer active")
	case "bid_too_low":
		logger.Info("Bid too low", logFields...)
		return Fail(c, 422, "bid_too_low", "Bid amount is too low")
	case "otp_expired":
		logger.Info("OTP expired", logFields...)
		return Fail(c, 422, "otp_expired", "OTP has expired")
	case "otp_invalid":
		logger.Info("OTP invalid", logFields...)
		return Fail(c, 422, "otp_invalid", "Invalid OTP code")
	case "otp_max_attempts":
		logger.Warn("OTP max attempts reached", logFields...)
		return Fail(c, 429, "otp_max_attempts", "Too many attempts, please request a new OTP")
	case "phone_already_registered":
		logger.Info("Registration attempt with existing phone", logFields...)
		return Fail(c, 409, "duplicate_phone", "Phone number already registered")
	default:
		errStr := err.Error()
		if strings.Contains(errStr, "at least 1 minute in the future") {
			logger.Info("Auction date validation failed", logFields...)
			return BadRequest(c, "تاريخ الإغلاق يجب أن يكون في المستقبل")
		}
		
		// Gestion des erreurs de base de données courantes (Postgres)
		if contains(errStr, "23505") || contains(errStr, "duplicate key") {
			logger.Warn("Database unique constraint violation", logFields...)
			return Fail(c, 409, "duplicate_record", "A record with this unique identifier already exists")
		}
		if contains(errStr, "23503") || contains(errStr, "foreign key") {
			logger.Warn("Database foreign key violation", logFields...)
			return Fail(c, 400, "invalid_reference", "The provided reference is invalid or does not exist")
		}

		// Erreur interne imprévue : Log au niveau ERROR
		logger.Error("Unhandled internal error", logFields...)
		return InternalError(c)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && (func() bool {
		for i := 0; i < len(s)-len(substr)+1; i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()))
}



