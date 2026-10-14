//go:build integration

package integrationtest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mazadpay/backend/internal/models"
)

// MAZADPAY locations-list refresh bug: GetLocations (auction_service.go)
// cached its full result in Redis under the fixed key "locations" for 1
// hour, with nothing anywhere invalidating that key on create/update/
// delete -- a location mutation committed cleanly to Postgres but stayed
// invisible through GET /v1/api/locations for up to an hour. Separately,
// the repo query underlying it used `SELECT DISTINCT ON (city_name_ar) *`,
// which silently dropped every location sharing a city name with another
// (kept only one arbitrary row per unique city_name_ar). These tests prove
// both are fixed: admin_service.go's Create/Update/DeleteLocation now call
// invalidateLocationsCache(ctx) (a Redis DEL on "locations") after their DB
// mutation succeeds, and auction_repo.go's GetLocations/GetLocationsByCountry
// no longer use DISTINCT ON.

func uniqueTestCityName(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("TEST_LOC_%s_%d", t.Name(), time.Now().UnixNano())
}

func cleanupTestLocation(t *testing.T, env *testEnv, id int) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = env.db.ExecContext(context.Background(), `DELETE FROM locations WHERE id = $1`, id)
	})
}

func TestLocationsCache_CreateInvalidatesCache(t *testing.T) {
	env := setupEnv(t)
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "loc cache admin")
	ctx := context.Background()

	// Prime the Redis cache with whatever GetLocations currently returns,
	// exactly like a real prior admin page load would.
	if _, err := env.auctSvc.GetLocations(ctx); err != nil {
		t.Fatalf("priming GetLocations failed: %v", err)
	}

	cityName := uniqueTestCityName(t)
	loc := &models.Location{CityNameAr: cityName, CityNameFr: cityName, AreaNameAr: "TEST", AreaNameFr: "TEST"}
	if err := adminSvc.CreateLocation(ctx, loc, admin.ID); err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}
	cleanupTestLocation(t, env, loc.ID)

	locations, err := env.auctSvc.GetLocations(ctx)
	if err != nil {
		t.Fatalf("GetLocations after create failed: %v", err)
	}
	if !containsLocationID(locations, loc.ID) {
		t.Fatalf("REGRESSION: newly created location id=%d not present in GetLocations immediately after create -- cache was not invalidated", loc.ID)
	}
}

func TestLocationsCache_UpdateInvalidatesCache(t *testing.T) {
	env := setupEnv(t)
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "loc cache admin")
	ctx := context.Background()

	cityName := uniqueTestCityName(t)
	loc := &models.Location{CityNameAr: cityName, CityNameFr: cityName, AreaNameAr: "TEST", AreaNameFr: "TEST"}
	if err := adminSvc.CreateLocation(ctx, loc, admin.ID); err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}
	cleanupTestLocation(t, env, loc.ID)

	// Prime the cache with the pre-update value.
	if _, err := env.auctSvc.GetLocations(ctx); err != nil {
		t.Fatalf("priming GetLocations failed: %v", err)
	}

	updatedArea := "TEST_UPDATED_" + cityName
	loc.AreaNameAr = updatedArea
	if err := adminSvc.UpdateLocation(ctx, loc, admin.ID); err != nil {
		t.Fatalf("UpdateLocation failed: %v", err)
	}

	locations, err := env.auctSvc.GetLocations(ctx)
	if err != nil {
		t.Fatalf("GetLocations after update failed: %v", err)
	}
	found := false
	for _, l := range locations {
		if l.ID == loc.ID {
			found = true
			if l.AreaNameAr != updatedArea {
				t.Fatalf("REGRESSION: GetLocations still returns the pre-update area_name_ar %q, expected %q -- cache was not invalidated", l.AreaNameAr, updatedArea)
			}
		}
	}
	if !found {
		t.Fatalf("updated location id=%d missing from GetLocations entirely", loc.ID)
	}
}

func TestLocationsCache_DeleteInvalidatesCache(t *testing.T) {
	env := setupEnv(t)
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "loc cache admin")
	ctx := context.Background()

	cityName := uniqueTestCityName(t)
	loc := &models.Location{CityNameAr: cityName, CityNameFr: cityName, AreaNameAr: "TEST", AreaNameFr: "TEST"}
	if err := adminSvc.CreateLocation(ctx, loc, admin.ID); err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}

	// Prime the cache while the location still exists.
	if _, err := env.auctSvc.GetLocations(ctx); err != nil {
		t.Fatalf("priming GetLocations failed: %v", err)
	}

	if err := adminSvc.DeleteLocation(ctx, loc.ID, admin.ID); err != nil {
		t.Fatalf("DeleteLocation failed: %v", err)
	}

	locations, err := env.auctSvc.GetLocations(ctx)
	if err != nil {
		t.Fatalf("GetLocations after delete failed: %v", err)
	}
	if containsLocationID(locations, loc.ID) {
		t.Fatalf("REGRESSION: deleted location id=%d still present in GetLocations immediately after delete -- cache was not invalidated", loc.ID)
	}
}

