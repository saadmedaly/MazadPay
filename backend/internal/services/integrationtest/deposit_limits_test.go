//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
)

// Client feedback (deposit min/max limits): the deposit request screen's
// "تعليمات الدفع" (payment instructions) box previously showed no min/max
// amount at all, and source inspection confirmed the backend never enforced
// any either -- WalletHandler.Deposit's validator tag was only "gt=0", and
// WalletService.InitiateDeposit's own defensive check matched. Fixed by
// adding services.MinDepositAmountMRU (100) / MaxDepositAmountMRU (100000),
// enforced ONLY for the generic wallet-top-up deposit path
// (auctionRequestID == nil) and ONLY for MRU-denominated wallets -- the
// auction-subscription path's amount is server-stamped (subscription_fee),
// never client-supplied, so it is structurally exempt; a non-MRU wallet
// (e.g. MAD/TND) is exempt too since the client's exact reference values
// (MRU 100 / MRU 100,000) are Mauritania-specific.
//
// These tests exercise the REAL production path
// (WalletService.InitiateDeposit), not a raw SQL insert.

func TestInitiateDeposit_MinAmount_Accepted(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	walletSvc := newWalletSvc(env)
	user := createTestUser(t, env, "TEST DEPOSIT MIN ACCEPTED") // MR -> MRU

	txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(int64(services.MinDepositAmountMRU)), "bankily", "bankily", "", nil)
	if err != nil {
		t.Fatalf("expected exactly the minimum deposit amount (100 MRU) to be accepted, got: %v", err)
	}
	if !txn.Amount.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("expected stored amount 100, got %s", txn.Amount.String())
	}
}

func TestInitiateDeposit_MaxAmount_Accepted(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	walletSvc := newWalletSvc(env)
	user := createTestUser(t, env, "TEST DEPOSIT MAX ACCEPTED") // MR -> MRU

	txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(int64(services.MaxDepositAmountMRU)), "bankily", "bankily", "", nil)
	if err != nil {
		t.Fatalf("expected exactly the maximum deposit amount (100,000 MRU) to be accepted, got: %v", err)
	}
	if !txn.Amount.Equal(decimal.NewFromInt(100000)) {
		t.Fatalf("expected stored amount 100000, got %s", txn.Amount.String())
	}
}

func TestInitiateDeposit_BelowMinAmount_Rejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	walletSvc := newWalletSvc(env)
	user := createTestUser(t, env, "TEST DEPOSIT BELOW MIN") // MR -> MRU

	_, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(99), "bankily", "bankily", "", nil)
	if err != apperr.ErrDepositTooLow {
		t.Fatalf("expected ErrDepositTooLow for a 99 MRU deposit, got: %v", err)
	}
}

func TestInitiateDeposit_AboveMaxAmount_Rejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	walletSvc := newWalletSvc(env)
	user := createTestUser(t, env, "TEST DEPOSIT ABOVE MAX") // MR -> MRU

	_, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(100001), "bankily", "bankily", "", nil)
	if err != apperr.ErrDepositTooHigh {
		t.Fatalf("expected ErrDepositTooHigh for a 100,001 MRU deposit, got: %v", err)
	}
}

func TestInitiateDeposit_AuctionSubscriptionPath_ExemptFromLimits(t *testing.T) {
	// The auction-subscription path substitutes the client-supplied amount
	// with the request's own server-stamped subscription_fee -- the client
	// value sent here (1, far below the 100 MRU minimum) is what would be
	// rejected on the generic top-up path, proving the min/max check is
	// skipped entirely for auctionRequestID != nil rather than merely
	// happening to pass (the request's real fee, 100, is what actually gets
	// stored, and 100 itself is only at the boundary -- the client-sent 1 is
	// the value that would fail if this path were NOT exempt).
	env := setupEnv(t)
	ctx := context.Background()
	walletSvc := newWalletSvc(env)
	user := createTestUser(t, env, "TEST DEPOSIT SUBSCRIPTION EXEMPT")
	standardCategoryID := createTestCategory(t, env, models.FeeTierStandard)

	req := newAuctionRequest(user.ID, "deposit-exempt-"+uuid.New().String()[:6])
	req.CategoryID = standardCategoryID
	if err := env.reqSvc.CreateAuctionRequest(ctx, req); err != nil {
		t.Fatalf("CreateAuctionRequest failed: %v", err)
	}

	txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(1), "bankily", "bankily", "", &req.ID)
	if err != nil {
		t.Fatalf("expected the auction-subscription path to be exempt from deposit min/max limits (client sent 1, would fail on the generic path), got: %v", err)
	}
	if !txn.Amount.Equal(models.StandardSubscriptionFee) {
		t.Fatalf("expected the server-stamped subscription_fee to be stored, got %s", txn.Amount.String())
	}
}

func TestInitiateDeposit_NonMRUWallet_ExemptFromMRULimits(t *testing.T) {
	// The client's exact reference values (MRU 100 / MRU 100,000) are
	// Mauritania-specific -- a non-MRU wallet must not be bound by them.
	env := setupEnv(t)
	ctx := context.Background()
	walletSvc := newWalletSvc(env)
	user := createTestUserWithCountry(t, env, "TEST DEPOSIT NON MRU EXEMPT", "MA")

	txn, err := walletSvc.InitiateDeposit(ctx, user.ID, decimal.NewFromInt(10), "bank_transfer", "bank_transfer", "", nil)
	if err != nil {
		t.Fatalf("expected a non-MRU (MAD) deposit below 100 to be unaffected by the MRU-scoped min/max limits, got: %v", err)
	}
	if txn.CurrencyCode == nil || *txn.CurrencyCode != "MAD" {
		t.Fatalf("expected currency_code MAD, got %v", txn.CurrencyCode)
	}
}
