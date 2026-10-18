package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
)

// BankilyOutcome is a non-success state change for an unresolved attempt.
type BankilyOutcome struct {
	Status              string // PENDING_VERIFICATION | FAILED
	BankilyErrorCode    *string
	BankilyErrorMessage *string
	BankilyStatus       *string
	BankilyTxID         *string
	FailureReason       *string
	NextCheckAt         *time.Time
	NeedsReview         bool
	CountCheck          bool // increments check_attempts / last_checked_at
}

type BankilyPaymentRepository interface {
	Create(ctx context.Context, p *models.BankilyPayment) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.BankilyPayment, error)
	GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*models.BankilyPayment, error)
	GetByIdempotencyKey(ctx context.Context, userID uuid.UUID, key string) (*models.BankilyPayment, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]models.BankilyPayment, error)
	ListAdmin(ctx context.Context, status string, page, perPage int) ([]models.BankilyPaymentAdminRow, int, error)
	MarkSubmitted(ctx context.Context, id uuid.UUID) error
	// ApplyOutcome only ever touches an unresolved row (INITIATED/SUBMITTED/
	// PENDING_VERIFICATION) -- a SUCCESS or FAILED row can never be rewritten.
	// Returns false when the row was already terminal.
	ApplyOutcome(ctx context.Context, id uuid.UUID, o BankilyOutcome) (bool, error)
	// ClaimDue leases up to limit unresolved rows whose next check is due, so
	// concurrent workers (several instances) never check the same row at once.
	ClaimDue(ctx context.Context, limit int, lease time.Duration) ([]models.BankilyPayment, error)
	LockForUpdate(ctx context.Context, dbtx *sqlx.Tx, id uuid.UUID) (*models.BankilyPayment, error)
	MarkSuccessTx(ctx context.Context, dbtx *sqlx.Tx, id uuid.UUID, walletTxID uuid.UUID, bankilyTxID, bankilyStatus *string) error
	FlagStale(ctx context.Context, olderThan time.Time) (int64, error)
	// FailOrphanInitiated fails INITIATED rows older than the cutoff. INITIATED
	// is only ever set before MarkSubmitted, which runs before the /payment
	// call -- so such a row was provably never sent to Bankily.
	FailOrphanInitiated(ctx context.Context, olderThan time.Time) (int64, error)
}

type bankilyPaymentRepo struct{ db *sqlx.DB }

func NewBankilyPaymentRepository(db *sqlx.DB) BankilyPaymentRepository {
	return &bankilyPaymentRepo{db: db}
}

const bankilyUnresolved = `('INITIATED','SUBMITTED','PENDING_VERIFICATION')`

func (r *bankilyPaymentRepo) Create(ctx context.Context, p *models.BankilyPayment) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO bankily_payments
			(id, user_id, operation_id, idempotency_key, amount, currency_code,
			 client_phone, language, status)
		VALUES
			(:id, :user_id, :operation_id, :idempotency_key, :amount, :currency_code,
			 :client_phone, :language, :status)`, p)
	return err
}

func (r *bankilyPaymentRepo) get(ctx context.Context, q string, args ...any) (*models.BankilyPayment, error) {
	var p models.BankilyPayment
	if err := r.db.GetContext(ctx, &p, q, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *bankilyPaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.BankilyPayment, error) {
	return r.get(ctx, `SELECT * FROM bankily_payments WHERE id = $1`, id)
}

func (r *bankilyPaymentRepo) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*models.BankilyPayment, error) {
	return r.get(ctx, `SELECT * FROM bankily_payments WHERE id = $1 AND user_id = $2`, id, userID)
}

func (r *bankilyPaymentRepo) GetByIdempotencyKey(ctx context.Context, userID uuid.UUID, key string) (*models.BankilyPayment, error) {
	return r.get(ctx, `SELECT * FROM bankily_payments WHERE user_id = $1 AND idempotency_key = $2`, userID, key)
}

func (r *bankilyPaymentRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]models.BankilyPayment, error) {
	out := []models.BankilyPayment{}
	err := r.db.SelectContext(ctx, &out, `
		SELECT * FROM bankily_payments WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2`, userID, limit)
	return out, err
}

func (r *bankilyPaymentRepo) ListAdmin(ctx context.Context, status string, page, perPage int) ([]models.BankilyPaymentAdminRow, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM bankily_payments WHERE ($1 = '' OR status = $1)`, status); err != nil {
		return nil, 0, err
	}
	out := []models.BankilyPaymentAdminRow{}
	err := r.db.SelectContext(ctx, &out, `
		SELECT bp.*, u.full_name AS user_full_name, u.phone AS user_phone
		FROM bankily_payments bp
		LEFT JOIN users u ON u.id = bp.user_id
		WHERE ($1 = '' OR bp.status = $1)
		ORDER BY bp.created_at DESC
		LIMIT $2 OFFSET $3`, status, perPage, (page-1)*perPage)
	return out, total, err
}

