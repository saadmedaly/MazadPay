package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type TransactionRepository interface {
	ListPaginated(ctx context.Context, page, perPage int, status string, userID *uuid.UUID) ([]models.Transaction, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	Create(ctx context.Context, tx *models.Transaction) error
	UpdateReceipt(ctx context.Context, id uuid.UUID, url string, status string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status, notes string, adminID uuid.UUID) error
	GetStats(ctx context.Context) (float64, float64, error) // Total, Today
}

type transactionRepo struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) ListPaginated(ctx context.Context, page, perPage int, status string, userID *uuid.UUID) ([]models.Transaction, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", len(args)+1)
		args = append(args, status)
	}
	if userID != nil {
		where += fmt.Sprintf(" AND user_id = $%d", len(args)+1)
		args = append(args, *userID)
	}

	var total int
	err := r.db.GetContext(ctx, &total, fmt.Sprintf("SELECT COUNT(*) FROM transactions %s", where), args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf("SELECT * FROM transactions %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d", 
		where, len(args)+1, len(args)+2)
	
	listArgs := append(args, perPage, offset)
	txs := []models.Transaction{}
	err = r.db.SelectContext(ctx, &txs, query, listArgs...)
	
	return txs, total, err
}

func (r *transactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.GetContext(ctx, &tx, "SELECT * FROM transactions WHERE id = $1", id)
	return &tx, err
}

func (r *transactionRepo) Create(ctx context.Context, tx *models.Transaction) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO transactions 
			(id, user_id, auction_id, type, amount, gateway, status, reference, 
			 receipt_url, admin_notes, reviewed_by, reviewed_at, wallet_hold_id)
		VALUES 
			(:id, :user_id, :auction_id, :type, :amount, :gateway, :status, :reference,
			 :receipt_url, :admin_notes, :reviewed_by, :reviewed_at, :wallet_hold_id)
	`, tx)
	return err
}

func (r *transactionRepo) UpdateReceipt(ctx context.Context, id uuid.UUID, url string, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE transactions SET receipt_url = $1, status = $2 
		WHERE id = $3`, url, status, id)
	return err
}

func (r *transactionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status, notes string, adminID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE transactions 
		SET status = $1, admin_notes = $2, reviewed_by = $3, reviewed_at = now() 
		WHERE id = $4`, status, notes, adminID, id)
	return err
}

func (r *transactionRepo) GetStats(ctx context.Context) (float64, float64, error) {
	var total, today float64
	err := r.db.GetContext(ctx, &total, "SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE status = 'completed'")
	if err != nil {
		return 0, 0, err
	}
	err = r.db.GetContext(ctx, &today, "SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE status = 'completed' AND created_at >= CURRENT_DATE")
	return total, today, err
}
