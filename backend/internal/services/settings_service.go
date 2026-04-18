package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
)

type SettingsService interface {
	Get(ctx context.Context, key string) (*models.SystemSettings, error)
	Set(ctx context.Context, key, value, settingType string, userID uuid.UUID) error
	List(ctx context.Context) ([]models.SystemSettings, error)
}

type settingsService struct {
	repo repository.SettingsRepository
}

func NewSettingsService(repo repository.SettingsRepository) SettingsService {
	return &settingsService{repo: repo}
}

func (s *settingsService) Get(ctx context.Context, key string) (*models.SystemSettings, error) {
	return s.repo.Get(ctx, key)
}

func (s *settingsService) Set(ctx context.Context, key, value, settingType string, userID uuid.UUID) error {
	return s.repo.Set(ctx, key, value, settingType)
}

func (s *settingsService) List(ctx context.Context) ([]models.SystemSettings, error) {
	return s.repo.List(ctx)
}