func (r *bankilyPaymentRepo) MarkSubmitted(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE bankily_payments
		SET status = 'SUBMITTED', submitted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'INITIATED'`, id)
	return err
}

func (r *bankilyPaymentRepo) ApplyOutcome(ctx context.Context, id uuid.UUID, o BankilyOutcome) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE bankily_payments SET
			status                 = $2::varchar,
			bankily_error_code     = COALESCE($3, bankily_error_code),
			bankily_error_message  = COALESCE($4, bankily_error_message),
			bankily_status         = COALESCE($5, bankily_status),
			bankily_transaction_id = COALESCE(bankily_transaction_id, $6),
			failure_reason         = COALESCE($7, failure_reason),
			next_check_at          = CASE WHEN $2::varchar = 'FAILED' THEN NULL ELSE $8::timestamptz END,
			needs_review           = needs_review OR $9::boolean,
			check_attempts         = check_attempts + CASE WHEN $10::boolean THEN 1 ELSE 0 END,
			last_checked_at        = CASE WHEN $10::boolean THEN NOW() ELSE last_checked_at END,
			completed_at           = CASE WHEN $2::varchar = 'FAILED' THEN NOW() ELSE completed_at END,
			updated_at             = NOW()
		WHERE id = $1 AND status IN `+bankilyUnresolved,
		id, o.Status, o.BankilyErrorCode, o.BankilyErrorMessage, o.BankilyStatus, o.BankilyTxID,
		o.FailureReason, o.NextCheckAt, o.NeedsReview, o.CountCheck)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}

func (r *bankilyPaymentRepo) ClaimDue(ctx context.Context, limit int, lease time.Duration) ([]models.BankilyPayment, error) {
	out := []models.BankilyPayment{}
	err := r.db.SelectContext(ctx, &out, `
		UPDATE bankily_payments SET next_check_at = NOW() + make_interval(secs => $2)
		WHERE id IN (
			SELECT id FROM bankily_payments
			WHERE status IN ('SUBMITTED','PENDING_VERIFICATION')
			  AND COALESCE(next_check_at, submitted_at, created_at) <= NOW()
			ORDER BY COALESCE(next_check_at, submitted_at, created_at)
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING *`, limit, lease.Seconds())
	return out, err
}

func (r *bankilyPaymentRepo) LockForUpdate(ctx context.Context, dbtx *sqlx.Tx, id uuid.UUID) (*models.BankilyPayment, error) {
	var p models.BankilyPayment
	if err := dbtx.GetContext(ctx, &p, `SELECT * FROM bankily_payments WHERE id = $1 FOR UPDATE`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *bankilyPaymentRepo) MarkSuccessTx(ctx context.Context, dbtx *sqlx.Tx, id uuid.UUID, walletTxID uuid.UUID, bankilyTxID, bankilyStatus *string) error {
	res, err := dbtx.ExecContext(ctx, `
		UPDATE bankily_payments SET
			status = 'SUCCESS',
			wallet_transaction_id  = $2,
			bankily_transaction_id = COALESCE(bankily_transaction_id, $3),
			bankily_status         = COALESCE($4, bankily_status),
			check_attempts  = check_attempts + 1,
			last_checked_at = NOW(),
			next_check_at = NULL,
			completed_at  = NOW(),
			updated_at    = NOW()
		WHERE id = $1 AND status IN `+bankilyUnresolved, id, walletTxID, bankilyTxID, bankilyStatus)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n != 1 {
		return apperr.ErrConflict
	}
	return nil
}

func (r *bankilyPaymentRepo) FlagStale(ctx context.Context, olderThan time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE bankily_payments SET needs_review = TRUE, updated_at = NOW()
		WHERE status IN `+bankilyUnresolved+` AND NOT needs_review AND created_at < $1`, olderThan)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *bankilyPaymentRepo) FailOrphanInitiated(ctx context.Context, olderThan time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE bankily_payments
		SET status = 'FAILED', failure_reason = 'not_submitted', completed_at = NOW(), updated_at = NOW()
		WHERE status = 'INITIATED' AND created_at < $1`, olderThan)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
