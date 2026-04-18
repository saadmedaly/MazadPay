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
	FindByID(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*models.Transaction, error)
	Create(ctx context.Context, tx *models.Transaction) error
	UpdateReceipt(ctx context.Context, id uuid.UUID, url string, status string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status, notes string, adminID uuid.UUID) error
	GetStats(ctx context.Context) (float64, float64, error) // Total, Today
	GetPendingCount(ctx context.Context) (int, error)
	GetWeeklySum(ctx context.Context) (float64, error)
	GetDailyRevenueChart(ctx context.Context) ([]map[string]interface{}, error)
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

func (r *transactionRepo) FindByID(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*models.Transaction, error) {
	var tx models.Transaction
	var err error
	if userID != nil {
		err = r.db.GetContext(ctx, &tx, "SELECT * FROM transactions WHERE id = $1 AND user_id = $2", id, userID)
	} else {
		err = r.db.GetContext(ctx, &tx, "SELECT * FROM transactions WHERE id = $1", id)
	}
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

func (r *transactionRepo) GetPendingCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM transactions WHERE status = 'pending_review'")
	return count, err
}

func (r *transactionRepo) GetWeeklySum(ctx context.Context) (float64, error) {
	var sum float64
	err := r.db.GetContext(ctx, &sum, "SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE status = 'completed' AND created_at >= now() - interval '7 days'")
	return sum, err
}

func (r *transactionRepo) GetDailyRevenueChart(ctx context.Context) ([]map[string]interface{}, error) {
	var data []map[string]interface{}
	query := `
		SELECT 
			TO_CHAR(d, 'YYYY-MM-DD') as date,
			COALESCE(SUM(t.amount), 0) as amount
		FROM 
			generate_series(now() - interval '29 days', now(), interval '1 day') d
		LEFT JOIN 
			transactions t ON t.created_at::date = d::date AND t.status = 'completed'
		GROUP BY 
			d
		ORDER BY 
			d ASC
	`
	rows, err := r.db.QueryxContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		m := make(map[string]interface{})
		err := rows.MapScan(m)
		if err != nil {
			return nil, err
		}
		data = append(data, m)
	}
	return data, nil
}
