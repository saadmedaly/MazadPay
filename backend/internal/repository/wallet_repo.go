package repository

import (
    "context"
    "database/sql"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    apperr "github.com/mazadpay/backend/internal/errors"
    "github.com/mazadpay/backend/internal/models"
    "github.com/shopspring/decimal"
)

type WalletRepository interface {
    GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error)
    FindForUpdate(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) (*models.Wallet, error)
    FindActiveHold(ctx context.Context, tx *sqlx.Tx, userID, auctionID uuid.UUID) (*models.WalletHold, error)
    CreateHold(ctx context.Context, tx *sqlx.Tx, userID, auctionID uuid.UUID, amount decimal.Decimal) error
    DebitFreezeBalance(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, amount decimal.Decimal, version int) error
    ReleaseHold(ctx context.Context, tx *sqlx.Tx, holdID uuid.UUID) error
    ReleaseHoldsForAuction(ctx context.Context, tx *sqlx.Tx, auctionID uuid.UUID) error
    ReleaseHoldsForNonWinners(ctx context.Context, tx *sqlx.Tx, auctionID uuid.UUID, winnerID *uuid.UUID) error
    FreezeForWithdraw(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, amount decimal.Decimal) error
    CaptureFrozenForWithdraw(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, amount decimal.Decimal) error
    ReleaseFrozenForWithdraw(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, amount decimal.Decimal) error
    GetPaymentMethods(ctx context.Context) ([]models.PaymentMethod, error)
}

type walletRepo struct{ db *sqlx.DB }

func NewWalletRepository(db *sqlx.DB) WalletRepository {
    return &walletRepo{db: db}
}

func (r *walletRepo) GetPaymentMethods(ctx context.Context) ([]models.PaymentMethod, error) {
    var methods []models.PaymentMethod
    err := r.db.SelectContext(ctx, &methods,
        `SELECT id, code, name_ar, name_fr, name_en, logo_url, is_active, country_id, created_at
         FROM payment_methods WHERE is_active = true ORDER BY id`)
    return methods, err
}

func (r *walletRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
    var w models.Wallet
    err := r.db.GetContext(ctx, &w, `SELECT * FROM wallets WHERE user_id = $1`, userID)
    if err == nil {
        return &w, nil
    }
    // No wallet row yet — create one with zero balance then return it
    _, err = r.db.ExecContext(ctx,
        `INSERT INTO wallets (user_id, balance, frozen_amount, version, updated_at)
         VALUES ($1, 0, 0, 1, now())
         ON CONFLICT (user_id) DO NOTHING`,
        userID)
    if err != nil {
        return nil, err
    }
    err = r.db.GetContext(ctx, &w, `SELECT * FROM wallets WHERE user_id = $1`, userID)
    if err != nil {
        return nil, err
    }
    return &w, nil
}

func (r *walletRepo) FindForUpdate(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) (*models.Wallet, error) {
    var w models.Wallet
    err := tx.GetContext(ctx, &w, `SELECT * FROM wallets WHERE user_id = $1 FOR UPDATE`, userID)
    if err != nil {
        return nil, apperr.ErrNotFound
    }
    return &w, nil
}

func (r *walletRepo) FindActiveHold(ctx context.Context, tx *sqlx.Tx, userID, auctionID uuid.UUID) (*models.WalletHold, error) {
    var hold models.WalletHold
    err := tx.GetContext(ctx, &hold, 
        `SELECT * FROM wallet_holds WHERE user_id = $1 AND auction_id = $2 AND status = 'active'`, 
        userID, auctionID)
    if err != nil {
        return nil, err
    }
    return &hold, nil
}

func (r *walletRepo) CreateHold(ctx context.Context, tx *sqlx.Tx, userID, auctionID uuid.UUID, amount decimal.Decimal) error {
    _, err := tx.ExecContext(ctx,
        `INSERT INTO wallet_holds (id, user_id, auction_id, amount, status)
         VALUES ($1, $2, $3, $4, 'active')`,
        uuid.New(), userID, auctionID, amount)
    return err
}

func (r *walletRepo) DebitFreezeBalance(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, amount decimal.Decimal, version int) error {
    result, err := tx.ExecContext(ctx,
        `UPDATE wallets SET balance = balance - $1, frozen_amount = frozen_amount + $1, version = version + 1
         WHERE user_id = $2 AND version = $3 AND balance >= $1`,
        amount, userID, version)
    if err != nil {
        return err
    }
    n, _ := result.RowsAffected()
    if n == 0 {
        return apperr.ErrInsufficientBalance
    }
    return nil
}

