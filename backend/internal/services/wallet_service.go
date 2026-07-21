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
	InitiateDeposit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, gateway, paymentMethod, receiptImageTemp string) (*models.Transaction, error)
	UploadReceipt(ctx context.Context, txID uuid.UUID, userID uuid.UUID, receiptURL string) error
	RequestWithdraw(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, gateway string) (*models.Transaction, error)
	GetTransactions(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Transaction, int, error)
	GetTransaction(ctx context.Context, userID uuid.UUID, txID uuid.UUID) (*models.Transaction, error)
	GetPaymentMethods(ctx context.Context) ([]models.PaymentMethod, error)
}

type walletService struct {
	walletRepo repository.WalletRepository
	txRepo     repository.TransactionRepository
	notifSvc   NotificationService
}

func NewWalletService(walletRepo repository.WalletRepository, txRepo repository.TransactionRepository, notifSvc NotificationService) WalletService {
	return &walletService{walletRepo: walletRepo, txRepo: txRepo, notifSvc: notifSvc}
}

func (s *walletService) GetBalance(ctx context.Context, userID uuid.UUID) (*models.Wallet, error) {
	return s.walletRepo.GetByUserID(ctx, userID)
}

func (s *walletService) GetPaymentMethods(ctx context.Context) ([]models.PaymentMethod, error) {
	return s.walletRepo.GetPaymentMethods(ctx)
}

func (s *walletService) InitiateDeposit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, gateway, paymentMethod, receiptImageTemp string) (*models.Transaction, error) {
	// Vérification défensive indépendante du handler (audit de sécurité V05-bis) :
	// ce service ne doit jamais faire confiance uniquement à la validation côté handler/mobile.
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperr.ErrBadRequest
	}
	tx := &models.Transaction{
		ID:                 uuid.New(),
		UserID:             userID,
		Type:               "deposit",
		Amount:             amount,
		Gateway:            &gateway,
		Status:             "pending",
		PaymentMethod:      &paymentMethod,
		ReceiptImageTemp:   &receiptImageTemp,
	}
	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *walletService) UploadReceipt(ctx context.Context, txID uuid.UUID, userID uuid.UUID, receiptURL string) error {
	return s.txRepo.UpdateReceipt(ctx, txID, userID, receiptURL, "pending_review")
}

func (s *walletService) RequestWithdraw(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, gateway string) (*models.Transaction, error) {
	// Vérification défensive indépendante du handler (audit de sécurité V08) : un montant
	// négatif ou nul contournait le contrôle de solde ci-dessous (balance >= montant négatif
	// est toujours vrai) et pouvait, une fois approuvé côté admin, augmenter le solde au lieu
	// de le débiter (voir transaction_repo.go: balance = balance - amount).
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperr.ErrBadRequest
	}
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
