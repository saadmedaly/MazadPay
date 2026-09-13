package models

// GlobalWSEvent -- Customer Request #20's mobile-wide invalidation event,
// sent over the authenticated global channel (GET /ws/global). Intentionally
// tiny: only enough metadata for a connected client to decide WHICH provider
// to refetch via REST, never the entity's own data. REST stays the source of
// truth (see GlobalHub doc comment) -- sending complete entities here would
// duplicate the REST contract and risk leaking fields the requesting user
// isn't authorized to see.
type GlobalWSEvent struct {
	Type       string `json:"type"`        // e.g. auction.updated, faq.updated, banner.updated, category.updated
	EntityType string `json:"entity_type"` // e.g. "auction", "faq", "banner", "category"
	EntityID   string `json:"entity_id"`
	UpdatedAt  string `json:"updated_at"` // RFC3339; informational only, clients must still refetch rather than trust this as fresh data
}

// Global WS event type constants -- used both when emitting (services) and
// when the mobile client's message switch decides which provider to
// invalidate, so the literal strings live in exactly one place.
const (
	EventAuctionCreated       = "auction.created"
	EventAuctionUpdated       = "auction.updated"
	EventAuctionStatusChanged = "auction.status_changed"
	EventAuctionDeleted       = "auction.deleted"

	EventFAQUpdated     = "faq.updated"
	EventBannerUpdated  = "banner.updated"
	EventCategoryUpdated = "category.updated"

	// EventRequestUpdated (Customer #20 hardening round): a private, owner-only
	// event -- never broadcast globally or market-wide (see GlobalHub.BroadcastToUser).
	// Signals "your auction/banner request was reviewed" to an open Requests
	// page; the actual review outcome/notes are fetched via REST, never carried
	// in this event's payload.
	EventRequestUpdated = "request.updated"
)
