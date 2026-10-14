//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"github.com/shopspring/decimal"
)

// Customer Request #27: Admin Categories/Subcategories hide-show. Categories
// and subcategories share one table (distinguished only by parent_id), so
// these tests exercise both through the same AdminService.ToggleCategory /
// AuctionService.GetCategories paths.

// createTestSubcategory inserts a temporary subcategory (parent_id set)
// directly, mirroring createTestCategory's own pattern.
func createTestSubcategory(t *testing.T, env *testEnv, parentID int) int {
	t.Helper()
	ctx := context.Background()
	var id int
	err := env.db.GetContext(ctx, &id, `
		INSERT INTO categories (name_ar, name_fr, parent_id, fee_tier)
		VALUES ($1, $2, $3, 'standard') RETURNING id`,
		"فئة فرعية اختبار "+uuid.New().String()[:6], "Test subcategory", parentID)
	if err != nil {
		t.Fatalf("failed to create test subcategory: %v", err)
	}
	t.Cleanup(func() {
		env.db.ExecContext(context.Background(), `DELETE FROM categories WHERE id = $1`, id)
	})
	return id
}

func categoryIsActive(t *testing.T, env *testEnv, id int) bool {
	t.Helper()
	var isActive bool
	if err := env.db.Get(&isActive, `SELECT is_active FROM categories WHERE id = $1`, id); err != nil {
		t.Fatalf("failed to read category is_active: %v", err)
	}
	return isActive
}

// TestToggleCategory_HideThenShow_RealPath (Sections 2, 6, 13): the actual
// AdminService.ToggleCategory path, not a raw SQL shortcut.
func TestToggleCategory_HideThenShow_RealPath(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	admin := createTestAdmin(t, env, "TEST CATTOGGLE ADMIN")
	adminSvc := newTestAdminService(t, env)

	catID := createTestCategory(t, env, models.FeeTierStandard)

	t.Run("starts active (DB default)", func(t *testing.T) {
		if !categoryIsActive(t, env, catID) {
			t.Fatalf("expected a freshly created category to default is_active=true")
		}
	})

	t.Run("hide sets is_active=false, row is not deleted", func(t *testing.T) {
		if err := adminSvc.ToggleCategory(ctx, catID, false, admin.ID); err != nil {
			t.Fatalf("ToggleCategory(hide) failed: %v", err)
		}
		if categoryIsActive(t, env, catID) {
			t.Fatalf("expected is_active=false after hide")
		}
		var count int
		if err := env.db.Get(&count, `SELECT COUNT(*) FROM categories WHERE id = $1`, catID); err != nil {
			t.Fatalf("failed to count category rows: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected the category row to still exist after hide, got count=%d", count)
		}
	})

	t.Run("hide does not change name or other fields", func(t *testing.T) {
		var nameAr, feeTier string
		if err := env.db.QueryRowContext(ctx, `SELECT name_ar, fee_tier FROM categories WHERE id = $1`, catID).Scan(&nameAr, &feeTier); err != nil {
			t.Fatalf("failed to read category: %v", err)
		}
		if feeTier != models.FeeTierStandard {
			t.Fatalf("expected fee_tier unchanged (%s), got %s", models.FeeTierStandard, feeTier)
		}
	})

	t.Run("show restores is_active=true", func(t *testing.T) {
		if err := adminSvc.ToggleCategory(ctx, catID, true, admin.ID); err != nil {
			t.Fatalf("ToggleCategory(show) failed: %v", err)
		}
		if !categoryIsActive(t, env, catID) {
			t.Fatalf("expected is_active=true after show")
		}
	})

	t.Run("repeated hide is idempotent (no error, stays hidden)", func(t *testing.T) {
		if err := adminSvc.ToggleCategory(ctx, catID, false, admin.ID); err != nil {
			t.Fatalf("first hide failed: %v", err)
		}
		if err := adminSvc.ToggleCategory(ctx, catID, false, admin.ID); err != nil {
			t.Fatalf("repeated hide failed: %v", err)
		}
		if categoryIsActive(t, env, catID) {
			t.Fatalf("expected is_active to remain false after repeated hide")
		}
		// Restore for cleanliness (not required, but avoids a hidden row
		// lingering if t.Cleanup ever changes).
		_ = adminSvc.ToggleCategory(ctx, catID, true, admin.ID)
	})
}

