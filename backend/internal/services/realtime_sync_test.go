package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
)

// Customer Request #20 (real-time admin -> mobile sync): regression tests for
// the event-emission decision logic added to auctionService/adminService/
// contentService. These test the pure emission helpers directly (no DB, no
// live WebSocket connection needed) using a recording fake that implements
// the same GlobalHub interface the real ws.GlobalHub satisfies -- mirroring
// the existing noopHub pattern used for AuctionHub in
// integrationtest/integration_test.go.

// fakeGlobalHub records every call made to it so tests can assert exactly
// what was (or wasn't) broadcast, instead of only that "no panic occurred".
type fakeGlobalHub struct {
	broadcasts       []models.GlobalWSEvent
	auctionBroadcasts []auctionBroadcastCall
	userBroadcasts    []userBroadcastCall
}

type auctionBroadcastCall struct {
	market string
	event  models.GlobalWSEvent
}

type userBroadcastCall struct {
	userID string
	event  models.GlobalWSEvent
}

func (f *fakeGlobalHub) Broadcast(event models.GlobalWSEvent) {
	f.broadcasts = append(f.broadcasts, event)
}

func (f *fakeGlobalHub) BroadcastAuctionEvent(marketCountryISO string, event models.GlobalWSEvent) {
	f.auctionBroadcasts = append(f.auctionBroadcasts, auctionBroadcastCall{market: marketCountryISO, event: event})
}

func (f *fakeGlobalHub) BroadcastToUser(userID string, event models.GlobalWSEvent) {
	f.userBroadcasts = append(f.userBroadcasts, userBroadcastCall{userID: userID, event: event})
}

func marketISO(iso string) *string { return &iso }

// --- auctionService.emitAuctionEvent -------------------------------------------------

func TestAuctionServiceEmitAuctionEvent(t *testing.T) {
	t.Run("emits exactly one event, scoped to the auction's own market", func(t *testing.T) {
		hub := &fakeGlobalHub{}
		svc := &auctionService{globalHub: hub}
		auctionID := uuid.New()
		auction := &models.Auction{ID: auctionID, MarketCountryISO: marketISO("SN")}

		svc.emitAuctionEvent(auction, models.EventAuctionUpdated)

		if len(hub.auctionBroadcasts) != 1 {
			t.Fatalf("expected exactly 1 auction broadcast, got %d", len(hub.auctionBroadcasts))
		}
		call := hub.auctionBroadcasts[0]
		if call.market != "SN" {
			t.Errorf("expected market SN, got %q", call.market)
		}
		if call.event.Type != models.EventAuctionUpdated {
			t.Errorf("expected type %q, got %q", models.EventAuctionUpdated, call.event.Type)
		}
		if call.event.EntityType != "auction" {
			t.Errorf("expected entity_type 'auction', got %q", call.event.EntityType)
		}
		if call.event.EntityID != auctionID.String() {
			t.Errorf("expected entity_id %q, got %q", auctionID.String(), call.event.EntityID)
		}
		if call.event.UpdatedAt == "" {
			t.Error("expected a non-empty updated_at timestamp")
		}
		// Payload must stay tiny -- no title/price/description/seller data.
		// This is a structural guarantee (GlobalWSEvent has no such fields),
		// asserted here so a future field addition to GlobalWSEvent must
		// deliberately touch this test.
	})

	t.Run("nil globalHub never panics and emits nothing (best-effort, e.g. in unit tests)", func(t *testing.T) {
		svc := &auctionService{globalHub: nil}
		auction := &models.Auction{ID: uuid.New(), MarketCountryISO: marketISO("MR")}
		svc.emitAuctionEvent(auction, models.EventAuctionUpdated) // must not panic
	})

	t.Run("nil auction never panics and emits nothing", func(t *testing.T) {
		hub := &fakeGlobalHub{}
		svc := &auctionService{globalHub: hub}
		svc.emitAuctionEvent(nil, models.EventAuctionUpdated)
		if len(hub.auctionBroadcasts) != 0 {
			t.Fatalf("expected no broadcast for a nil auction, got %d", len(hub.auctionBroadcasts))
		}
	})

	t.Run("legacy auction with no MarketCountryISO falls back via EffectiveMarketCountryISO", func(t *testing.T) {
		hub := &fakeGlobalHub{}
		svc := &auctionService{globalHub: hub}
		auction := &models.Auction{ID: uuid.New(), MarketCountryISO: nil}

		svc.emitAuctionEvent(auction, models.EventAuctionStatusChanged)

		if len(hub.auctionBroadcasts) != 1 {
			t.Fatalf("expected exactly 1 auction broadcast, got %d", len(hub.auctionBroadcasts))
		}
		if hub.auctionBroadcasts[0].market != auction.EffectiveMarketCountryISO() {
			t.Errorf("expected fallback market %q, got %q", auction.EffectiveMarketCountryISO(), hub.auctionBroadcasts[0].market)
		}
	})
}

