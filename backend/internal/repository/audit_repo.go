package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type AuditRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]models.AuditLog, error)
	ListPaginated(ctx context.Context, page, perPage int) ([]models.AuditLog, int, error)
}

type auditRepo struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) AuditRepository {
	return &auditRepo{db: db}
}

func (r *auditRepo) Create(ctx context.Context, log *models.AuditLog) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO audit_logs (id, admin_id, action, entity_type, entity_id, details)
		VALUES (:id, :admin_id, :action, :entity_type, :entity_id, :details)
	`, log)
	return err
}

func (r *auditRepo) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.SelectContext(ctx, &logs, `
		SELECT * FROM audit_logs 
		WHERE entity_type = $1 AND entity_id = $2 
		ORDER BY created_at DESC
	`, entityType, entityID)
	return logs, err
}

func (r *auditRepo) ListPaginated(ctx context.Context, page, perPage int) ([]models.AuditLog, int, error) {
	var logs []models.AuditLog
	offset := (page - 1) * perPage

	var total int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM audit_logs")
	if err != nil {
		return nil, 0, err
	}

	err = r.db.SelectContext(ctx, &logs, `
		SELECT * FROM audit_logs 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`, perPage, offset)
	return logs, total, err
}