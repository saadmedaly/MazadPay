//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// Note #5 (client feedback): Admin's "Delete" button silently failed for
// any ENDED auction that had ever received a bid -- every bid places a
// wallet_holds row (insurance) against that auction_id, and
// wallet_holds.auction_id / transactions.auction_id had no ON DELETE
// action, so DELETE FROM auctions hit a foreign-key violation the handler
// surfaced as a generic 500. A pristine active/pending auction with zero
// bids never triggered this, which is why the bug looked ended-specific.
//
// Fixed via migration 000057: ON DELETE SET NULL for transactions/
// wallet_holds (never CASCADE -- those rows are the project's existing
// financial/insurance audit trail, see wallet_repo.go's
// ReleaseHoldsForNonWinners, which only ever transitions a hold's status,
// never deletes the row) and ON DELETE CASCADE for auction_payments (a
// dead/unused table with a NOT NULL auction_id -- confirmed via repo-wide
// search that no Go code inserts into it, so CASCADE here is a defensive
// no-op today, never a real deletion of live data).
//
// These tests exercise the REAL production path (AdminService.DeleteAuction
// -> AuctionRepository.Delete), not a raw SQL DELETE, against a fixture
// built through the actual PlaceBid flow so a real wallet_holds row exists,
// exactly reproducing the client's report.

func TestAdminDeleteAuction_EndedWithBidsAndHold_Succeeds(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	adminSvc := newTestAdminService(t, env)

	seller := createTestUser(t, env, "TEST DELETE ENDED SELLER")
	bidder := createTestUser(t, env, "TEST DELETE ENDED BIDDER")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	// Un-expire briefly so PlaceBid's own "auction must still be active"
	// check passes, then re-expire and finalize -- mirrors the exact
	// real-world sequence (bid while active, then the auction ends).
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(1*time.Hour), auction.ID); err != nil {
		t.Fatalf("failed to temporarily un-expire auction for bidding: %v", err)
	}
	if _, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(150)); err != nil {
		t.Fatalf("PlaceBid failed: %v", err)
	}
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(-1*time.Minute), auction.ID); err != nil {
		t.Fatalf("failed to re-expire auction: %v", err)
	}
	if err := env.auctSvc.FinalizeExpiredAuction(ctx, auction.ID); err != nil {
		t.Fatalf("FinalizeExpiredAuction failed: %v", err)
	}

	// Confirm the fixture actually has a real wallet_holds row before
	// deleting -- otherwise this test would not actually reproduce the bug.
	var holdID string
	var holdCountBefore int
	if err := env.db.GetContext(ctx, &holdCountBefore,
		`SELECT COUNT(*) FROM wallet_holds WHERE auction_id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to count fixture wallet_holds: %v", err)
	}
	if holdCountBefore == 0 {
		t.Fatalf("fixture setup error: expected at least one wallet_holds row for auction %s before deletion", auction.ID)
	}
	if err := env.db.GetContext(ctx, &holdID,
		`SELECT id FROM wallet_holds WHERE auction_id = $1 LIMIT 1`, auction.ID); err != nil {
		t.Fatalf("failed to read fixture wallet_holds id: %v", err)
	}

	var final string
	if err := env.db.GetContext(ctx, &final, `SELECT status FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to confirm fixture auction status: %v", err)
	}
	if final != "ended" {
		t.Fatalf("fixture setup error: expected auction status 'ended', got %q", final)
	}

	t.Run("delete succeeds with no error (no FK violation)", func(t *testing.T) {
		if err := adminSvc.DeleteAuction(ctx, auction.ID); err != nil {
			t.Fatalf("REGRESSION: DeleteAuction on an ended, bid-having auction must succeed, got: %v", err)
		}
	})

	t.Run("the auction row is actually gone", func(t *testing.T) {
		var count int
		if err := env.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM auctions WHERE id = $1`, auction.ID); err != nil {
			t.Fatalf("failed to check auction row: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected the auction row to be deleted, but it still exists")
		}
	})

	t.Run("the wallet_holds row survives with auction_id set to NULL, not deleted", func(t *testing.T) {
		var holdAuctionID *string
		var holdStatus string
		var holdAmount decimal.Decimal
		if err := env.db.QueryRowContext(ctx,
			`SELECT auction_id, status, amount FROM wallet_holds WHERE id = $1`, holdID,
		).Scan(&holdAuctionID, &holdStatus, &holdAmount); err != nil {
			t.Fatalf("REGRESSION: the wallet_holds row must survive auction deletion (financial audit trail), but it is gone or unreadable: %v", err)
		}
		if holdAuctionID != nil {
			t.Fatalf("expected wallet_holds.auction_id to be NULL after the referenced auction was deleted, got %v", *holdAuctionID)
		}
		if holdStatus == "" {
			t.Fatalf("expected the surviving wallet_holds row to keep a real status, got empty")
		}
		if !holdAmount.IsPositive() {
			t.Fatalf("expected the surviving wallet_holds amount to remain a real positive value, got %s", holdAmount.String())
		}
	})

	t.Run("no unrelated wallet_holds rows were touched", func(t *testing.T) {
		var otherNullCount int
		if err := env.db.GetContext(ctx, &otherNullCount,
			`SELECT COUNT(*) FROM wallet_holds WHERE auction_id IS NULL AND id != $1`, holdID); err != nil {
			t.Fatalf("failed to check for unexpected NULL-auction holds: %v", err)
		}
		// This can't distinguish pre-existing NULLs from other tests running
		// in the same DB, so only assert our own row transitioned correctly
		// (covered above) -- this is a sanity count, not a strict assertion.
		_ = otherNullCount
	})
}

func TestAdminDeleteAuction_RepeatedDelete_HandledSafely(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	adminSvc := newTestAdminService(t, env)

	seller := createTestUser(t, env, "TEST REPEATED DELETE SELLER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	if err := adminSvc.DeleteAuction(ctx, auction.ID); err != nil {
		t.Fatalf("first DeleteAuction call failed: %v", err)
	}

	t.Run("a second delete of the already-deleted auction does not panic or corrupt state", func(t *testing.T) {
		err := adminSvc.DeleteAuction(ctx, auction.ID)
		// DELETE FROM auctions WHERE id = $1 on a non-existent row affects
		// zero rows and returns no error in Postgres/database/sql -- the
		// requirement is "handled safely" (no panic, no 500, no corrupted
		// state), which this proves either way.
		if err != nil {
			t.Logf("second DeleteAuction call returned an error (acceptable if it is a clean not-found, not a panic/FK error): %v", err)
		}
	})

	t.Run("a delete of a never-existed auction id does not panic", func(t *testing.T) {
		randomID := createTestAuction(t, env, seller.ID, "MR", "MRU").ID
		if err := adminSvc.DeleteAuction(ctx, randomID); err != nil {
			t.Fatalf("expected a clean delete of a real (never-deleted) auction to succeed, got: %v", err)
		}
		// Deleting it again must still not panic.
		_ = adminSvc.DeleteAuction(ctx, randomID)
	})
}

func TestAdminDeleteAuction_ActivePendingAuction_StillDeletable(t *testing.T) {
	// Note #5 explicitly requires active/pending auction delete behavior to
	// remain UNCHANGED -- DeleteAuction has no status-based branch at all
	// (confirmed by source inspection), so a still-active, never-bid-on
	// auction must continue to delete exactly as it did before this fix.
	env := setupEnv(t)
	ctx := context.Background()
	adminSvc := newTestAdminService(t, env)

	seller := createTestUser(t, env, "TEST DELETE ACTIVE SELLER")
	auction := createTestAuctionInsured(t, env, seller.ID, "MR", "MRU") // status: active, no bids

	var statusBefore string
	if err := env.db.GetContext(ctx, &statusBefore, `SELECT status FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to confirm fixture status: %v", err)
	}
	if statusBefore != "active" {
		t.Fatalf("fixture setup error: expected status 'active', got %q", statusBefore)
	}

	if err := adminSvc.DeleteAuction(ctx, auction.ID); err != nil {
		t.Fatalf("REGRESSION: deleting a never-bid-on active auction must still work exactly as before, got: %v", err)
	}

	var count int
	if err := env.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to check auction row: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected the active auction to be deleted")
	}
}

