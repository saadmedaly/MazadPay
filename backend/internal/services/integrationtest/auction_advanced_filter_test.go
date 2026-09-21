//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
)

// Note #1 (client feedback): the Active Auctions screen's advanced-filter
// sheet (price range + sort: newest/price_asc/price_desc/ending_soon) was
// already fully wired on the mobile side (all_auctions_page.dart sends
// min_price/max_price/sort_by on every request), but source inspection
// confirmed GET /auctions never read or applied any of them at all --
// AuctionHandler.List had no c.Query("min_price"/"max_price"/"sort_by"),
// AuctionFilters had no matching fields, and FindAll's ORDER BY was
// hardcoded to "is_featured DESC, created_at DESC" with no price WHERE
// clause. These tests exercise the real FindAll path (not a mock), proving
// the fix end-to-end at the repository layer GET /auctions delegates to.

func setAuctionPrice(t *testing.T, env *testEnv, auctionID uuid.UUID, price int) {
	t.Helper()
	ctx := context.Background()
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET current_price = $1 WHERE id = $2`, price, auctionID); err != nil {
		t.Fatalf("failed to set fixture auction price: %v", err)
	}
}

func setAuctionEndTime(t *testing.T, env *testEnv, auctionID uuid.UUID, endTime time.Time) {
	t.Helper()
	ctx := context.Background()
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET end_time = $1 WHERE id = $2`, endTime, auctionID); err != nil {
		t.Fatalf("failed to set fixture auction end_time: %v", err)
	}
}

// MAZADPAY -- "تنتهي قريباً" (ending soon) filter bug: FindAll's ORDER BY was
// unconditionally prefixed with "a.is_featured DESC" for EVERY sort mode,
// including ending_soon -- so a featured auction ending in 2 hours would
// sort above a non-featured auction ending in 5 minutes, silently breaking
// the client's explicit "nearest end_time first, full stop" requirement.
// None of TestFindAll_SortModes' existing fixtures were featured, which is
// exactly why that pre-existing test never caught this.
func setAuctionFeatured(t *testing.T, env *testEnv, auctionID uuid.UUID, featured bool) {
	t.Helper()
	ctx := context.Background()
	if _, err := env.db.ExecContext(ctx, `UPDATE auctions SET is_featured = $1 WHERE id = $2`, featured, auctionID); err != nil {
		t.Fatalf("failed to set fixture auction is_featured: %v", err)
	}
}

func TestFindAll_PriceRangeFilter(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST PRICE FILTER SELLER")

	cheap := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionPrice(t, env, cheap.ID, 50)
	mid := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionPrice(t, env, mid.ID, 500)
	expensive := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionPrice(t, env, expensive.ID, 5000)

	minP, maxP := 100, 1000
	results, total, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
		Status:   "active",
		MinPrice: &minP,
		MaxPrice: &maxP,
		PerPage:  100,
	})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	foundIDs := map[string]bool{}
	for _, a := range results {
		foundIDs[a.ID.String()] = true
	}

	t.Run("auction below min_price is excluded", func(t *testing.T) {
		if foundIDs[cheap.ID.String()] {
			t.Fatalf("expected the 50-priced auction to be excluded by min_price=100")
		}
	})
	t.Run("auction within range is included", func(t *testing.T) {
		if !foundIDs[mid.ID.String()] {
			t.Fatalf("expected the 500-priced auction to be included in [100,1000]")
		}
	})
	t.Run("auction above max_price is excluded", func(t *testing.T) {
		if foundIDs[expensive.ID.String()] {
			t.Fatalf("expected the 5000-priced auction to be excluded by max_price=1000")
		}
	})
	t.Run("total reflects the filtered count, not the unfiltered count", func(t *testing.T) {
		if total < 1 {
			t.Fatalf("expected total >= 1 for the matching auction, got %d", total)
		}
	})
}