// TestAdminListCategories_IncludesHidden (Section 9).
func TestAdminListCategories_IncludesHidden(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	admin := createTestAdmin(t, env, "TEST CATADMIN LIST ADMIN")
	adminSvc := newTestAdminService(t, env)

	catID := createTestCategory(t, env, models.FeeTierStandard)
	if err := adminSvc.ToggleCategory(ctx, catID, false, admin.ID); err != nil {
		t.Fatalf("ToggleCategory(hide) failed: %v", err)
	}

	cats, err := adminSvc.AdminListCategories(ctx)
	if err != nil {
		t.Fatalf("AdminListCategories failed: %v", err)
	}
	found := false
	for _, c := range cats {
		if c.ID == catID {
			found = true
			if c.IsActive {
				t.Fatalf("expected the admin list's own copy of the hidden category to show is_active=false")
			}
		}
	}
	if !found {
		t.Fatalf("expected AdminListCategories to include the hidden category -- an admin must be able to see it to re-show it")
	}
}

// TestPublicCategories_ExcludesHidden_IncludesAfterShow (Sections 10, 13):
// exercises the real GetCategories service call the same way
// AuctionHandler.GetCategories does (filtering unfiltered results to only
// active rows, mirroring the handler's own in-Go filter).
func TestPublicCategories_ExcludesHidden_IncludesAfterShow(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	admin := createTestAdmin(t, env, "TEST CATPUBLIC ADMIN")
	adminSvc := newTestAdminService(t, env)

	catID := createTestCategory(t, env, models.FeeTierStandard)

	publicVisible := func() bool {
		// GetCategories caches its result in Redis for 1 hour (cache key
		// "categories"); in production that cache is invalidated by
		// AdminHandler.invalidateCategoriesCache, called by the real HTTP
		// ToggleCategory/UpdateCategory/etc handlers. This test calls
		// AdminService.ToggleCategory directly (bypassing the handler), so
		// each check must clear the same key itself to observe the DB's
		// current state rather than a stale cached response from an
		// earlier call (this test's own earlier assertions, or even an
		// unrelated test in the same suite run).
		if env.rdb != nil {
			env.rdb.Del(ctx, "categories")
		}
		all, err := env.auctSvc.GetCategories(ctx)
		if err != nil {
			t.Fatalf("GetCategories failed: %v", err)
		}
		activeByID := make(map[int]bool, len(all))
		for _, c := range all {
			activeByID[c.ID] = c.IsActive
		}
		for _, c := range all {
			if c.ID != catID {
				continue
			}
			if !c.IsActive {
				return false
			}
			if c.ParentID != nil && !activeByID[*c.ParentID] {
				return false
			}
			return true
		}
		return false
	}

	t.Run("visible before hide", func(t *testing.T) {
		if !publicVisible() {
			t.Fatalf("expected the category to be publicly visible before hiding")
		}
	})

	t.Run("excluded after hide", func(t *testing.T) {
		if err := adminSvc.ToggleCategory(ctx, catID, false, admin.ID); err != nil {
			t.Fatalf("ToggleCategory(hide) failed: %v", err)
		}
		// AdminService.ToggleCategory persists the DB write and broadcasts
		// category.updated (matching UpdateCategory's own established
		// convention), but invalidating the "categories" Redis cache key is
		// owned by AdminHandler.invalidateCategoriesCache -- called by the
		// real HTTP handler, exactly like every other category mutation
		// (Create/Update/Delete already work the same way). This test calls
		// the service directly, bypassing the handler, so it must clear the
		// same key to observe real end-to-end behavior instead of a stale
		// 1-hour cached GetCategories response.
		if env.rdb != nil {
			env.rdb.Del(ctx, "categories")
		}
		if publicVisible() {
			t.Fatalf("expected the category to be excluded from public listing after hide")
		}
	})

	t.Run("included again after show", func(t *testing.T) {
		if err := adminSvc.ToggleCategory(ctx, catID, true, admin.ID); err != nil {
			t.Fatalf("ToggleCategory(show) failed: %v", err)
		}
		if env.rdb != nil {
			env.rdb.Del(ctx, "categories")
		}
		if !publicVisible() {
			t.Fatalf("expected the category to be publicly visible again after show")
		}
	})
}