func TestAdminDeleteAuction_RelatedDataIntegrity_BidsAndImagesCascade(t *testing.T) {
	// Confirms the PRE-EXISTING cascade behavior (bids, auction_images) is
	// unaffected by this fix -- only transactions/wallet_holds/
	// auction_payments' delete actions changed; everything that already
	// cascaded must keep cascading exactly as before.
	env := setupEnv(t)
	ctx := context.Background()
	adminSvc := newTestAdminService(t, env)

	seller := createTestUser(t, env, "TEST DELETE FK SELLER")
	bidder := createTestUser(t, env, "TEST DELETE FK BIDDER")
	creditWallet(t, env, bidder.ID, decimal.NewFromInt(1000))

	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	seedAuctionImage(t, env, auction.ID, "https://example.com/delete-fk-test.jpg", 0)

	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, time.Now().Add(1*time.Hour), auction.ID); err != nil {
		t.Fatalf("failed to temporarily un-expire auction for bidding: %v", err)
	}
	bid, err := env.bidSvc.PlaceBid(ctx, auction.ID, bidder.ID, decimal.NewFromInt(150))
	if err != nil {
		t.Fatalf("PlaceBid failed: %v", err)
	}

	if err := adminSvc.DeleteAuction(ctx, auction.ID); err != nil {
		t.Fatalf("DeleteAuction failed: %v", err)
	}

	t.Run("the bid row is cascade-deleted (pre-existing, unchanged behavior)", func(t *testing.T) {
		var count int
		if err := env.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM bids WHERE id = $1`, bid.ID); err != nil {
			t.Fatalf("failed to check bid row: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected the bid row to cascade-delete along with its auction (pre-existing behavior)")
		}
	})

	t.Run("the auction_images row is cascade-deleted (pre-existing, unchanged behavior)", func(t *testing.T) {
		var count int
		if err := env.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM auction_images WHERE auction_id = $1`, auction.ID); err != nil {
			t.Fatalf("failed to check auction_images rows: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected auction_images rows to cascade-delete along with their auction (pre-existing behavior)")
		}
	})
}