// ReleaseHold libère un wallet_hold actif : remet le montant dans balance, retire
// frozen_amount, et marque le hold comme "released". Idempotent — si le hold n'est
// plus "active" (déjà released/captured), aucune ligne n'est affectée et aucune erreur
// n'est retournée (voir audit de sécurité V03/V09).
func (r *walletRepo) ReleaseHold(ctx context.Context, tx *sqlx.Tx, holdID uuid.UUID) error {
    var amount decimal.Decimal
    var userID uuid.UUID
    err := tx.QueryRowContext(ctx,
        `UPDATE wallet_holds SET status = 'released', released_at = now()
         WHERE id = $1 AND status = 'active'
         RETURNING amount, user_id`, holdID).Scan(&amount, &userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil // déjà libéré ou introuvable : idempotent, pas d'erreur
        }
        return err
    }
    _, err = tx.ExecContext(ctx,
        `UPDATE wallets SET balance = balance + $1, frozen_amount = frozen_amount - $1
         WHERE user_id = $2`, amount, userID)
    return err
}

// ReleaseHoldsForAuction libère tous les wallet_holds actifs d'un auction donné
// (utilisé lors de l'annulation d'un auction — tous les enchérisseurs récupèrent
// leur caution).
func (r *walletRepo) ReleaseHoldsForAuction(ctx context.Context, tx *sqlx.Tx, auctionID uuid.UUID) error {
    rows, err := tx.QueryContext(ctx,
        `SELECT id FROM wallet_holds WHERE auction_id = $1 AND status = 'active'`, auctionID)
    if err != nil {
        return err
    }
    var holdIDs []uuid.UUID
    for rows.Next() {
        var id uuid.UUID
        if err := rows.Scan(&id); err != nil {
            rows.Close()
            return err
        }
        holdIDs = append(holdIDs, id)
    }
    rows.Close()
    for _, id := range holdIDs {
        if err := r.ReleaseHold(ctx, tx, id); err != nil {
            return err
        }
    }
    return nil
}

// ReleaseHoldsForNonWinners libère les wallet_holds actifs de tous les enchérisseurs
// d'un auction SAUF le gagnant (utilisé à la clôture de l'auction — le hold du
// gagnant reste actif en attente d'une décision de capture/remboursement ultérieure,
// qui n'existe pas encore dans ce projet, voir audit V03/V09).
func (r *walletRepo) ReleaseHoldsForNonWinners(ctx context.Context, tx *sqlx.Tx, auctionID uuid.UUID, winnerID *uuid.UUID) error {
    var rows *sql.Rows
    var err error
    if winnerID != nil {
        rows, err = tx.QueryContext(ctx,
            `SELECT id FROM wallet_holds WHERE auction_id = $1 AND status = 'active' AND user_id != $2`,
            auctionID, *winnerID)
    } else {
        rows, err = tx.QueryContext(ctx,
            `SELECT id FROM wallet_holds WHERE auction_id = $1 AND status = 'active'`, auctionID)
    }
    if err != nil {
        return err
    }
    var holdIDs []uuid.UUID
    for rows.Next() {
        var id uuid.UUID
        if err := rows.Scan(&id); err != nil {
            rows.Close()
            return err
        }
        holdIDs = append(holdIDs, id)
    }
    rows.Close()
    for _, id := range holdIDs {
        if err := r.ReleaseHold(ctx, tx, id); err != nil {
            return err
        }
    }
    return nil
}

// FreezeForWithdraw déplace `amount` de balance vers frozen_amount de façon atomique
// (row locking via la condition WHERE, pas besoin de FOR UPDATE séparé grâce au
// compare-and-set). Retourne ErrInsufficientBalance si le solde est insuffisant.
func (r *walletRepo) FreezeForWithdraw(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, amount decimal.Decimal) error {
    result, err := tx.ExecContext(ctx,
        `UPDATE wallets SET balance = balance - $1, frozen_amount = frozen_amount + $1
         WHERE user_id = $2 AND balance >= $1`,
        amount, userID)
    if err != nil {
        return err
    }
    n, _ := result.RowsAffected()
    if n == 0 {
        return apperr.ErrInsufficientBalance
    }
    return nil
}

// CaptureFrozenForWithdraw finalise un retrait approuvé : retire simplement le
// montant de frozen_amount (il a déjà quitté balance lors de la demande via
// FreezeForWithdraw). Ne touche jamais balance ici pour éviter un double débit.
func (r *walletRepo) CaptureFrozenForWithdraw(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, amount decimal.Decimal) error {
    result, err := tx.ExecContext(ctx,
        `UPDATE wallets SET frozen_amount = frozen_amount - $1
         WHERE user_id = $2 AND frozen_amount >= $1`,
        amount, userID)
    if err != nil {
        return err
    }
    n, _ := result.RowsAffected()
    if n == 0 {
        return apperr.ErrInsufficientBalance
    }
    return nil
}

// ReleaseFrozenForWithdraw annule un retrait rejeté : remet le montant dans balance
// et le retire de frozen_amount.
func (r *walletRepo) ReleaseFrozenForWithdraw(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, amount decimal.Decimal) error {
    result, err := tx.ExecContext(ctx,
        `UPDATE wallets SET balance = balance + $1, frozen_amount = frozen_amount - $1
         WHERE user_id = $2 AND frozen_amount >= $1`,
        amount, userID)
    if err != nil {
        return err
    }
    n, _ := result.RowsAffected()
    if n == 0 {
        return apperr.ErrInsufficientBalance
    }
    return nil
}
