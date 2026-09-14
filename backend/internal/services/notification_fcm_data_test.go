package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Customer Request #22 hardening round: proves the FCM notification_id
// contract end to end. buildFCMData itself was extracted from sendPush
// specifically so this logic is unit-testable without a live
// *messaging.Client (the real FCM send path is architecturally untestable
// here -- s.fcm is a concrete *firebase.google.com/go/v4/messaging.Client
// with no interface seam, and every existing test passes fcm=nil, which
// short-circuits sendPush before fcmData is ever built). The per-recipient
// distinctness and image-sharing guarantees are proven via the real
// notificationService.SendBroadcastWithImage using the fake in-memory repo,
// exactly matching how each recipient's real notification.ID would flow
// into buildFCMData in production (see sendPush: `buildFCMData(data,
// notification.ID)`, called with THAT recipient's own freshly-created row).

// TestBuildFCMData_NotificationIDPresent_NonEmpty proves 1 and 2.
func TestBuildFCMData_NotificationIDPresent_NonEmpty(t *testing.T) {
	id := uuid.New()
	fcmData := buildFCMData(nil, id)

	value, ok := fcmData["notification_id"]
	if !ok {
		t.Fatalf("expected notification_id key to be present in FCM data")
	}
	if value == "" {
		t.Fatalf("expected notification_id to be non-empty")
	}
	if value != id.String() {
		t.Fatalf("expected notification_id %q, got %q", id.String(), value)
	}
}

// TestBuildFCMData_PreservesCallerData proves existing SendPush callers
// (which pass their own data map, e.g. {"auction_id": "..."}) remain
// compatible: their keys survive alongside the new notification_id key.
func TestBuildFCMData_PreservesCallerData(t *testing.T) {
	callerData := map[string]string{"auction_id": "abc123", "custom": "value"}
	id := uuid.New()

	fcmData := buildFCMData(callerData, id)

	if fcmData["auction_id"] != "abc123" {
		t.Fatalf("expected caller's auction_id to be preserved, got %q", fcmData["auction_id"])
	}
	if fcmData["custom"] != "value" {
		t.Fatalf("expected caller's custom field to be preserved, got %q", fcmData["custom"])
	}
	if fcmData["notification_id"] != id.String() {
		t.Fatalf("expected notification_id to be added, got %q", fcmData["notification_id"])
	}
	// The caller's original map must never be mutated (SendBroadcast reuses
	// the same map across every recipient in its loop).
	if _, mutated := callerData["notification_id"]; mutated {
		t.Fatalf("buildFCMData must not mutate the caller's original data map")
	}
}

// TestBuildFCMData_NilData_StillProducesNotificationID proves a text-only /
// no-extra-data notification (item 6: "text-only notifications still have
// notification_id") still gets the field.
func TestBuildFCMData_NilData_StillProducesNotificationID(t *testing.T) {
	id := uuid.New()
	fcmData := buildFCMData(nil, id)

	if len(fcmData) != 1 {
		t.Fatalf("expected exactly 1 key (notification_id) for nil caller data, got %d: %v", len(fcmData), fcmData)
	}
	if fcmData["notification_id"] != id.String() {
		t.Fatalf("expected notification_id to be set even with nil data")
	}
}

// TestBuildFCMData_DistinctIDsProduceDistinctPayloads is the direct proof
// that two different recipients (two different notification.ID values, as
// sendPush would pass for two different broadcast recipients) produce two
// distinct fcmData maps with distinct notification_id values -- never one
// shared ID for a whole broadcast.
func TestBuildFCMData_DistinctIDsProduceDistinctPayloads(t *testing.T) {
	idA := uuid.New()
	idB := uuid.New()

	fcmDataA := buildFCMData(nil, idA)
	fcmDataB := buildFCMData(nil, idB)

	if fcmDataA["notification_id"] == fcmDataB["notification_id"] {
		t.Fatalf("expected distinct notification_id values for distinct recipients, both were %q", fcmDataA["notification_id"])
	}
	if fcmDataA["notification_id"] != idA.String() || fcmDataB["notification_id"] != idB.String() {
		t.Fatalf("expected each fcmData to carry its OWN recipient's row ID")
	}
}

