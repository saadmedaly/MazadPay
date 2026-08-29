package services

import "testing"

// TestPubliclyVisibleAuctionStatuses is the regression test for the security fix in
// auction_handler.go's public GET /auctions/:id and GET /auctions: before the fix,
// neither endpoint filtered by status at all, so an anonymous visitor could view a
// pending/rejected/canceled auction by guessing/reusing its ID, or bulk-enumerate them
// via an arbitrary ?status= query parameter with zero authentication. Both handlers now
// consult this exact map — this test locks its contents so a future change can't
// silently widen public visibility to a non-public status without a deliberate edit here.
func TestPubliclyVisibleAuctionStatuses(t *testing.T) {
	mustBePublic := []string{"active", "ended"}
	mustNotBePublic := []string{"pending", "rejected", "canceled", "draft", "", "ACTIVE", "Active"}

	for _, status := range mustBePublic {
		if !PubliclyVisibleAuctionStatuses[status] {
			t.Errorf("status %q must be publicly visible but PubliclyVisibleAuctionStatuses says it is not", status)
		}
	}

	for _, status := range mustNotBePublic {
		if PubliclyVisibleAuctionStatuses[status] {
			t.Errorf("status %q must NOT be publicly visible (this is the exact bug the fix closes) but PubliclyVisibleAuctionStatuses says it is", status)
		}
	}

	// Exact size check: catches an accidental addition of a new status to the whitelist
	// without an explicit, deliberate test update alongside it.
	if len(PubliclyVisibleAuctionStatuses) != 2 {
		t.Errorf("PubliclyVisibleAuctionStatuses has %d entries, want exactly 2 (active, ended) — a status was added or removed without updating this test", len(PubliclyVisibleAuctionStatuses))
	}
}
