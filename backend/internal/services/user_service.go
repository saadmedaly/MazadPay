package services

import (
    "context"

    "github.com/google/uuid"
    "github.com/mazadpay/backend/internal/models"
    "github.com/mazadpay/backend/internal/repository"
)

type UserService interface {
    GetProfile(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type userService struct {
    repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
    return &userService{repo: repo}
}

func (s *userService) GetProfile(ctx context.Context, id uuid.UUID) (*models.User, error) {
    return s.repo.FindByID(ctx, id)
}
