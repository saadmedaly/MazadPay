package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type AdminInvitationRepository interface {
	Create(ctx context.Context, inv *models.AdminInvitation) error
	GetByToken(ctx context.Context, token string) (*models.AdminInvitation, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
}

type adminInvitationRepo struct {
	db *sqlx.DB
}

func NewAdminInvitationRepository(db *sqlx.DB) AdminInvitationRepository {
	return &adminInvitationRepo{db: db}
}

func (r *adminInvitationRepo) Create(ctx context.Context, inv *models.AdminInvitation) error {
	query := `INSERT INTO admin_invitations (id, token, created_by, expires_at) VALUES (:id, :token, :created_by, :expires_at)`
	_, err := r.db.NamedExecContext(ctx, query, inv)
	return err
}

func (r *adminInvitationRepo) GetByToken(ctx context.Context, token string) (*models.AdminInvitation, error) {
	var inv models.AdminInvitation
	query := `SELECT * FROM admin_invitations WHERE token = $1`
	err := r.db.GetContext(ctx, &inv, query, token)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &inv, err
}

func (r *adminInvitationRepo) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE admin_invitations SET used_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}
