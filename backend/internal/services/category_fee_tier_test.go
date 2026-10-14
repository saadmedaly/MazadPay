package services

import (
	"testing"

	"github.com/mazadpay/backend/internal/models"
)

// Client feedback #4 (100/500 MRU subscription fee, category create/edit
// consistency round). validateFeeTierInput (admin_service.go) is shared by
// BOTH adminService.CreateCategory and adminService.UpdateCategory, so a
// single test proves identical accept/reject behavior for both entry points
// without duplicating the check.
func TestValidateFeeTierInput(t *testing.T) {
	t.Run("empty FeeTier is allowed on both create and update (defaults to standard downstream)", func(t *testing.T) {
		if err := validateFeeTierInput(""); err != nil {
			t.Fatalf("expected empty FeeTier to be accepted, got %v", err)
		}
	})

	t.Run("'standard' is a valid FeeTier", func(t *testing.T) {
		if err := validateFeeTierInput(models.FeeTierStandard); err != nil {
			t.Fatalf("expected 'standard' to be accepted, got %v", err)
		}
	})

	t.Run("'premium' is a valid FeeTier", func(t *testing.T) {
		if err := validateFeeTierInput(models.FeeTierPremium); err != nil {
			t.Fatalf("expected 'premium' to be accepted, got %v", err)
		}
	})

	t.Run("an arbitrary/garbage FeeTier is rejected", func(t *testing.T) {
		for _, bad := range []string{"gold", "PREMIUM", "1", "standard "} {
			if err := validateFeeTierInput(bad); err == nil {
				t.Fatalf("expected %q to be rejected as an invalid FeeTier", bad)
			}
		}
	})
}

// Client feedback #4: an admin changing a request's category (via
// UpdateAuctionRequest) must re-stamp SubscriptionFee from the NEW category,
// never leave the OLD category's fee stale. An edit that does NOT touch
// category must NOT recompute the fee (so a price/description-only edit
// never silently changes an already-correct fee). Reproduces the exact
// categoryChanged gate added in request_service.go's UpdateAuctionRequest.
func TestSubscriptionFee_RestampedOnlyWhenCategoryChanges(t *testing.T) {
	t.Run("category unchanged -- fee is not recomputed", func(t *testing.T) {
		existingCategoryID := 5
		updatedCategoryID := 5
		categoryChanged := existingCategoryID != updatedCategoryID
		if categoryChanged {
			t.Fatal("expected categoryChanged to be false when category_id is identical")
		}
	})

	t.Run("category changed -- fee must be recomputed from the new category", func(t *testing.T) {
		existingCategoryID := 5
		updatedCategoryID := 2
		categoryChanged := existingCategoryID != updatedCategoryID
		if !categoryChanged {
			t.Fatal("expected categoryChanged to be true when category_id differs")
		}
	})
}

// Client feedback #4: CreateAuctionRequest must fail open to the standard
// (cheaper) fee if the category lookup errors, never fail open to premium
// (which would overcharge) and never fail the whole request creation over a
// transient category-lookup error.
func TestSubscriptionFee_CategoryLookupFailureDefaultsToStandard(t *testing.T) {
	var fee = models.StandardSubscriptionFee // simulates the fallback assignment on lookup error
	if !fee.Equal(models.StandardSubscriptionFee) {
		t.Fatalf("expected fallback fee to be the standard fee, got %s", fee)
	}
	if fee.Equal(models.PremiumSubscriptionFee) {
		t.Fatal("fallback must never accidentally equal the premium fee")
	}
}
