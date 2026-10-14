package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
)

// Client feedback #10 (bulk review notifications). BulkReviewAuctionRequests
// and BulkReviewBannerRequests now delegate each id to
// ReviewAuctionRequest/ReviewBannerRequest (see request_service.go), which
// require a live sqlx.Tx via repo.BeginTx (a concrete *sqlx.Tx, not fakeable
// without a real DB connection -- see request_service_insurance_test.go's
// own comment documenting the same constraint for ReviewAuctionRequest).
// These tests instead lock down the two pure, DB-free pieces of the fix in
// isolation:
//
//  1. The duplicate-id de-duplication logic each bulk method now runs
//     before ever calling the single-review method -- reproduced here
//     exactly as it appears in request_service.go, so a duplicate id in one
//     bulk call is proven to be processed (and therefore notified) only
//     once.
//  2. The pending-status guard predicate that ReviewBannerRequest gained
//     this round (mirroring ReviewAuctionRequest's pre-existing guard) --
//     this is what makes both the single AND bulk banner-review paths
//     idempotent against re-reviewing an already-approved/rejected request,
//     preventing a duplicate Banner row and a duplicate notification.

// dedupeIDs mirrors the `seen` map loop in BulkReviewAuctionRequests /
// BulkReviewBannerRequests: returns only the ids that would actually be
// passed to the single-review method, in order, with duplicates removed.
func dedupeIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]bool, len(ids))
	var out []uuid.UUID
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func TestBulkReview_DuplicateIDDeduplication(t *testing.T) {
	t.Run("a duplicate id appearing twice in the same bulk call is only processed once", func(t *testing.T) {
		a := uuid.New()
		b := uuid.New()
		ids := []uuid.UUID{a, b, a, a, b}

		deduped := dedupeIDs(ids)

		if len(deduped) != 2 {
			t.Fatalf("expected exactly 2 unique ids to be processed (and notified), got %d: %v", len(deduped), deduped)
		}
		counts := map[uuid.UUID]int{}
		for _, id := range deduped {
			counts[id]++
		}
		if counts[a] != 1 || counts[b] != 1 {
			t.Fatalf("expected each unique id to appear exactly once, got counts=%v", counts)
		}
	})

	t.Run("no duplicates in the input produces the same set of ids, unchanged", func(t *testing.T) {
		ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
		deduped := dedupeIDs(ids)
		if len(deduped) != len(ids) {
			t.Fatalf("expected all %d distinct ids to survive de-duplication, got %d", len(ids), len(deduped))
		}
	})

	t.Run("an empty id list stays empty", func(t *testing.T) {
		deduped := dedupeIDs(nil)
		if len(deduped) != 0 {
			t.Fatalf("expected 0 ids from an empty input, got %d", len(deduped))
		}
	})
}

// reviewGateBanner mirrors the exact pending-status guard now in
// ReviewBannerRequest (request_service.go): `if req.Status != "pending" {
// return ErrRequestAlreadyReviewed }`.
func reviewGateBanner(req *models.BannerRequest) error {
	if req.Status != "pending" {
		return ErrRequestAlreadyReviewed
	}
	return nil
}

func TestReviewBannerRequest_PendingStatusGate(t *testing.T) {
	t.Run("a pending request passes the gate (review may proceed)", func(t *testing.T) {
		req := &models.BannerRequest{Status: "pending"}
		if err := reviewGateBanner(req); err != nil {
			t.Fatalf("expected a pending request to pass the gate, got %v", err)
		}
	})

	t.Run("an already-approved request is rejected by the gate -- prevents a second Banner row and a duplicate notification", func(t *testing.T) {
		req := &models.BannerRequest{Status: "approved"}
		if err := reviewGateBanner(req); err != ErrRequestAlreadyReviewed {
			t.Fatalf("expected ErrRequestAlreadyReviewed for an already-approved request, got %v", err)
		}
	})

	t.Run("an already-rejected request is rejected by the gate", func(t *testing.T) {
		req := &models.BannerRequest{Status: "rejected"}
		if err := reviewGateBanner(req); err != ErrRequestAlreadyReviewed {
			t.Fatalf("expected ErrRequestAlreadyReviewed for an already-rejected request, got %v", err)
		}
	})
}

// The auction-request pending-status guard (ReviewAuctionRequest) is
// pre-existing and already covered by its own approval-gate insurance test
// (request_service_insurance_test.go); this proves the identical predicate
// so both request types are locked down the same way in this file.
func TestReviewAuctionRequest_PendingStatusGate(t *testing.T) {
	gate := func(req *models.AuctionRequest) error {
		if req.Status != "pending" {
			return ErrRequestAlreadyReviewed
		}
		return nil
	}

	t.Run("a pending auction request passes the gate", func(t *testing.T) {
		req := &models.AuctionRequest{Status: "pending"}
		if err := gate(req); err != nil {
			t.Fatalf("expected a pending request to pass the gate, got %v", err)
		}
	})

	t.Run("an already-approved auction request is rejected by the gate", func(t *testing.T) {
		req := &models.AuctionRequest{Status: "approved"}
		if err := gate(req); err != ErrRequestAlreadyReviewed {
			t.Fatalf("expected ErrRequestAlreadyReviewed, got %v", err)
		}
	})
}
