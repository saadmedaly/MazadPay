package ws

import (
	"encoding/json"
	"testing"

	"github.com/mazadpay/backend/internal/models"
	"go.uber.org/zap"
)

// Customer Request #20: market isolation and private-targeting regression
// tests for GlobalHub. Client.send is a buffered channel (size 64, see
// client.go), so these tests read from it directly rather than needing a
// live network connection -- NewClient(nil, ...) is safe here because these
// tests never call Unregister (which would call the nil conn's Close()).

func testClient(userID, market string) *Client {
	c := NewClient(nil, userID, zap.NewNop())
	c.SetMarket(market)
	return c
}

func drain(t *testing.T, c *Client) *models.GlobalWSEvent {
	t.Helper()
	select {
	case payload := <-c.send:
		var event models.GlobalWSEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			t.Fatalf("failed to unmarshal event: %v", err)
		}
		return &event
	default:
		return nil
	}
}

func assertEmpty(t *testing.T, c *Client, label string) {
	t.Helper()
	select {
	case payload := <-c.send:
		t.Fatalf("%s: expected no message, got %s", label, string(payload))
	default:
	}
}

func TestGlobalHub_BroadcastAuctionEvent_MarketIsolation(t *testing.T) {
	hub := NewGlobalHub(zap.NewNop())

	snClient := testClient("user-sn", "SN")
	ciClient := testClient("user-ci", "CI")
	noMarketClient := testClient("user-none", "")

	hub.Register(snClient)
	hub.Register(ciClient)
	hub.Register(noMarketClient)

	hub.BroadcastAuctionEvent("SN", models.GlobalWSEvent{
		Type:       models.EventAuctionUpdated,
		EntityType: "auction",
		EntityID:   "auction-1",
	})

	got := drain(t, snClient)
	if got == nil {
		t.Fatal("SN client (matching market) must receive the auction event")
	}
	if got.EntityID != "auction-1" {
		t.Errorf("unexpected entity_id: %q", got.EntityID)
	}

	assertEmpty(t, ciClient, "CI client (different market) must NOT receive an SN-scoped auction event -- this is the exact privacy/market-isolation guarantee Customer #20 requires")
	assertEmpty(t, noMarketClient, "a client with no market stamped must never match any market broadcast (must not be treated as a wildcard)")
}

func TestGlobalHub_Broadcast_ReachesEveryClientRegardlessOfMarket(t *testing.T) {
	hub := NewGlobalHub(zap.NewNop())
	snClient := testClient("user-sn", "SN")
	ciClient := testClient("user-ci", "CI")
	hub.Register(snClient)
	hub.Register(ciClient)

	hub.Broadcast(models.GlobalWSEvent{Type: models.EventFAQUpdated, EntityType: "faq", EntityID: "1"})

	if drain(t, snClient) == nil {
		t.Error("Broadcast (FAQ/banner/category -- non-market content) must reach the SN client")
	}
	if drain(t, ciClient) == nil {
		t.Error("Broadcast (FAQ/banner/category -- non-market content) must reach the CI client")
	}
}

func TestGlobalHub_BroadcastToUser_OnlyTargetsIntendedUser(t *testing.T) {
	hub := NewGlobalHub(zap.NewNop())
	owner := testClient("owner-1", "SN")
	otherUserSameMarket := testClient("other-2", "SN")
	hub.Register(owner)
	hub.Register(otherUserSameMarket)

	hub.BroadcastToUser("owner-1", models.GlobalWSEvent{Type: "request.reviewed", EntityType: "auction_request", EntityID: "req-1"})

	if drain(t, owner) == nil {
		t.Error("the targeted owner must receive their own private event")
	}
	assertEmpty(t, otherUserSameMarket, "a different user in the SAME market must NOT receive another user's private/owner-targeted event, even though BroadcastAuctionEvent alone would let market match")
}

func TestGlobalHub_ConnectedCount(t *testing.T) {
	hub := NewGlobalHub(zap.NewNop())
	if hub.ConnectedCount() != 0 {
		t.Fatalf("expected 0 connected clients initially, got %d", hub.ConnectedCount())
	}
	hub.Register(testClient("u1", "SN"))
	hub.Register(testClient("u2", "CI"))
	if hub.ConnectedCount() != 2 {
		t.Fatalf("expected 2 connected clients, got %d", hub.ConnectedCount())
	}
}
