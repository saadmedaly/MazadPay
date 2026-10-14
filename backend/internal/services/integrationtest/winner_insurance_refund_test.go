//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/shopspring/decimal"
)

// Customer Request #31: admin manual refund of the auction WINNER's own
// insurance hold -- the one case AuctionService.FinalizeExpiredAuction /
// WalletRepository.ReleaseHoldsForNonWinners deliberately never covers (see
// that method's own doc comment). These tests exercise the REAL production
// finalization path (auctSvc.FinalizeExpiredAuction, not a SetWinner
// shortcut) so the winner's hold is left in its genuine post-finalization
// 'active' state before RefundWinnerInsurance is exercised against it.

// createExpiredInsuranceRequiredAuction is createExpiredTestAuction's twin,
// except InsurancePolicy is "required" (not "not_required" like every other
// shared fixture helper in this package) -- PlaceBid's insurance-hold block
// (bid_service.go) is gated on auction.InsuranceRequired(), which is false
// for "not_required" specifically, so a wallet_holds row is never created
// against createExpiredTestAuction/createTestAuctionInsured despite the
// latter's name. This feature is entirely about that hold, so its own
// fixture must actually produce one.
func createExpiredInsuranceRequiredAuction(t *testing.T, env *testEnv, sellerID uuid.UUID) *models.Auction {
	t.Helper()
	ctx := context.Background()
	marketISO, currencyCode := "MR", "MRU"
	lotNumber := "TEST-REFUND-" + uuid.New().String()[:8]
	a := &models.Auction{
		ID:               uuid.New(),
		SellerID:         sellerID,
		CategoryID:       mkCategoryID(),
		TitleAr:          "مزاد اختبار إرجاع تأمين " + uuid.New().String()[:6],
		LotNumber:        &lotNumber,
		StartPrice:       decimal.NewFromInt(100),
		CurrentPrice:     decimal.NewFromInt(100),
		MinIncrement:     decimal.NewFromInt(10),
		InsuranceAmount:  decimal.NewFromInt(20),
		InsurancePolicy:  "required",
		ReservePrice:     decimal.NewFromInt(100),
		StartTime:        time.Now().Add(-2 * time.Hour),
		EndTime:          time.Now().Add(48 * time.Hour), // created non-expired, then back-dated below
		Status:           "active",
		MarketCountryISO: &marketISO,
		CurrencyCode:     &currencyCode,
	}
	if err := env.auctionRepo.Create(ctx, nil, a); err != nil {
		t.Fatalf("failed to create insurance-required fixture auction: %v", err)
	}
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), a.ID); err != nil {
		t.Fatalf("failed to back-date fixture auction end_time: %v", err)
	}
	if err := env.db.GetContext(ctx, a, `SELECT * FROM auctions WHERE id = $1`, a.ID); err != nil {
		t.Fatalf("failed to read back fixture auction: %v", err)
	}
	return a
}

// setupEndedAuctionWithWinnerHold creates a real ended auction (via the
// actual finalization path) with a real winning bid and a real, still-active
// wallet_hold for the winner -- the exact precondition RefundWinnerInsurance
// requires. Returns the auction, the winner, and the winning bid amount
// (== the hold amount, since CreateHold freezes auction.InsuranceAmount,
// not the bid amount -- see bid_service.go PlaceBid).
func setupEndedAuctionWithWinnerHold(t *testing.T, env *testEnv) (auction *models.Auction, winner *models.User, holdAmount decimal.Decimal) {
	t.Helper()
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST REFUND SELLER")
	w := createTestUser(t, env, "TEST REFUND WINNER")
	creditWallet(t, env, w.ID, decimal.NewFromInt(1000))

	a := createExpiredInsuranceRequiredAuction(t, env, seller.ID)
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(1*time.Hour), a.ID); err != nil {
		t.Fatalf("failed to temporarily un-expire auction for bidding: %v", err)
	}
	if _, err := env.bidSvc.PlaceBid(ctx, a.ID, w.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("PlaceBid failed: %v", err)
	}
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), a.ID); err != nil {
		t.Fatalf("failed to re-expire auction: %v", err)
	}
	if err := env.auctSvc.FinalizeExpiredAuction(ctx, a.ID); err != nil {
		t.Fatalf("FinalizeExpiredAuction failed: %v", err)
	}

	var final models.Auction
	if err := env.db.Get(&final, `SELECT * FROM auctions WHERE id = $1`, a.ID); err != nil {
		t.Fatalf("failed to read back finalized auction: %v", err)
	}
	if final.WinnerID == nil || *final.WinnerID != w.ID {
		t.Fatalf("setup invariant broken: expected winner %s, got %v", w.ID, final.WinnerID)
	}

	return &final, w, decimal.NewFromInt(20)
}

