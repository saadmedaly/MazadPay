package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/database"
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
	// GetTransactionAny récupère une transaction sans filtrer par user_id — réservé aux
	// appelants qui ont déjà vérifié séparément que l'appelant est admin (audit sécurité,
	// utilisé par l'endpoint receipt-url présigné).
	GetTransactionAny(ctx context.Context, txID uuid.UUID) (*models.Transaction, error)
	GetPaymentMethods(ctx context.Context) ([]models.PaymentMethod, error)
}

type walletService struct {
	db         *sqlx.DB
	walletRepo repository.WalletRepository
	txRepo     repository.TransactionRepository
	notifSvc   NotificationService
	auditSvc   AuditService
}

func NewWalletService(db *sqlx.DB, walletRepo repository.WalletRepository, txRepo repository.TransactionRepository, notifSvc NotificationService, auditSvc AuditService) WalletService {
	return &walletService{db: db, walletRepo: walletRepo, txRepo: txRepo, notifSvc: notifSvc, auditSvc: auditSvc}
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
	if err := s.txRepo.UpdateReceipt(ctx, txID, userID, receiptURL, "pending_review"); err != nil {
		return err
	}
	if s.auditSvc != nil {
		// Ne jamais journaliser l'URL du reçu elle-même (audit de sécurité — voir
		// models.Transaction.ReceiptURL) : uniquement le fait qu'un reçu a été déposé.
		s.auditSvc.Log(ctx, userID, "receipt_uploaded", "transaction", &txID, "Receipt uploaded by user")
	}
	return nil
}

// RequestWithdraw crée une demande de retrait et gèle immédiatement le montant
// (balance -> frozen_amount) dans la même transaction SQL, avec row locking via
// FreezeForWithdraw (compare-and-set atomique). Ceci empêche qu'un utilisateur ne
// soumette plusieurs demandes de retrait concurrentes dont la somme dépasse son
// solde réel (audit de sécurité V09).
func (s *walletService) RequestWithdraw(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, gateway string) (*models.Transaction, error) {
	// Vérification défensive indépendante du handler (audit de sécurité V08) : un montant
	// négatif ou nul contournait le contrôle de solde (balance >= montant négatif est
	// toujours vrai) et pouvait, une fois approuvé côté admin, augmenter le solde au lieu
	// de le débiter.
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, apperr.ErrBadRequest
	}

	txModel := &models.Transaction{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    "withdraw",
		Amount:  amount,
		Gateway: &gateway,
		Status:  "pending_review",
	}

	err := database.WithTransaction(s.db, func(tx *sqlx.Tx) error {
		// Gèle le montant : balance -= amount, frozen_amount += amount (atomique,
		// échoue avec ErrInsufficientBalance si balance < amount).
		if err := s.walletRepo.FreezeForWithdraw(ctx, tx, userID, amount); err != nil {
			return err
		}
		if err := s.txRepo.CreateTx(ctx, tx, txModel); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if s.auditSvc != nil {
		details := fmt.Sprintf("user_id=%s amount=%s status=%s (balance frozen)", userID, amount.String(), txModel.Status)
		s.auditSvc.Log(ctx, userID, "withdraw_requested_funds_frozen", "transaction", &txModel.ID, details)
	}
	return txModel, nil
}

func (s *walletService) GetTransactions(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Transaction, int, error) {
	return s.txRepo.ListPaginated(ctx, page, perPage, "", &userID)
}

func (s *walletService) GetTransaction(ctx context.Context, userID uuid.UUID, txID uuid.UUID) (*models.Transaction, error) {
	return s.txRepo.FindByID(ctx, txID, &userID)
}

func (s *walletService) GetTransactionAny(ctx context.Context, txID uuid.UUID) (*models.Transaction, error) {
	return s.txRepo.FindByID(ctx, txID, nil)
}
