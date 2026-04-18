package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type ReportRepository interface {
	ListPaginated(ctx context.Context, page, perPage int, status string) ([]models.Report, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status, notes string, adminID uuid.UUID) error
	PendingCount(ctx context.Context) (int, error)
	Create(ctx context.Context, report *models.Report) error
}

type reportRepo struct {
	db *sqlx.DB
}

func NewReportRepository(db *sqlx.DB) ReportRepository {
	return &reportRepo{db: db}
}

func (r *reportRepo) ListPaginated(ctx context.Context, page, perPage int, status string) ([]models.Report, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if status != "" {
		where += " AND status = $1"
		args = append(args, status)
	}

	var total int
	err := r.db.GetContext(ctx, &total, fmt.Sprintf("SELECT COUNT(*) FROM reports %s", where), args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf("SELECT * FROM reports %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d", 
		where, len(args)+1, len(args)+2)
	
	listArgs := append(args, perPage, offset)
	reports := []models.Report{}
	err = r.db.SelectContext(ctx, &reports, query, listArgs...)
	
	return reports, total, err
}

func (r *reportRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status, notes string, adminID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE reports 
		SET status = $1, admin_notes = $2, reviewed_by = $3, reviewed_at = now() 
		WHERE id = $4`, status, notes, adminID, id)
	return err
}

func (r *reportRepo) PendingCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM reports WHERE status = 'pending'")
	return count, err
}

func (r *reportRepo) Create(ctx context.Context, report *models.Report) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO reports (id, auction_id, reporter_id, reason, status)
		VALUES (:id, :auction_id, :reporter_id, :reason, :status)
	`, report)
	return err
}
