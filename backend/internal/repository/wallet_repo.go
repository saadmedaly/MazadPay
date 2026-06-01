package repository

import (
    "context"

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