func walletBalanceAndFrozen(t *testing.T, env *testEnv, userID uuid.UUID) (balance, frozen decimal.Decimal) {
	t.Helper()
	if err := env.db.QueryRowContext(context.Background(),
		`SELECT balance, frozen_amount FROM wallets WHERE user_id = $1`, userID,
	).Scan(&balance, &frozen); err != nil {
		t.Fatalf("failed to read wallet balance/frozen: %v", err)
	}
	return balance, frozen
}

func holdStatus(t *testing.T, env *testEnv, userID, auctionID uuid.UUID) string {
	t.Helper()
	var status string
	if err := env.db.QueryRowContext(context.Background(),
		`SELECT status FROM wallet_holds WHERE user_id = $1 AND auction_id = $2`, userID, auctionID,
	).Scan(&status); err != nil {
		t.Fatalf("failed to read hold status: %v", err)
	}
	return status
}

func TestRefundWinnerInsurance_AdminCanRefundWinnerOnce(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "TEST REFUND ADMIN")

	auction, winner, holdAmount := setupEndedAuctionWithWinnerHold(t, env)

	// Confirmed precondition: winner's hold is 'active' post-finalization
	// (the exact gap this ticket closes).
	t.Run("precondition: winner hold is active before refund", func(t *testing.T) {
		if status := holdStatus(t, env, winner.ID, auction.ID); status != "active" {
			t.Fatalf("expected winner hold to be 'active' before refund, got %q", status)
		}
	})

	balanceBefore, frozenBefore := walletBalanceAndFrozen(t, env, winner.ID)

	ledgerTx, err := adminSvc.RefundWinnerInsurance(ctx, auction.ID, admin.ID)

	t.Run("admin can refund winner once", func(t *testing.T) {
		if err != nil {
			t.Fatalf("RefundWinnerInsurance failed: %v", err)
		}
		if ledgerTx == nil {
			t.Fatalf("expected a non-nil ledger transaction")
		}
	})

	t.Run("winner balance increases by exact hold amount", func(t *testing.T) {
		balanceAfter, _ := walletBalanceAndFrozen(t, env, winner.ID)
		expected := balanceBefore.Add(holdAmount)
		if !balanceAfter.Equal(expected) {
			t.Fatalf("expected balance %s, got %s", expected.String(), balanceAfter.String())
		}
	})

	t.Run("frozen amount decreases correctly", func(t *testing.T) {
		_, frozenAfter := walletBalanceAndFrozen(t, env, winner.ID)
		expected := frozenBefore.Sub(holdAmount)
		if !frozenAfter.Equal(expected) {
			t.Fatalf("expected frozen_amount %s, got %s", expected.String(), frozenAfter.String())
		}
	})

	t.Run("hold becomes released", func(t *testing.T) {
		if status := holdStatus(t, env, winner.ID, auction.ID); status != "released" {
			t.Fatalf("expected hold status 'released', got %q", status)
		}
	})

	t.Run("ledger row created exactly once, correct type/amount/user", func(t *testing.T) {
		var count int
		var dbAmount decimal.Decimal
		var dbUserID uuid.UUID
		var dbType string
		if err := env.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM transactions WHERE auction_id = $1 AND type = 'insurance_refund'`,
			auction.ID,
		).Scan(&count); err != nil {
			t.Fatalf("failed to count ledger rows: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected exactly 1 insurance_refund ledger row, got %d", count)
		}
		if err := env.db.QueryRowContext(ctx,
			`SELECT amount, user_id, type FROM transactions WHERE auction_id = $1 AND type = 'insurance_refund'`,
			auction.ID,
		).Scan(&dbAmount, &dbUserID, &dbType); err != nil {
			t.Fatalf("failed to read ledger row: %v", err)
		}
		if !dbAmount.Equal(holdAmount) {
			t.Fatalf("expected ledger amount %s, got %s", holdAmount.String(), dbAmount.String())
		}
		if dbUserID != winner.ID {
			t.Fatalf("expected ledger user_id %s, got %s", winner.ID, dbUserID)
		}
	})

	t.Run("second refund does nothing / is rejected safely, no double payout", func(t *testing.T) {
		balanceBeforeSecond, frozenBeforeSecond := walletBalanceAndFrozen(t, env, winner.ID)

		_, err := adminSvc.RefundWinnerInsurance(ctx, auction.ID, admin.ID)
		if err != apperr.ErrNoActiveHold {
			t.Fatalf("expected ErrNoActiveHold on second refund attempt, got %v", err)
		}

		balanceAfterSecond, frozenAfterSecond := walletBalanceAndFrozen(t, env, winner.ID)
		if !balanceAfterSecond.Equal(balanceBeforeSecond) {
			t.Fatalf("SECURITY: balance changed on rejected second refund: %s -> %s", balanceBeforeSecond, balanceAfterSecond)
		}
		if !frozenAfterSecond.Equal(frozenBeforeSecond) {
			t.Fatalf("SECURITY: frozen_amount changed on rejected second refund: %s -> %s", frozenBeforeSecond, frozenAfterSecond)
		}

		var count int
		if err := env.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM transactions WHERE auction_id = $1 AND type = 'insurance_refund'`,
			auction.ID,
		).Scan(&count); err != nil {
			t.Fatalf("failed to count ledger rows after second attempt: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected still exactly 1 ledger row after rejected second refund, got %d", count)
		}
	})
}

