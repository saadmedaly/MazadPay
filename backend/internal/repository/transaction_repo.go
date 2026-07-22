package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/shopspring/decimal"
)

type TransactionRepository interface {
	ListPaginated(ctx context.Context, page, perPage int, status string, userID *uuid.UUID) ([]models.Transaction, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	FindByID(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*models.Transaction, error)
	Create(ctx context.Context, tx *models.Transaction) error
	CreateTx(ctx context.Context, dbtx *sqlx.Tx, tx *models.Transaction) error
	UpdateReceipt(ctx context.Context, id uuid.UUID, userID uuid.UUID, url string, status string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status, notes string, adminID uuid.UUID) error
	GetStats(ctx context.Context) (float64, float64, error) // Total, Today
	GetPendingCount(ctx context.Context) (int, error)
	GetWeeklySum(ctx context.Context) (float64, error)
	GetDailyRevenueChart(ctx context.Context) ([]map[string]interface{}, error)
}

type transactionRepo struct {
	db         *sqlx.DB
	walletRepo WalletRepository
}

func NewTransactionRepository(db *sqlx.DB, walletRepo WalletRepository) TransactionRepository {
	return &transactionRepo{db: db, walletRepo: walletRepo}
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

// explicitTxJoinQuery énumère explicitement chaque colonne de transactions (plutôt que
// t.*) suivie du LEFT JOIN vers users — évite toute ambiguïté de scan sqlx liée à
// SELECT * combiné à des colonnes supplémentaires (audit : user_full_name/user_phone
// n'apparaissaient jamais dans la réponse API malgré des données correctes en base et un
// code source identique sur le remote ; cette requête explicite + scan manuel sert de
// garantie supplémentaire, indépendante de tout comportement implicite de StructScan).
const explicitTxJoinQuery = `
	SELECT
		t.id, t.user_id, t.auction_id, t.type, t.amount, t.gateway, t.status,
		t.reference, t.receipt_url, t.admin_notes, t.reviewed_by, t.reviewed_at,
		t.wallet_hold_id, t.receipt_image_temp, t.payment_method, t.fee_amount,
		t.net_amount, t.description, t.failure_reason, t.created_at,
		u.full_name AS user_full_name, u.phone AS user_phone
	FROM transactions t
	LEFT JOIN users u ON u.id = t.user_id
	WHERE t.id = $1`

// scanTxJoinRow scanne manuellement une ligne du résultat de explicitTxJoinQuery dans
// un models.Transaction — indépendant de StructScan pour garantir que user_full_name et
// user_phone sont bien assignés.
func scanTxJoinRow(row *sqlx.Row) (*models.Transaction, error) {
	var tx models.Transaction
	err := row.Scan(
		&tx.ID, &tx.UserID, &tx.AuctionID, &tx.Type, &tx.Amount, &tx.Gateway, &tx.Status,
		&tx.Reference, &tx.ReceiptURL, &tx.AdminNotes, &tx.ReviewedBy, &tx.ReviewedAt,
		&tx.WalletHoldID, &tx.ReceiptImageTemp, &tx.PaymentMethod, &tx.FeeAmount,
		&tx.NetAmount, &tx.Description, &tx.FailureReason, &tx.CreatedAt,
		&tx.UserFullName, &tx.UserPhone,
	)
	return &tx, err
}

// GetByID est utilisé par la vue admin "détail de transaction" — LEFT JOIN vers users
// avec scan manuel explicite (voir explicitTxJoinQuery/scanTxJoinRow ci-dessus) pour
// garantir que user_full_name/user_phone sont bien renvoyés à l'API.
func (r *transactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	row := r.db.QueryRowxContext(ctx, explicitTxJoinQuery, id)
	return scanTxJoinRow(row)
}

// FindByID inclut le même LEFT JOIN explicite que GetByID (voir explicitTxJoinQuery) —
// utilisé par l'admin (GetTransactionByID) et l'utilisateur (GetTransaction/GetTransactionAny).
func (r *transactionRepo) FindByID(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*models.Transaction, error) {
	if userID != nil {
		row := r.db.QueryRowxContext(ctx, explicitTxJoinQuery+" AND t.user_id = $2", id, userID)
		return scanTxJoinRow(row)
	}
	row := r.db.QueryRowxContext(ctx, explicitTxJoinQuery, id)
	return scanTxJoinRow(row)
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

// CreateTx est identique à Create mais s'exécute dans une transaction SQL existante
// (utilisé par RequestWithdraw pour que le gel du solde et la création de la
// transaction soient atomiques — audit de sécurité V09).
func (r *transactionRepo) CreateTx(ctx context.Context, dbtx *sqlx.Tx, tx *models.Transaction) error {
	_, err := dbtx.NamedExecContext(ctx, `
		INSERT INTO transactions
			(id, user_id, auction_id, type, amount, gateway, status, reference,
			 receipt_url, admin_notes, reviewed_by, reviewed_at, wallet_hold_id)
		VALUES
			(:id, :user_id, :auction_id, :type, :amount, :gateway, :status, :reference,
			 :receipt_url, :admin_notes, :reviewed_by, :reviewed_at, :wallet_hold_id)
	`, tx)
	return err
}

// UpdateReceipt met à jour le reçu d'une transaction. userID est requis et vérifié dans
// la clause WHERE pour empêcher un utilisateur de modifier le reçu d'une transaction
// qui ne lui appartient pas (voir audit de sécurité V05 - IDOR).
func (r *transactionRepo) UpdateReceipt(ctx context.Context, id uuid.UUID, userID uuid.UUID, url string, status string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE transactions SET receipt_url = $1, status = $2
		WHERE id = $3 AND user_id = $4`, url, status, id, userID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func (r *transactionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status, notes string, adminID uuid.UUID) error {
	dbtx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer dbtx.Rollback()

	var tx models.Transaction
	if err := dbtx.GetContext(ctx, &tx, "SELECT * FROM transactions WHERE id = $1 FOR UPDATE", id); err != nil {
		return err
	}

	// Dernier filet de sécurité avant tout mouvement de solde (audit V08) : un montant
	// nul ou négatif ne doit jamais créditer/débiter le portefeuille, même si une
	// transaction invalide existait déjà en base avant ce correctif.
	if status == "completed" && (tx.Type == "deposit" || tx.Type == "withdraw") && !tx.Amount.GreaterThan(decimal.Zero) {
		return apperr.ErrBadRequest
	}

	if _, err := dbtx.ExecContext(ctx, `
		UPDATE transactions
		SET status = $1, admin_notes = $2, reviewed_by = $3, reviewed_at = now()
		WHERE id = $4`, status, notes, adminID, id); err != nil {
		return err
	}

	// Credit wallet on deposit approval — only if not already completed (prevents double-credit)
	if status == "completed" && tx.Type == "deposit" && tx.Status != "completed" {
		// Ensure wallet row exists (users created before wallet auto-creation had no row)
		if _, err := dbtx.ExecContext(ctx,
			`INSERT INTO wallets (user_id, balance, frozen_amount, version, updated_at)
			 VALUES ($1, 0, 0, 1, now())
			 ON CONFLICT (user_id) DO NOTHING`,
			tx.UserID); err != nil {
			return err
		}
		if _, err := dbtx.ExecContext(ctx,
			`UPDATE wallets SET balance = balance + $1, version = version + 1 WHERE user_id = $2`,
			tx.Amount, tx.UserID); err != nil {
			return err
		}
	}

	// Withdrawal approval — le montant a déjà été gelé (balance -> frozen_amount) lors de
	// la demande via RequestWithdraw/FreezeForWithdraw (audit V09). On finalise ici en
	// retirant uniquement de frozen_amount, jamais de balance une seconde fois.
	if status == "completed" && tx.Type == "withdraw" && tx.Status != "completed" {
		if err := r.walletRepo.CaptureFrozenForWithdraw(ctx, dbtx, tx.UserID, tx.Amount); err != nil {
			return err
		}
	}

	// Withdrawal rejection — libère le montant gelé : il retourne dans balance et sort
	// de frozen_amount (audit V09). Ne s'applique que si le montant avait bien été gelé
	// (c'est-à-dire que la transaction était encore pending_review/pending).
	if status == "rejected" && tx.Type == "withdraw" && tx.Status != "completed" && tx.Status != "rejected" {
		if err := r.walletRepo.ReleaseFrozenForWithdraw(ctx, dbtx, tx.UserID, tx.Amount); err != nil {
			return err
		}
	}

	return dbtx.Commit()
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
