package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/mazadpay/backend/internal/services"
    "go.uber.org/zap"
)

type UserHandler struct {
    service services.UserService
    logger  *zap.Logger
}

func NewUserHandler(svc services.UserService, logger *zap.Logger) *UserHandler {
    return &UserHandler{service: svc, logger: logger}
}


func (h *UserHandler) GetMe(c *fiber.Ctx) error {
    userID, err := uuid.Parse(GetUserID(c))
    if err != nil {
        return Unauthorized(c)
    }

    user, err := h.service.GetProfile(c.Context(), userID)
    if err != nil {
        return MapError(c, h.logger, err)
    }


    return OK(c, user)
}
