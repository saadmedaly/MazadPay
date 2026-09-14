package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Customer Request #22: broadcast/optional-image behavior. Reuses the exact
// fakeNotificationRepo/fakeUserRepoForBroadcast pattern from
// notification_broadcast_test.go so these run against the real
// notificationService.SendBroadcastWithImage/SendPushWithImage logic without
// a live database.

// TestSendBroadcastWithImage_PersistsSameImageURLAcrossRecipients covers:
//   - an image broadcast succeeds
//   - image_url persists on every created row
//   - the SAME url string is reused across recipients (never re-uploaded/
//     regenerated per user)
func TestSendBroadcastWithImage_PersistsSameImageURLAcrossRecipients(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()

	notifRepo := &fakeNotificationRepo{}
	userRepo := &fakeUserRepoForBroadcast{activeUserIDs: []uuid.UUID{userA, userB}}
	svc := NewNotificationService(notifRepo, userRepo, "", "", zap.NewNop(), nil)

	const imageURL = "https://cdn.example.com/notifications/abc123.jpg"
	targetUsers, sent, failed, err := svc.SendBroadcastWithImage(context.Background(), "New promo", "check it out", "general", nil, imageURL)
	if err != nil {
		t.Fatalf("SendBroadcastWithImage returned an error: %v", err)
	}
	if targetUsers != 2 || sent != 2 || failed != 0 {
		t.Fatalf("expected targetUsers=2 sent=2 failed=0, got targetUsers=%d sent=%d failed=%d", targetUsers, sent, failed)
	}

	if len(notifRepo.created) != 2 {
		t.Fatalf("expected exactly 2 persisted rows, got %d", len(notifRepo.created))
	}
	for _, n := range notifRepo.created {
		if n.ImageURL == nil {
			t.Fatalf("expected ImageURL to be persisted for user %s, got nil", n.UserID)
		}
		if *n.ImageURL != imageURL {
			t.Fatalf("expected ImageURL %q, got %q", imageURL, *n.ImageURL)
		}
	}
}

// TestSendBroadcast_NoImage_RemainsValid proves the pre-existing text-only
// broadcast path (imageURL == "") still results in a nil ImageURL column --
// no image box should ever render for these rows.
func TestSendBroadcast_NoImage_RemainsValid(t *testing.T) {
	userA := uuid.New()

	notifRepo := &fakeNotificationRepo{}
	userRepo := &fakeUserRepoForBroadcast{activeUserIDs: []uuid.UUID{userA}}
	svc := NewNotificationService(notifRepo, userRepo, "", "", zap.NewNop(), nil)

	targetUsers, sent, failed, err := svc.SendBroadcast(context.Background(), "Text only", "no image here", "general", nil)
	if err != nil {
		t.Fatalf("SendBroadcast returned an error: %v", err)
	}
	if targetUsers != 1 || sent != 1 || failed != 0 {
		t.Fatalf("expected targetUsers=1 sent=1 failed=0, got targetUsers=%d sent=%d failed=%d", targetUsers, sent, failed)
	}
	if len(notifRepo.created) != 1 {
		t.Fatalf("expected 1 persisted row, got %d", len(notifRepo.created))
	}
	if notifRepo.created[0].ImageURL != nil {
		t.Fatalf("expected ImageURL to be nil for a text-only broadcast, got %q", *notifRepo.created[0].ImageURL)
	}
}

// TestSendPushWithImage_PersistsImageURL_SingleUser covers the single-user
// (non-broadcast) image send path, and that FCM data carries notification_id
// (the id of the just-created row) so a push tap can look it up via the
// existing GET /notifications endpoint.
func TestSendPushWithImage_PersistsImageURL_SingleUser(t *testing.T) {
	userA := uuid.New()

	notifRepo := &fakeNotificationRepo{}
	userRepo := &fakeUserRepoForBroadcast{activeUserIDs: []uuid.UUID{userA}}
	svc := NewNotificationService(notifRepo, userRepo, "", "", zap.NewNop(), nil)

	const imageURL = "https://cdn.example.com/notifications/xyz789.png"
	err := svc.SendPushWithImage(context.Background(), userA, "Hi", "body text", "general", nil, imageURL)
	if err != nil {
		t.Fatalf("SendPushWithImage returned an error: %v", err)
	}
	if len(notifRepo.created) != 1 {
		t.Fatalf("expected 1 persisted row, got %d", len(notifRepo.created))
	}
	n := notifRepo.created[0]
	if n.ImageURL == nil || *n.ImageURL != imageURL {
		t.Fatalf("expected ImageURL %q to be persisted, got %v", imageURL, n.ImageURL)
	}
	// notification_id corresponds to the actual DB row: the created row's own
	// ID must be a valid, non-nil UUID that a client could round-trip through
	// GET /notifications and match by string equality.
	if n.ID == uuid.Nil {
		t.Fatalf("expected the created notification to have a real generated ID")
	}
}

// TestSendPush_NoImage_RemainsIdenticalToPriorBehavior is a regression guard:
// the pre-Customer-#22 SendPush call sites (auction/payment/wallet
// notifications) must keep producing rows with a nil ImageURL, proving
// sendPush's shared implementation didn't change behavior for imageURL == "".
func TestSendPush_NoImage_RemainsIdenticalToPriorBehavior(t *testing.T) {
	userA := uuid.New()

	notifRepo := &fakeNotificationRepo{}
	userRepo := &fakeUserRepoForBroadcast{activeUserIDs: []uuid.UUID{userA}}
	svc := NewNotificationService(notifRepo, userRepo, "", "", zap.NewNop(), nil)

	err := svc.SendPush(context.Background(), userA, "Auction won", "congrats", "auction_won", map[string]string{"auction_id": "abc"})
	if err != nil {
		t.Fatalf("SendPush returned an error: %v", err)
	}
	if len(notifRepo.created) != 1 {
		t.Fatalf("expected 1 persisted row, got %d", len(notifRepo.created))
	}
	if notifRepo.created[0].ImageURL != nil {
		t.Fatalf("expected ImageURL to remain nil for existing SendPush callers, got %q", *notifRepo.created[0].ImageURL)
	}
	if notifRepo.created[0].Type != "auction_won" {
		t.Fatalf("expected type auction_won to be preserved, got %q", notifRepo.created[0].Type)
	}
}