// TestSendBroadcastWithImage_PerRecipientDistinctRowIDs_SharedImageURL proves
// the full real-service contract (items 3, 4, 5 of the hardening brief):
//   - each broadcast recipient gets their OWN distinct DB row ID (proving
//     what would flow into buildFCMData as notification.ID per sendPush call)
//   - the image URL is nonetheless IDENTICAL (shared) across every recipient
//
// This directly exercises notificationService.SendBroadcastWithImage ->
// sendBroadcast -> sendPush (once per recipient), the exact call chain that
// in production also calls buildFCMData(data, notification.ID) per
// recipient -- proving the DB-row side of the "distinct IDs, shared image"
// contract for real, with buildFCMData's own unit tests above covering the
// FCM-payload-construction side that has no live-FCM-client seam to test end
// to end.
func TestSendBroadcastWithImage_PerRecipientDistinctRowIDs_SharedImageURL(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()

	notifRepo := &fakeNotificationRepo{}
	userRepo := &fakeUserRepoForBroadcast{activeUserIDs: []uuid.UUID{userA, userB}}
	svc := NewNotificationService(notifRepo, userRepo, "", "", zap.NewNop(), nil)

	const imageURL = "https://cdn.example.com/notifications/shared.jpg"
	_, sent, failed, err := svc.SendBroadcastWithImage(context.Background(), "Broadcast", "body", "general", nil, imageURL)
	if err != nil {
		t.Fatalf("SendBroadcastWithImage returned an error: %v", err)
	}
	if sent != 2 || failed != 0 {
		t.Fatalf("expected sent=2 failed=0, got sent=%d failed=%d", sent, failed)
	}
	if len(notifRepo.created) != 2 {
		t.Fatalf("expected 2 persisted rows, got %d", len(notifRepo.created))
	}

	rowA, rowB := notifRepo.created[0], notifRepo.created[1]

	if rowA.ID == uuid.Nil || rowB.ID == uuid.Nil {
		t.Fatalf("expected both rows to have real generated IDs")
	}
	if rowA.ID == rowB.ID {
		t.Fatalf("SECURITY/CORRECTNESS REGRESSION: two different recipients got the SAME notification row ID (%s) -- this would mean a shared notification_id across a broadcast, not per-recipient", rowA.ID)
	}

	if rowA.ImageURL == nil || rowB.ImageURL == nil {
		t.Fatalf("expected ImageURL to be persisted for both recipients")
	}
	if *rowA.ImageURL != imageURL || *rowB.ImageURL != imageURL {
		t.Fatalf("expected the SAME image URL shared across recipients, got %q and %q", *rowA.ImageURL, *rowB.ImageURL)
	}

	// The exact contract: distinct notification_id per recipient, shared
	// image_url across all of them.
	fcmDataA := buildFCMData(nil, rowA.ID)
	fcmDataB := buildFCMData(nil, rowB.ID)
	if fcmDataA["notification_id"] == fcmDataB["notification_id"] {
		t.Fatalf("expected distinct notification_id per recipient")
	}
}

// TestBuildFCMData_ExistingSendPushCallers_Compatible is a regression guard:
// the auction/payment/wallet call sites that predate Customer #22 pass their
// own data map (e.g. {"auction_id": "..."}) and never knew about
// notification_id -- proves that shape still works exactly as before, with
// notification_id simply appended, never replacing/renaming any existing
// key.
func TestBuildFCMData_ExistingSendPushCallers_Compatible(t *testing.T) {
	auctionData := map[string]string{"auction_id": "auc-1", "type": "auction_won"}
	id := uuid.New()

	fcmData := buildFCMData(auctionData, id)

	if len(fcmData) != 3 {
		t.Fatalf("expected 3 keys (auction_id, type, notification_id), got %d: %v", len(fcmData), fcmData)
	}
	if fcmData["auction_id"] != "auc-1" || fcmData["type"] != "auction_won" {
		t.Fatalf("expected existing caller keys unchanged, got %v", fcmData)
	}
}
