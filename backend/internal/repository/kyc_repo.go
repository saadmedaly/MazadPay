package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
)

type KYCRepository interface {
	Create(ctx context.Context, kyc *models.KYCVerification) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.KYCVerification, error)
	List(ctx context.Context, status string) ([]models.KYCVerification, error)
	UpdateStatus(ctx context.Context, userID uuid.UUID, status string, notes string, adminID uuid.UUID) error
}

type kycRepo struct {
	db *sqlx.DB
}

func NewKYCRepository(db *sqlx.DB) KYCRepository {
	return &kycRepo{db: db}
}

func (r *kycRepo) Create(ctx context.Context, kyc *models.KYCVerification) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO kyc_verifications (user_id, id_card_front_url, id_card_back_url, nni_number, status)
		VALUES (:user_id, :id_card_front_url, :id_card_back_url, :nni_number, :status)
		ON CONFLICT (user_id) DO UPDATE SET
			id_card_front_url = EXCLUDED.id_card_front_url,
			id_card_back_url = EXCLUDED.id_card_back_url,
			nni_number = EXCLUDED.nni_number,
			status = EXCLUDED.status,
			created_at = now()
	`, kyc)
	return err
}

func (r *kycRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.KYCVerification, error) {
	var kyc models.KYCVerification
	err := r.db.GetContext(ctx, &kyc, `SELECT * FROM kyc_verifications WHERE user_id = $1`, userID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return &kyc, nil
}

func (r *kycRepo) List(ctx context.Context, status string) ([]models.KYCVerification, error) {
	query := `SELECT * FROM kyc_verifications`
	var args []interface{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC`

	kycs := []models.KYCVerification{}
	err := r.db.SelectContext(ctx, &kycs, query, args...)
	return kycs, err
}

func (r *kycRepo) UpdateStatus(ctx context.Context, userID uuid.UUID, status string, notes string, adminID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE kyc_verifications SET 
			status = $1, 
			admin_notes = $2, 
			reviewed_by = $3, 
			reviewed_at = $4 
		WHERE user_id = $5`,
		status, notes, adminID, time.Now(), userID)
	return err
}