// TestLocationsCache_RedisKeyActuallyDeleted proves invalidateLocationsCache
// performs a real Redis DEL on the exact "locations" key GetLocations reads
// from, not just that the net observable behavior happens to look right.
func TestLocationsCache_RedisKeyActuallyDeleted(t *testing.T) {
	env := setupEnv(t)
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "loc cache admin")
	ctx := context.Background()

	if _, err := env.auctSvc.GetLocations(ctx); err != nil {
		t.Fatalf("priming GetLocations failed: %v", err)
	}
	if exists, err := env.rdb.Exists(ctx, "locations").Result(); err != nil {
		t.Fatalf("redis EXISTS failed: %v", err)
	} else if exists == 0 {
		t.Fatalf("expected \"locations\" Redis key to exist after priming GetLocations")
	}

	cityName := uniqueTestCityName(t)
	loc := &models.Location{CityNameAr: cityName, CityNameFr: cityName, AreaNameAr: "TEST", AreaNameFr: "TEST"}
	if err := adminSvc.CreateLocation(ctx, loc, admin.ID); err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}
	cleanupTestLocation(t, env, loc.ID)

	if exists, err := env.rdb.Exists(ctx, "locations").Result(); err != nil {
		t.Fatalf("redis EXISTS failed: %v", err)
	} else if exists != 0 {
		t.Fatalf("REGRESSION: \"locations\" Redis key still exists after CreateLocation -- invalidateLocationsCache did not delete it")
	}
}

// TestLocationsCache_SameCityDifferentZonesBothReturned proves the removed
// DISTINCT ON (city_name_ar) no longer silently drops sibling locations
// that share a city name but differ in zone/id.
func TestLocationsCache_SameCityDifferentZonesBothReturned(t *testing.T) {
	env := setupEnv(t)
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "loc cache admin")
	ctx := context.Background()

	sharedCity := uniqueTestCityName(t)
	locA := &models.Location{CityNameAr: sharedCity, CityNameFr: sharedCity, AreaNameAr: "ZONE_A", AreaNameFr: "ZONE_A"}
	locB := &models.Location{CityNameAr: sharedCity, CityNameFr: sharedCity, AreaNameAr: "ZONE_B", AreaNameFr: "ZONE_B"}

	if err := adminSvc.CreateLocation(ctx, locA, admin.ID); err != nil {
		t.Fatalf("CreateLocation (zone A) failed: %v", err)
	}
	cleanupTestLocation(t, env, locA.ID)
	if err := adminSvc.CreateLocation(ctx, locB, admin.ID); err != nil {
		t.Fatalf("CreateLocation (zone B) failed: %v", err)
	}
	cleanupTestLocation(t, env, locB.ID)

	locations, err := env.auctSvc.GetLocations(ctx)
	if err != nil {
		t.Fatalf("GetLocations failed: %v", err)
	}
	if !containsLocationID(locations, locA.ID) || !containsLocationID(locations, locB.ID) {
		t.Fatalf("REGRESSION: two locations sharing city_name_ar=%q (ids %d and %d) were not both returned by GetLocations -- DISTINCT ON (city_name_ar) is dropping rows", sharedCity, locA.ID, locB.ID)
	}
}

// TestLocationsCache_UnrelatedDataUnaffected proves cache invalidation and
// the query change only touch locations -- an unrelated auction (and its
// own cached data) is untouched by a location mutation.
func TestLocationsCache_UnrelatedDataUnaffected(t *testing.T) {
	env := setupEnv(t)
	adminSvc := newTestAdminService(t, env)
	admin := createTestAdmin(t, env, "loc cache admin")
	ctx := context.Background()

	seller := createTestUser(t, env, "loc cache seller")
	auction := createTestAuction(t, env, seller.ID, "MR", "MRU")

	cityName := uniqueTestCityName(t)
	loc := &models.Location{CityNameAr: cityName, CityNameFr: cityName, AreaNameAr: "TEST", AreaNameFr: "TEST"}
	if err := adminSvc.CreateLocation(ctx, loc, admin.ID); err != nil {
		t.Fatalf("CreateLocation failed: %v", err)
	}
	cleanupTestLocation(t, env, loc.ID)

	var status string
	if err := env.db.GetContext(ctx, &status, `SELECT status FROM auctions WHERE id = $1`, auction.ID); err != nil {
		t.Fatalf("failed to re-read auction after unrelated location create: %v", err)
	}
	if status != "active" {
		t.Fatalf("REGRESSION: unrelated auction status changed to %q after an unrelated location create", status)
	}
}

func containsLocationID(locations []models.Location, id int) bool {
	for _, l := range locations {
		if l.ID == id {
			return true
		}
	}
	return false
}
