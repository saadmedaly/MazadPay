package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
)

type AuditService interface {
	Log(ctx context.Context, adminID uuid.UUID, action, entityType string, entityID *uuid.UUID, details string) error
	GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]models.AuditLog, error)
	List(ctx context.Context, page, perPage int) ([]models.AuditLog, int, error)
}

type auditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) Log(ctx context.Context, adminID uuid.UUID, action, entityType string, entityID *uuid.UUID, details string) error {
	log := &models.AuditLog{
		ID:         uuid.New(),
		AdminID:    adminID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:   details,
	}
	return s.repo.Create(ctx, log)
}

func (s *auditService) GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]models.AuditLog, error) {
	return s.repo.ListByEntity(ctx, entityType, entityID)
}

func (s *auditService) List(ctx context.Context, page, perPage int) ([]models.AuditLog, int, error) {
	return s.repo.ListPaginated(ctx, page, perPage)
}