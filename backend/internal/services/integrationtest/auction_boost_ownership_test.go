//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Security fix SEC-03: auction_boost_handler.go's CreateBoost and
// CancelBoost previously had no ownership check at all (CreateBoost only
// checked market isolation; CancelBoost checked nothing) -- any
// authenticated same-market user could boost or cancel a boost on an
// auction they didn't own. Both handlers now require auction.SellerID ==
// the authenticated caller's ID, returning 404 (matching this handler
// file's established anti-enumeration convention, same as the pre-existing
// cross-market denial) for both a nonexistent resource and one that exists
// but isn't the caller's.
//
// TestCreateBoost_SameMarket_Succeeds and the new
// TestCreateBoost_NonOwnerSameMarket_Denied (integration_test.go) cover
// CreateBoost's owner/non-owner cases directly at the HTTP level using the
// existing httpCreateBoost helper -- this file covers CancelBoost plus the
// remaining required scenarios (unauthenticated, nonexistent, and the
// "cannot bypass via a different boost/auction ID" case).

// seedBoostRow inserts a boost row directly (bypassing the HTTP layer,
// which is what's under test) so tests have a real, owned boost to try to
// cancel.
func seedBoostRow(t *testing.T, env *testEnv, auctionID uuid.UUID) uuid.UUID {
	t.Helper()
	var boostID uuid.UUID
	err := env.db.GetContext(context.Background(), &boostID, `
		INSERT INTO auction_boosts (auction_id, boost_type, start_at, end_at, status)
		VALUES ($1, 'featured', $2, $3, 'active')
		RETURNING id
	`, auctionID, time.Now(), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("failed to seed boost row: %v", err)
	}
	return boostID
}

// 3. Owner cancels their own boost -> PASS.
func TestCancelBoost_Owner_Succeeds(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST CANCELBOOST OWNER SELLER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	boostID := seedBoostRow(t, env, auction.ID)

	status := httpCancelBoost(t, env, auction.ID, boostID, seller.ID)
	if status != 200 {
		t.Fatalf("(3) expected 200 for the owner cancelling their own boost, got %d", status)
	}
	if got := getBoostStatus(t, env, boostID); got != "cancelled" {
		t.Fatalf("(3) expected boost status='cancelled' after owner cancel, got %q", got)
	}
}

// 4. A different authenticated user cancelling someone else's boost -> DENIED.
func TestCancelBoost_NonOwner_Denied(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST CANCELBOOST NONOWNER SELLER")
	otherUser := createTestUser(t, env, "TEST CANCELBOOST NONOWNER OTHER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	boostID := seedBoostRow(t, env, auction.ID)

	status := httpCancelBoost(t, env, auction.ID, boostID, otherUser.ID)
	if status != 404 {
		t.Fatalf("(4) SEC-03 REGRESSION: expected 404 for a non-owner cancelling another user's boost, got %d", status)
	}
	if got := getBoostStatus(t, env, boostID); got != "active" {
		t.Fatalf("(4) SEC-03 REGRESSION: boost status changed to %q after a denied cancel attempt -- must remain 'active'", got)
	}
}

// 5. Unauthenticated request -> DENIED.
func TestCancelBoost_Unauthenticated_Denied(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST CANCELBOOST UNAUTH SELLER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")
	boostID := seedBoostRow(t, env, auction.ID)

	status := httpCancelBoostUnauthenticated(t, env, auction.ID, boostID)
	if status != 401 {
		t.Fatalf("(5) expected 401 for an unauthenticated cancel attempt, got %d", status)
	}
	if got := getBoostStatus(t, env, boostID); got != "active" {
		t.Fatalf("(5) boost status changed to %q after an unauthenticated cancel attempt -- must remain 'active'", got)
	}
}

// 6. Non-existent boost ID -> expected project behavior (404, matching
// every other not-found case in this handler file).
func TestCancelBoost_NonexistentBoost_ReturnsNotFound(t *testing.T) {
	env := setupEnv(t)
	seller := createTestUser(t, env, "TEST CANCELBOOST NOTFOUND SELLER")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	status := httpCancelBoost(t, env, auction.ID, uuid.New(), seller.ID)
	if status != 404 {
		t.Fatalf("(6) expected 404 for a nonexistent boost_id, got %d", status)
	}
}

// 7. Ownership cannot be bypassed by supplying a different, real boost_id
// belonging to a DIFFERENT auction the caller also doesn't own -- proves
// the check is tied to the specific boost's own auction, not just "any
// boost the caller supplies alongside any auction_id in the URL". The
// handler derives the owning auction from the boost itself (never trusts
// the URL's :id segment for authorization), so this also confirms a
// mismatched :id in the URL path has no effect on the outcome.
func TestCancelBoost_CannotBypassViaDifferentAuctionInURL(t *testing.T) {
	env := setupEnv(t)
	sellerA := createTestUser(t, env, "TEST CANCELBOOST BYPASS SELLER A")
	sellerB := createTestUser(t, env, "TEST CANCELBOOST BYPASS SELLER B")
	auctionA := createTestAuction(t, env, sellerA.ID, "MR", "MRU")
	auctionB := createTestAuction(t, env, sellerB.ID, "MR", "MRU")
	boostOnA := seedBoostRow(t, env, auctionA.ID)

	// sellerB is not the owner of auctionA or boostOnA -- attempting to
	// cancel boostOnA while putting auctionB's ID in the URL path must
	// still be denied (the handler must derive ownership from the boost's
	// real auction_id, not the URL's :id).
	status := httpCancelBoost(t, env, auctionB.ID, boostOnA, sellerB.ID)
	if status != 404 {
		t.Fatalf("(7) SEC-03 REGRESSION: expected 404 when a non-owner supplies a mismatched auction ID in the URL to try to cancel another auction's boost, got %d", status)
	}
	if got := getBoostStatus(t, env, boostOnA); got != "active" {
		t.Fatalf("(7) SEC-03 REGRESSION: boost status changed to %q via the URL-mismatch bypass attempt -- must remain 'active'", got)
	}

	// sellerA (the real owner) can still cancel it normally, proving the
	// denial above was genuinely ownership-based, not a broken route/lookup.
	okStatus := httpCancelBoost(t, env, auctionA.ID, boostOnA, sellerA.ID)
	if okStatus != 200 {
		t.Fatalf("(7) sanity check failed: the real owner could not cancel their own boost after the bypass attempt, got %d", okStatus)
	}
}
