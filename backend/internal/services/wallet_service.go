package services

import (
	"context"

	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/shopspring/decimal"
)

type WalletService interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (*models.Wallet, error)
	InitiateDeposit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, gateway string) (*models.Transaction, error)
	UploadReceipt(ctx context.Context, txID uuid.UUID, receiptURL string) error
	RequestWithdraw(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, gateway string) (*models.Transaction, error)
	GetTransactions(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Transaction, int, error)
	GetTransaction(ctx context.Context, userID uuid.UUID, txID uuid.UUID) (*models.Transaction, error)
}

type walletService struct {
	walletRepo repository.WalletRepository
	txRepo     repository.TransactionRepository
}

func NewWalletService(walletRepo repository.WalletRepository, txRepo repository.TransactionRepository) WalletService {
	return &walletService{walletRepo: walletRepo, txRepo: txRepo}
}

func (s *walletService) GetBalance(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	return s.walletRepo.GetByUserID(ctx, userID)
}

func (s *walletService) InitiateDeposit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, gateway string) (*models.Transaction, error) {
	tx := &models.Transaction{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      "deposit",
		Amount:    amount,
		Gateway:   &gateway,
		Status:    "pending",
	}
	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *walletService) UploadReceipt(ctx context.Context, txID uuid.UUID, receiptURL string) error {
	// Status becomes pending_review after upload
	return s.txRepo.UpdateReceipt(ctx, txID, receiptURL, "pending_review")
}

func (s *walletService) RequestWithdraw(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, gateway string) (*models.Transaction, error) {
	// Check if balance enough
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet.Balance.LessThan(amount) {
		return nil, apperr.ErrInsufficientBalance
	}

	tx := &models.Transaction{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    "withdraw",
		Amount:  amount,
		Gateway: &gateway,
		Status:  "pending_review",
	}
	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *walletService) GetTransactions(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Transaction, int, error) {
	return s.txRepo.ListPaginated(ctx, page, perPage, "", &userID)
}

func (s *walletService) GetTransaction(ctx context.Context, userID uuid.UUID, txID uuid.UUID) (*models.Transaction, error) {
	return s.txRepo.FindByID(ctx, txID, &userID)
}
