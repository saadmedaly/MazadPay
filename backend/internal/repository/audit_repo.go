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
	// actor_type a un DEFAULT 'unknown' en base (voir 000042_audit_logs_schema_improvement) ;
	// on l'aligne ici uniquement si l'appelant ne l'a pas renseigné, pour ne pas
	// dépendre d'un ordre d'exécution particulier entre Log() et Create().
	if log.ActorType == "" {
		log.ActorType = "unknown"
	}
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO audit_logs (
			id, admin_id, action, entity_type, entity_id, details,
			actor_id, actor_type, ip_address, user_agent, details_json, entity_key
		)
		VALUES (
			:id, :admin_id, :action, :entity_type, :entity_id, :details,
			:actor_id, :actor_type, :ip_address, :user_agent, :details_json, :entity_key
		)
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