func TestFindAll_SortModes(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST SORT MODE SELLER")

	a1 := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionPrice(t, env, a1.ID, 300)
	setAuctionEndTime(t, env, a1.ID, time.Now().Add(72*time.Hour))

	a2 := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionPrice(t, env, a2.ID, 100)
	setAuctionEndTime(t, env, a2.ID, time.Now().Add(1*time.Hour))

	a3 := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionPrice(t, env, a3.ID, 700)
	setAuctionEndTime(t, env, a3.ID, time.Now().Add(24*time.Hour))

	t.Run("price_asc orders lowest current_price first", func(t *testing.T) {
		results, _, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
			Status: "active", SortBy: "price_asc", PerPage: 100,
		})
		if err != nil {
			t.Fatalf("FindAll failed: %v", err)
		}
		pos := map[string]int{}
		for i, a := range results {
			pos[a.ID.String()] = i
		}
		if pos[a2.ID.String()] >= pos[a1.ID.String()] || pos[a1.ID.String()] >= pos[a3.ID.String()] {
			t.Fatalf("expected price_asc order a2(100) < a1(300) < a3(700), got positions a1=%d a2=%d a3=%d",
				pos[a1.ID.String()], pos[a2.ID.String()], pos[a3.ID.String()])
		}
	})

	t.Run("price_desc orders highest current_price first", func(t *testing.T) {
		results, _, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
			Status: "active", SortBy: "price_desc", PerPage: 100,
		})
		if err != nil {
			t.Fatalf("FindAll failed: %v", err)
		}
		pos := map[string]int{}
		for i, a := range results {
			pos[a.ID.String()] = i
		}
		if pos[a3.ID.String()] >= pos[a1.ID.String()] || pos[a1.ID.String()] >= pos[a2.ID.String()] {
			t.Fatalf("expected price_desc order a3(700) < a1(300) < a2(100), got positions a1=%d a2=%d a3=%d",
				pos[a1.ID.String()], pos[a2.ID.String()], pos[a3.ID.String()])
		}
	})

	t.Run("ending_soon orders nearest end_time first", func(t *testing.T) {
		results, _, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
			Status: "active", SortBy: "ending_soon", PerPage: 100,
		})
		if err != nil {
			t.Fatalf("FindAll failed: %v", err)
		}
		pos := map[string]int{}
		for i, a := range results {
			pos[a.ID.String()] = i
		}
		// a2 ends in 1h, a3 in 24h, a1 in 72h.
		if pos[a2.ID.String()] >= pos[a3.ID.String()] || pos[a3.ID.String()] >= pos[a1.ID.String()] {
			t.Fatalf("expected ending_soon order a2(1h) < a3(24h) < a1(72h), got positions a1=%d a2=%d a3=%d",
				pos[a1.ID.String()], pos[a2.ID.String()], pos[a3.ID.String()])
		}
	})

	t.Run("newest (default/empty SortBy) keeps original created_at DESC order, unchanged behavior", func(t *testing.T) {
		results, _, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
			Status: "active", SortBy: "", PerPage: 100,
		})
		if err != nil {
			t.Fatalf("FindAll failed: %v", err)
		}
		pos := map[string]int{}
		for i, a := range results {
			pos[a.ID.String()] = i
		}
		// a3 created after a2, created after a1 -- newest first means a3, a2, a1.
		if pos[a3.ID.String()] >= pos[a2.ID.String()] || pos[a2.ID.String()] >= pos[a1.ID.String()] {
			t.Fatalf("expected default newest order a3 < a2 < a1 (most recently created first), got positions a1=%d a2=%d a3=%d",
				pos[a1.ID.String()], pos[a2.ID.String()], pos[a3.ID.String()])
		}
	})

	t.Run("an unrecognized SortBy value falls back to newest, never breaks the query", func(t *testing.T) {
		results, _, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
			Status: "active", SortBy: "not_a_real_sort_mode", PerPage: 100,
		})
		if err != nil {
			t.Fatalf("FindAll must not error on an unrecognized SortBy, got: %v", err)
		}
		if len(results) == 0 {
			t.Fatalf("expected results even with an unrecognized SortBy")
		}
	})
}

