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

// Create écrit désormais target_phone_hash/target_phone_masked (Admin Authorization
// Phase 1C-B) — nullable si l'appelant ne les fournit pas (compatibilité), mais
// GenerateAdminInvitation les fournit systématiquement depuis cette phase.
func (r *adminInvitationRepo) Create(ctx context.Context, inv *models.AdminInvitation) error {
	query := `INSERT INTO admin_invitations (id, token, created_by, expires_at, target_phone_hash, target_phone_masked)
		VALUES (:id, :token, :created_by, :expires_at, :target_phone_hash, :target_phone_masked)`
	_, err := r.db.NamedExecContext(ctx, query, inv)
	return err
}

// GetByToken utilise SELECT * : les nouvelles colonnes target_phone_hash/
// target_phone_masked sont automatiquement peuplées dans le struct via leurs tags
// `db:` dès que la migration 000043 est appliquée (NULL pour les invitations
// existantes, comme pour toute invitation créée avant Phase 1C-B) — aucun
// changement de requête nécessaire ici.
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