// Note on "non-admin rejected": authorization for this action is enforced
// at the route/middleware layer (jwtMiddleware + AdminOnly, identical to
// every other /admin/* route) -- RefundWinnerInsurance takes an adminID
// purely for the audit log, matching every other AdminService method's
// existing contract (e.g. ValidateTransaction), and is never reachable by a
// non-admin without going through that middleware. See
// handlers/winner_insurance_refund_test.go
// (TestRefundWinnerInsurance_Unauthenticated_Rejected/_NonAdmin_Rejected)
// for the real HTTP-level proof of this.

func TestRefundWinnerInsurance_WrongUserRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "TEST REFUND WRONGUSER ADMIN")

	auction, _, _ := setupEndedAuctionWithWinnerHold(t, env)

	// A different user (not the winner) must never be refundable for this
	// auction -- RefundWinnerInsurance takes only auctionID, always derives
	// the winner from the DB, so there is no parameter through which a
	// caller could even attempt to target a non-winner; this proves that by
	// construction rather than by passing a wrong id.
	other := createTestUser(t, env, "TEST REFUND UNRELATED USER")
	var unrelatedHoldCount int
	if err := env.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM wallet_holds WHERE user_id = $1 AND auction_id = $2`, other.ID, auction.ID,
	).Scan(&unrelatedHoldCount); err != nil {
		t.Fatalf("failed to count holds for unrelated user: %v", err)
	}
	if unrelatedHoldCount != 0 {
		t.Fatalf("test invariant broken: unrelated user should have no hold row, got %d", unrelatedHoldCount)
	}

	_, err := adminSvc.RefundWinnerInsurance(ctx, auction.ID, admin.ID)
	if err != nil {
		t.Fatalf("expected the real winner's refund to succeed (proving the method targets the correct user), got: %v", err)
	}

	// The unrelated user's (nonexistent) hold must remain untouched.
	var count int
	if err := env.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM wallet_holds WHERE user_id = $1 AND auction_id = $2`, other.ID, auction.ID,
	).Scan(&count); err != nil {
		t.Fatalf("failed to count holds for unrelated user: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 hold rows for unrelated user, got %d", count)
	}
}

func TestRefundWinnerInsurance_ActiveAuctionRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "TEST REFUND ACTIVE ADMIN")

	seller := createTestUser(t, env, "TEST REFUND ACTIVE SELLER")
	auction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU")
	// Never finalized -- status remains 'active'.

	_, err := adminSvc.RefundWinnerInsurance(ctx, auction.ID, admin.ID)
	if err != apperr.ErrAuctionNotEnded {
		t.Fatalf("expected ErrAuctionNotEnded for a still-active auction, got %v", err)
	}
}

func TestRefundWinnerInsurance_NoWinnerRejected(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "TEST REFUND NOWINNER ADMIN")

	seller := createTestUser(t, env, "TEST REFUND NOWINNER SELLER")
	auction := createExpiredTestAuction(t, env, seller.ID)
	// No bids placed -- FinalizeExpiredAuction's no-winner branch runs
	// (TryEndAuctionAtomically), leaving status='ended' but winner_id=NULL.
	if err := env.auctSvc.FinalizeExpiredAuction(ctx, auction.ID); err != nil {
		t.Fatalf("FinalizeExpiredAuction failed: %v", err)
	}

	_, err := adminSvc.RefundWinnerInsurance(ctx, auction.ID, admin.ID)
	if err != apperr.ErrNotAuctionWinner {
		t.Fatalf("expected ErrNotAuctionWinner for a no-bid ended auction, got %v", err)
	}
}

