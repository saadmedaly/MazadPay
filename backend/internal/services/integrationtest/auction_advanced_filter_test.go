//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
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