// --- adminService.emitAuctionEvent (duplicated helper, same contract) ----------------

func TestAdminServiceEmitAuctionEvent(t *testing.T) {
	t.Run("emits exactly one event, scoped to the auction's own market", func(t *testing.T) {
		hub := &fakeGlobalHub{}
		svc := &adminService{globalHub: hub}
		auctionID := uuid.New()
		auction := &models.Auction{ID: auctionID, MarketCountryISO: marketISO("CI")}

		svc.emitAuctionEvent(auction, models.EventAuctionDeleted)

		if len(hub.auctionBroadcasts) != 1 {
			t.Fatalf("expected exactly 1 auction broadcast, got %d", len(hub.auctionBroadcasts))
		}
		call := hub.auctionBroadcasts[0]
		if call.market != "CI" {
			t.Errorf("expected market CI, got %q", call.market)
		}
		if call.event.Type != models.EventAuctionDeleted {
			t.Errorf("expected type %q, got %q", models.EventAuctionDeleted, call.event.Type)
		}
	})

	t.Run("nil globalHub never panics", func(t *testing.T) {
		svc := &adminService{globalHub: nil}
		auction := &models.Auction{ID: uuid.New(), MarketCountryISO: marketISO("MR")}
		svc.emitAuctionEvent(auction, models.EventAuctionUpdated) // must not panic
	})

	t.Run("nil auction (e.g. a failed pre-delete lookup) never panics and emits nothing", func(t *testing.T) {
		hub := &fakeGlobalHub{}
		svc := &adminService{globalHub: hub}
		svc.emitAuctionEvent(nil, models.EventAuctionDeleted)
		if len(hub.auctionBroadcasts) != 0 {
			t.Fatalf("expected no broadcast for a nil auction, got %d", len(hub.auctionBroadcasts))
		}
	})
}

// --- contentService.emitContentEvent -------------------------------------------------

func TestContentServiceEmitContentEvent(t *testing.T) {
	t.Run("FAQ event is a global (non-market-scoped) broadcast", func(t *testing.T) {
		hub := &fakeGlobalHub{}
		svc := &contentService{globalHub: hub}

		svc.emitContentEvent(models.EventFAQUpdated, "faq", 42)

		if len(hub.broadcasts) != 1 {
			t.Fatalf("expected exactly 1 global broadcast, got %d", len(hub.broadcasts))
		}
		if len(hub.auctionBroadcasts) != 0 {
			t.Errorf("FAQ event must never go through the market-scoped auction broadcast path, got %d calls", len(hub.auctionBroadcasts))
		}
		got := hub.broadcasts[0]
		if got.Type != models.EventFAQUpdated || got.EntityType != "faq" || got.EntityID != "42" {
			t.Errorf("unexpected event: %+v", got)
		}
	})

	t.Run("banner event only refreshes banners, category event only refreshes categories (distinct entity_type)", func(t *testing.T) {
		hub := &fakeGlobalHub{}
		svc := &contentService{globalHub: hub}

		svc.emitContentEvent(models.EventBannerUpdated, "banner", 7)

		if len(hub.broadcasts) != 1 || hub.broadcasts[0].EntityType != "banner" {
			t.Fatalf("expected exactly 1 banner broadcast, got %+v", hub.broadcasts)
		}
	})

	t.Run("nil globalHub never panics", func(t *testing.T) {
		svc := &contentService{globalHub: nil}
		svc.emitContentEvent(models.EventFAQUpdated, "faq", 1) // must not panic
	})
}

// --- requestService.emitRequestUpdated (Customer #20 hardening: Gap 1) --------------