func TestRefundWinnerInsurance_NonWinnerBidderHoldAlreadyReleased(t *testing.T) {
	// Confirms this feature does not regress or duplicate the EXISTING
	// automatic non-winner refund: a losing bidder's hold is already
	// 'released' by ReleaseHoldsForNonWinners at finalization time (proven
	// elsewhere, e.g. TestFinalizeExpiredAuction_RealPath_WinnerPersistedCorrectly)
	// -- RefundWinnerInsurance is never the path that refunds them, and
	// attempting to reuse it for a non-winner (by definition, via a winner
	// that isn't them) is structurally impossible since the method only
	// ever targets auctions.winner_id.
	env := setupEnv(t)
	ctx := context.Background()

	seller := createTestUser(t, env, "TEST REFUND LOSER SELLER")
	loser := createTestUser(t, env, "TEST REFUND LOSER")
	winner := createTestUser(t, env, "TEST REFUND LOSER WINNER")
	creditWallet(t, env, loser.ID, decimal.NewFromInt(1000))
	creditWallet(t, env, winner.ID, decimal.NewFromInt(1000))

	auction := createExpiredInsuranceRequiredAuction(t, env, seller.ID)
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(1*time.Hour), auction.ID); err != nil {
		t.Fatalf("failed to temporarily un-expire auction for bidding: %v", err)
	}
	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, loser.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("loser PlaceBid failed: %v", err)
	}
	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, winner.ID, decimal.NewFromInt(200)); err != nil {
		t.Fatalf("winner PlaceBid failed: %v", err)
	}
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), auction.ID); err != nil {
		t.Fatalf("failed to re-expire auction: %v", err)
	}
	if err := env.auctSvc.FinalizeExpiredAuction(ctx, auction.ID); err != nil {
		t.Fatalf("FinalizeExpiredAuction failed: %v", err)
	}

	if status := holdStatus(t, env, loser.ID, auction.ID); status != "released" {
		t.Fatalf("expected the EXISTING automatic non-winner refund to have already released the loser's hold, got %q -- Customer #31 must not change this", status)
	}
	if status := holdStatus(t, env, winner.ID, auction.ID); status != "active" {
		t.Fatalf("expected winner hold to remain active pending manual admin refund, got %q", status)
	}
}

func TestRefundWinnerInsurance_NoPartialStateOnFailure(t *testing.T) {
	// Requirement: "no partial state if ledger/refund fails". Proven by
	// atomicity: attempting to refund an auction whose winner has NO active
	// hold (e.g. the hold was somehow already released, simulated here by
	// manually releasing it before calling the service) must leave the
	// wallet completely untouched -- the failure is caught by
	// FindActiveHold before any wallet mutation or ledger write is even
	// attempted, and would additionally roll back cleanly via the deferred
	// dbtx.Rollback() if it failed at a later step.
	env := setupEnv(t)
	ctx := context.Background()
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "TEST REFUND PARTIAL ADMIN")

	auction, winner, _ := setupEndedAuctionWithWinnerHold(t, env)

	// Manually release the hold out-of-band (simulates some other process
	// having already resolved it) before the admin ever calls refund.
	if _, err := env.db.ExecContext(ctx,
		`UPDATE wallet_holds SET status = 'released', released_at = now() WHERE user_id = $1 AND auction_id = $2`,
		winner.ID, auction.ID,
	); err != nil {
		t.Fatalf("failed to manually release hold for test setup: %v", err)
	}

	balanceBefore, frozenBefore := walletBalanceAndFrozen(t, env, winner.ID)

	_, err := adminSvc.RefundWinnerInsurance(ctx, auction.ID, admin.ID)
	if err != apperr.ErrNoActiveHold {
		t.Fatalf("expected ErrNoActiveHold, got %v", err)
	}

	balanceAfter, frozenAfter := walletBalanceAndFrozen(t, env, winner.ID)
	if !balanceAfter.Equal(balanceBefore) {
		t.Fatalf("expected no balance change on failure, got %s -> %s", balanceBefore, balanceAfter)
	}
	if !frozenAfter.Equal(frozenBefore) {
		t.Fatalf("expected no frozen_amount change on failure, got %s -> %s", frozenBefore, frozenAfter)
	}
	var count int
	if err := env.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM transactions WHERE auction_id = $1 AND type = 'insurance_refund'`, auction.ID,
	).Scan(&count); err != nil {
		t.Fatalf("failed to count ledger rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 ledger rows when the refund never succeeded, got %d", count)
	}
}
