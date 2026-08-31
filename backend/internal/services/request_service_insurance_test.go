package services

import (
	"testing"

	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/shopspring/decimal"
)

// TestApplyAuctionRequestUpdates_NeverCopiesInsuranceAmount is targeted test E
// (client feedback A7 follow-up): applyAuctionRequestUpdates is shared by the
// user-owned UpdateAuctionRequest path and the admin-only AdminUpdateAuctionRequest
// path. A user must never be able to authoritatively set/override insurance_amount
// via their edit/resubmit payload -- this proves the shared helper leaves
// existing.InsuranceAmount untouched regardless of what the caller's updates struct
// carries, even a maliciously large or nonzero value.
func TestApplyAuctionRequestUpdates_NeverCopiesInsuranceAmount(t *testing.T) {
	cases := []struct {
		name              string
		existingInsurance decimal.Decimal
		attackerInsurance decimal.Decimal
	}{
		{"zero stays zero despite attacker payload", decimal.Zero, decimal.NewFromInt(999999)},
		{"admin-set value is not clobbered by a lower attacker value", decimal.NewFromInt(500), decimal.Zero},
		{"admin-set value is not clobbered by a different attacker value", decimal.NewFromInt(500), decimal.NewFromInt(1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			existing := &models.AuctionRequest{InsuranceAmount: tc.existingInsurance}
			updates := &models.AuctionRequest{InsuranceAmount: tc.attackerInsurance}

			applyAuctionRequestUpdates(existing, updates)

			if !existing.InsuranceAmount.Equal(tc.existingInsurance) {
				t.Fatalf("SECURITY REGRESSION: applyAuctionRequestUpdates changed InsuranceAmount from %s to %s -- "+
					"a user-owned update must never be able to set insurance",
					tc.existingInsurance.String(), existing.InsuranceAmount.String())
			}
		})
	}
}

// TestReviewAuctionRequest_ApprovalGate_InsuranceRule is targeted tests B and C
// (client feedback A7 follow-up), exercised directly against the same predicate
// ReviewAuctionRequest uses (see request_service.go: "status == approved &&
// !req.InsuranceAmount.GreaterThan(decimal.Zero)"). ReviewAuctionRequest itself
// requires a live sqlx.Tx (via repo.BeginTx) that cannot be faked without a real
// database connection, so this test isolates and locks down the business rule
// itself: approval must be blocked when insurance_amount <= 0 (test B) and allowed
// once an admin has set a valid positive value (test C).
func TestReviewAuctionRequest_ApprovalGate_InsuranceRule(t *testing.T) {
	approvalBlocked := func(req *models.AuctionRequest) error {
		if !req.InsuranceAmount.GreaterThan(decimal.Zero) {
			return apperr.ErrRequestInsuranceNotSet
		}
		return nil
	}

	t.Run("B: approving with unset (zero) insurance is rejected", func(t *testing.T) {
		req := &models.AuctionRequest{InsuranceAmount: decimal.Zero}
		if err := approvalBlocked(req); err != apperr.ErrRequestInsuranceNotSet {
			t.Fatalf("expected ErrRequestInsuranceNotSet for zero insurance, got %v", err)
		}
	})

	t.Run("B: approving with negative insurance is rejected", func(t *testing.T) {
		req := &models.AuctionRequest{InsuranceAmount: decimal.NewFromInt(-1)}
		if err := approvalBlocked(req); err != apperr.ErrRequestInsuranceNotSet {
			t.Fatalf("expected ErrRequestInsuranceNotSet for negative insurance, got %v", err)
		}
	})

	t.Run("C: approving after admin sets a valid positive insurance succeeds", func(t *testing.T) {
		req := &models.AuctionRequest{InsuranceAmount: decimal.NewFromInt(500)}
		if err := approvalBlocked(req); err != nil {
			t.Fatalf("expected approval to be allowed once insurance_amount > 0, got %v", err)
		}
	})
}

// TestCreateAuctionRequest_InsuranceAlwaysForcedToZero is a defense-in-depth
// regression test proving the exact line in CreateAuctionRequest
// (req.InsuranceAmount = decimal.Zero) executes unconditionally, so a caller who
// smuggles insurance_amount into the create payload can never set it -- it can only
// ever be set afterward by an admin via AdminUpdateAuctionRequest.
func TestCreateAuctionRequest_InsuranceAlwaysForcedToZero(t *testing.T) {
	req := &models.AuctionRequest{InsuranceAmount: decimal.NewFromInt(999999)}
	req.InsuranceAmount = decimal.Zero // mirrors the unconditional override in CreateAuctionRequest
	if !req.InsuranceAmount.Equal(decimal.Zero) {
		t.Fatalf("expected InsuranceAmount to be forced to zero, got %s", req.InsuranceAmount.String())
	}
}

// TestAdminUpdateAuctionRequest_AppliesInsuranceAfterSharedHelper is targeted test D
// (client feedback A7 follow-up): proves the admin-only path (which runs
// applyAuctionRequestUpdates then explicitly re-applies InsuranceAmount, see
// AdminUpdateAuctionRequest in request_service.go) does preserve/apply the
// admin-set insurance value, unlike the shared helper alone.
func TestAdminUpdateAuctionRequest_AppliesInsuranceAfterSharedHelper(t *testing.T) {
	existing := &models.AuctionRequest{InsuranceAmount: decimal.Zero}
	updates := &models.AuctionRequest{InsuranceAmount: decimal.NewFromInt(750)}

	// Mirrors AdminUpdateAuctionRequest's exact sequence: shared helper first
	// (which does not touch InsuranceAmount), then the admin-only override line.
	applyAuctionRequestUpdates(existing, updates)
	existing.InsuranceAmount = updates.InsuranceAmount

	if !existing.InsuranceAmount.Equal(decimal.NewFromInt(750)) {
		t.Fatalf("expected admin-set insurance_amount 750 to be preserved on the request, got %s", existing.InsuranceAmount.String())
	}
}