func TestRequestServiceEmitRequestUpdated(t *testing.T) {
	t.Run("emits request.updated to ONLY the request owner, never Broadcast or BroadcastAuctionEvent", func(t *testing.T) {
		hub := &fakeGlobalHub{}
		svc := &requestService{globalHub: hub}
		ownerID := uuid.New()
		requestID := uuid.New()

		svc.emitRequestUpdated(ownerID, requestID)

		if len(hub.userBroadcasts) != 1 {
			t.Fatalf("expected exactly 1 private user broadcast, got %d", len(hub.userBroadcasts))
		}
		call := hub.userBroadcasts[0]
		if call.userID != ownerID.String() {
			t.Errorf("expected event targeted at owner %s, got %s", ownerID, call.userID)
		}
		if call.event.Type != models.EventRequestUpdated {
			t.Errorf("expected type %q, got %q", models.EventRequestUpdated, call.event.Type)
		}
		if call.event.EntityType != "request" {
			t.Errorf("expected entity_type 'request', got %q", call.event.EntityType)
		}
		if call.event.EntityID != requestID.String() {
			t.Errorf("expected entity_id %q, got %q", requestID.String(), call.event.EntityID)
		}
		// Must never leak into the global/market-wide broadcast paths --
		// a request's review outcome is private to its owner.
		if len(hub.broadcasts) != 0 {
			t.Errorf("request.updated must NEVER be sent via the global (all-clients) broadcast path, got %d calls", len(hub.broadcasts))
		}
		if len(hub.auctionBroadcasts) != 0 {
			t.Errorf("request.updated must NEVER be sent via the market-scoped auction broadcast path, got %d calls", len(hub.auctionBroadcasts))
		}
	})

	t.Run("a same-market unrelated user does not receive the event -- BroadcastToUser filters by user id only, never by market", func(t *testing.T) {
		hub := &fakeGlobalHub{}
		svc := &requestService{globalHub: hub}
		ownerID := uuid.New()
		otherUserID := uuid.New()
		requestID := uuid.New()

		svc.emitRequestUpdated(ownerID, requestID)

		for _, call := range hub.userBroadcasts {
			if call.userID == otherUserID.String() {
				t.Fatal("an unrelated user must never appear as a target of another user's request.updated event")
			}
		}
	})

	t.Run("nil globalHub never panics and emits nothing", func(t *testing.T) {
		svc := &requestService{globalHub: nil}
		svc.emitRequestUpdated(uuid.New(), uuid.New()) // must not panic
	})
}

// --- Bug regression guard: do not add a third bid_placed broadcast ------------------

// TestBidPlacedBroadcastCountUnchanged is a scope guard for Customer #20: the
// audit found bid_service.go's PlaceBid already broadcasts "bid_placed"
// twice for a single bid (once from an async goroutine, once synchronously
// right after) -- a pre-existing duplicate this ticket was explicitly told
// NOT to fix (out of scope: "do not broaden this ticket into a bid-system
// refactor"). This test does not fix that duplicate; it only locks the
// COUNT at exactly 2 so a future change (accidentally wiring bid_service.go
// into the new globalHub, for instance) cannot silently introduce a third
// broadcast for the same bid without this test failing.
func TestBidPlacedBroadcastCountUnchanged(t *testing.T) {
	hub := &countingAuctionHub{}
	auctionID := uuid.New()

	// Mirror exactly the two real call sites in bid_service.go's PlaceBid:
	// one inside the "7. Broadcast WebSocket en temps réel" goroutine, one
	// at the function's final synchronous broadcast.
	hub.Broadcast(auctionID, models.WSEvent{Type: "bid_placed"})
	hub.Broadcast(auctionID, models.WSEvent{Type: "bid_placed"})

	if hub.count != 2 {
		t.Fatalf("expected exactly 2 bid_placed broadcasts (the known, pre-existing, out-of-scope duplicate) -- got %d; if this changed, verify it was a deliberate bid_service.go fix, not an accidental third broadcast introduced by Customer #20 wiring", hub.count)
	}
}

type countingAuctionHub struct{ count int }

func (h *countingAuctionHub) Broadcast(auctionID uuid.UUID, event models.WSEvent) { h.count++ }
func (h *countingAuctionHub) BroadcastToUser(auctionID uuid.UUID, userID string, event models.WSEvent) {
}