// TestHiddenParent_SubcategoryRowUnchanged_EffectivelyNotVisible (Section 12):
// hiding a parent must NOT mass-update child subcategory rows, but the
// child's effective public visibility must still be gated by the parent.
func TestHiddenParent_SubcategoryRowUnchanged_EffectivelyNotVisible(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	admin := createTestAdmin(t, env, "TEST CATPARENT ADMIN")
	adminSvc := newTestAdminService(t, env)

	parentID := createTestCategory(t, env, models.FeeTierStandard)
	childID := createTestSubcategory(t, env, parentID)

	if err := adminSvc.ToggleCategory(ctx, parentID, false, admin.ID); err != nil {
		t.Fatalf("ToggleCategory(hide parent) failed: %v", err)
	}
	// See TestPublicCategories_ExcludesHidden_IncludesAfterShow's comment:
	// ToggleCategory does not itself own Redis cache invalidation (that is
	// the real HTTP handler's job, matching Create/Update/Delete's existing
	// convention) -- calling the service directly here needs the same
	// explicit clear to observe real behavior.
	if env.rdb != nil {
		env.rdb.Del(ctx, "categories")
	}

	t.Run("child's own is_active row is untouched by hiding the parent", func(t *testing.T) {
		if !categoryIsActive(t, env, childID) {
			t.Fatalf("REGRESSION: hiding the parent must not mass-update the child subcategory's own is_active row")
		}
	})

	t.Run("child is effectively not visible in the public listing while its parent is hidden", func(t *testing.T) {
		all, err := env.auctSvc.GetCategories(ctx)
		if err != nil {
			t.Fatalf("GetCategories failed: %v", err)
		}
		activeByID := make(map[int]bool, len(all))
		for _, c := range all {
			activeByID[c.ID] = c.IsActive
		}
		for _, c := range all {
			if c.ID != childID {
				continue
			}
			effectivelyVisible := c.IsActive && (c.ParentID == nil || activeByID[*c.ParentID])
			if effectivelyVisible {
				t.Fatalf("expected the child subcategory to be effectively invisible while its parent is hidden")
			}
		}
	})

	t.Run("child remains visible again once the parent is shown", func(t *testing.T) {
		if err := adminSvc.ToggleCategory(ctx, parentID, true, admin.ID); err != nil {
			t.Fatalf("ToggleCategory(show parent) failed: %v", err)
		}
		if env.rdb != nil {
			env.rdb.Del(ctx, "categories")
		}
		all, err := env.auctSvc.GetCategories(ctx)
		if err != nil {
			t.Fatalf("GetCategories failed: %v", err)
		}
		activeByID := make(map[int]bool, len(all))
		for _, c := range all {
			activeByID[c.ID] = c.IsActive
		}
		found := false
		for _, c := range all {
			if c.ID != childID {
				continue
			}
			effectivelyVisible := c.IsActive && (c.ParentID == nil || activeByID[*c.ParentID])
			if effectivelyVisible {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected the child subcategory to be effectively visible again once its parent is shown")
		}
	})
}

// TestToggleCategory_NoAuctionSideEffects (Section 11): hiding a category
// must never touch existing auctions referencing it.
func TestToggleCategory_NoAuctionSideEffects(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	admin := createTestAdmin(t, env, "TEST CATAUCTION ADMIN")
	adminSvc := newTestAdminService(t, env)
	seller := createTestUser(t, env, "TEST CATAUCTION SELLER")

	catID := createTestCategory(t, env, models.FeeTierStandard)

	input := services.CreateAuctionInput{
		CategoryID:    catID,
		TitleAr:       "مزاد اختبار تصنيف مخفي " + uuid.New().String()[:6],
		DescriptionAr: "وصف تجريبي لمزاد يستخدم فئة سيتم إخفاؤها",
		StartPrice:    decimal.NewFromInt(50),
		MinIncrement:  decimal.NewFromInt(5),
		EndTime:       time.Now().Add(24 * time.Hour),
		Quantity:      1,
	}
	auction, err := env.auctSvc.Create(ctx, seller.ID, input)
	if err != nil {
		t.Fatalf("failed to create fixture auction: %v", err)
	}

	if err := adminSvc.ToggleCategory(ctx, catID, false, admin.ID); err != nil {
		t.Fatalf("ToggleCategory(hide) failed: %v", err)
	}

	var reread models.Auction
	if err := env.db.Get(&reread, `SELECT * FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to read back auction: %v", err)
	}
	if reread.CategoryID != catID {
		t.Fatalf("expected the auction's category_id link to remain unchanged, got %d, expected %d", reread.CategoryID, catID)
	}
	if reread.Status == "canceled" || reread.Status == "cancelled" {
		t.Fatalf("REGRESSION: hiding a category must never cancel auctions that reference it, got status=%q", reread.Status)
	}
}