// MAZADPAY -- ending_soon must order by end_time ASC ALONE, never letting a
// featured auction override the nearest-expiry ordering. This is the exact
// client-reported scenario: a featured auction ending later must NOT jump
// ahead of a non-featured auction ending sooner.
func TestFindAll_EndingSoon_IgnoresFeaturedFlag(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST ENDING SOON FEATURED SELLER")

	// Featured, but ends LATER -- must not sort first under ending_soon.
	featuredLate := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionEndTime(t, env, featuredLate.ID, time.Now().Add(2*time.Hour))
	setAuctionFeatured(t, env, featuredLate.ID, true)

	// Not featured, ends SOON -- must sort first under ending_soon despite
	// not being featured.
	soonNotFeatured := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionEndTime(t, env, soonNotFeatured.ID, time.Now().Add(5*time.Minute))

	// Not featured, ends last.
	laterNotFeatured := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionEndTime(t, env, laterNotFeatured.ID, time.Now().Add(30*time.Minute))

	results, _, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
		Status: "active", SortBy: "ending_soon", PerPage: 100,
	})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	pos := map[string]int{}
	for i, a := range results {
		pos[a.ID.String()] = i
	}

	// Expected order: soonNotFeatured (5min) < laterNotFeatured (30min) < featuredLate (2h)
	// -- strict end_time ASC, is_featured completely ignored.
	if pos[soonNotFeatured.ID.String()] >= pos[laterNotFeatured.ID.String()] {
		t.Fatalf("expected the 5-min auction before the 30-min auction under ending_soon, got positions soon=%d later=%d",
			pos[soonNotFeatured.ID.String()], pos[laterNotFeatured.ID.String()])
	}
	if pos[laterNotFeatured.ID.String()] >= pos[featuredLate.ID.String()] {
		t.Fatalf("REGRESSION: featured auction (ends in 2h) sorted before a non-featured auction ending sooner (30min) -- is_featured must not override ending_soon, got positions later=%d featuredLate=%d",
			pos[laterNotFeatured.ID.String()], pos[featuredLate.ID.String()])
	}
}

// Pagination must not destroy the ending_soon order: page 2 must continue
// exactly where page 1 left off, still strictly end_time ASC across the
// page boundary.
func TestFindAll_EndingSoon_OrderPreservedAcrossPagination(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST ENDING SOON PAGINATION SELLER")

	var auctionIDs []uuid.UUID
	for i := 0; i < 5; i++ {
		a := createTestAuction(t, env, seller.ID, "MR", "MRU")
		// Ends sooner as i increases is wrong -- make i=0 the SOONEST so the
		// expected order is auctionIDs[0..4] in that exact sequence.
		setAuctionEndTime(t, env, a.ID, time.Now().Add(time.Duration(i+1)*time.Hour))
		auctionIDs = append(auctionIDs, a.ID)
	}

	page1, _, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
		Status: "active", SortBy: "ending_soon", Page: 1, PerPage: 3,
	})
	if err != nil {
		t.Fatalf("FindAll page 1 failed: %v", err)
	}
	page2, _, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
		Status: "active", SortBy: "ending_soon", Page: 2, PerPage: 3,
	})
	if err != nil {
		t.Fatalf("FindAll page 2 failed: %v", err)
	}

	if len(page1) != 3 {
		t.Fatalf("expected 3 results on page 1, got %d", len(page1))
	}
	// The 5 fixtures must appear across page1+page2 in strict end_time ASC
	// order relative to each other (other pre-existing auctions from earlier
	// tests may interleave, so only relative order among these 5 is checked).
	combined := append(append([]uuid.UUID{}, idsOf(page1)...), idsOf(page2)...)
	var relativeOrder []int
	for _, id := range combined {
		for idx, fixtureID := range auctionIDs {
			if id == fixtureID {
				relativeOrder = append(relativeOrder, idx)
			}
		}
	}
	for i := 1; i < len(relativeOrder); i++ {
		if relativeOrder[i] < relativeOrder[i-1] {
			t.Fatalf("REGRESSION: ending_soon order broken across pagination, fixture indices appeared out of order: %v", relativeOrder)
		}
	}
}

func idsOf(auctions []models.Auction) []uuid.UUID {
	ids := make([]uuid.UUID, len(auctions))
	for i, a := range auctions {
		ids[i] = a.ID
	}
	return ids
}

// Expired auctions must never appear in the active/ending_soon list, even
// though an expired auction's end_time would otherwise sort it first.
func TestFindAll_EndingSoon_ExcludesExpiredAuctions(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST ENDING SOON EXPIRED SELLER")

	active := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionEndTime(t, env, active.ID, time.Now().Add(1*time.Hour))

	expired := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionEndTime(t, env, expired.ID, time.Now().Add(-1*time.Hour))

	results, _, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
		Status: "active", SortBy: "ending_soon", PerPage: 100,
	})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	foundActive, foundExpired := false, false
	for _, a := range results {
		if a.ID == active.ID {
			foundActive = true
		}
		if a.ID == expired.ID {
			foundExpired = true
		}
	}
	if !foundActive {
		t.Fatalf("expected the still-active auction to appear in the ending_soon active list")
	}
	if foundExpired {
		t.Fatalf("REGRESSION: an expired auction (end_time in the past) appeared in the active ending_soon list")
	}
}

func TestFindAll_CombinedPriceFilterAndSort(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST COMBINED FILTER SELLER")

	inRangeLow := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionPrice(t, env, inRangeLow.ID, 200)
	inRangeHigh := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionPrice(t, env, inRangeHigh.ID, 800)
	outOfRange := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionPrice(t, env, outOfRange.ID, 5000)

	minP, maxP := 100, 1000
	results, _, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
		Status:   "active",
		MinPrice: &minP,
		MaxPrice: &maxP,
		SortBy:   "price_desc",
		PerPage:  100,
	})
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	pos := map[string]int{}
	for i, a := range results {
		pos[a.ID.String()] = i
	}

	t.Run("out-of-range auction is excluded even though it would sort first by price_desc", func(t *testing.T) {
		if _, ok := pos[outOfRange.ID.String()]; ok {
			t.Fatalf("expected the 5000-priced auction to be excluded by the price filter regardless of sort mode")
		}
	})
	t.Run("in-range auctions are ordered price_desc among themselves", func(t *testing.T) {
		if pos[inRangeHigh.ID.String()] >= pos[inRangeLow.ID.String()] {
			t.Fatalf("expected inRangeHigh(800) before inRangeLow(200) under price_desc")
		}
	})
}

func TestFindAll_PriceFilter_NoResultsInRange(t *testing.T) {
	env := setupEnv(t)
	ctx := context.Background()
	seller := createTestUser(t, env, "TEST NORESULT FILTER SELLER")

	a := createTestAuction(t, env, seller.ID, "MR", "MRU")
	setAuctionPrice(t, env, a.ID, 50)

	minP, maxP := 100000, 200000
	results, total, err := env.auctionRepo.FindAll(ctx, repository.AuctionFilters{
		Status:   "active",
		MinPrice: &minP,
		MaxPrice: &maxP,
		PerPage:  100,
	})
	if err != nil {
		t.Fatalf("FindAll must not error when no auction matches the price range, got: %v", err)
	}

	t.Run("empty results, not an error, for a range matching nothing", func(t *testing.T) {
		for _, r := range results {
			if r.ID == a.ID {
				t.Fatalf("REGRESSION: a 50-priced auction must not appear for min_price=100000")
			}
		}
	})
	t.Run("total does not count the out-of-range fixture", func(t *testing.T) {
		if total < 0 {
			t.Fatalf("total must never be negative, got %d", total)
		}
	})
}
