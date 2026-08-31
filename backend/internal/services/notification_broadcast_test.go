package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"go.uber.org/zap"
)

// fakeNotificationRepo is an in-memory implementation of
// repository.NotificationRepository, used to exercise the REAL
// notificationService.SendBroadcast/ListNotifications logic (Staging blocker
// fix, client feedback item 10/16) without a database. Every method the
// interface requires is implemented (small interface, fully fakeable),
// unlike request_service's repositories which require a live *sqlx.Tx.
type fakeNotificationRepo struct {
	created []models.Notification
}

func (f *fakeNotificationRepo) Create(ctx context.Context, n *models.Notification) error {
	f.created = append(f.created, *n)
	return nil
}

func (f *fakeNotificationRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error) {
	var out []models.Notification
	for _, n := range f.created {
		if n.UserID == userID {
			out = append(out, n)
		}
	}
	return out, nil
}

func (f *fakeNotificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error { return nil }
func (f *fakeNotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return nil
}
func (f *fakeNotificationRepo) SavePushToken(ctx context.Context, token *models.PushToken) error {
	return nil
}
func (f *fakeNotificationRepo) GetPushTokens(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// No tokens registered for any user in these tests -- this is the exact
	// Staging scenario: a broadcast must still persist a notification for
	// every target user even when nobody has a live push token.
	return nil, nil
}
func (f *fakeNotificationRepo) DeactivateToken(ctx context.Context, fcmToken string) error { return nil }
func (f *fakeNotificationRepo) GetAllActiveTokens(ctx context.Context) ([]models.PushToken, error) {
	return nil, nil
}
func (f *fakeNotificationRepo) DeleteOld(ctx context.Context, days int) error { return nil }
func (f *fakeNotificationRepo) AdminList(ctx context.Context, userID uuid.UUID, status string, limit int) ([]models.Notification, error) {
	return nil, nil
}
func (f *fakeNotificationRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

var _ repository.NotificationRepository = (*fakeNotificationRepo)(nil)

// fakeUserRepoForBroadcast embeds the real interface and overrides only
// ListAllActiveUserIDs -- the one method SendBroadcast actually calls. Any
// other method would panic if called (nil embedded interface), which is
// intentional: it proves the test only exercises what SendBroadcast really
// touches.
type fakeUserRepoForBroadcast struct {
	repository.UserRepository
	activeUserIDs []uuid.UUID
}

func (f *fakeUserRepoForBroadcast) ListAllActiveUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	return f.activeUserIDs, nil
}

// TestSendBroadcast_PersistsPerTargetUser_NoDuplicates_ReadIsolated covers
// targeted tests for the Staging blocker fix (client feedback item 10/16):
//   - broadcast persists a notification for a target user, even with zero
//     push tokens registered (the exact Staging scenario -- previously
//     SendBroadcast iterated GetAllActiveTokens, so a user with no token
//     never got a row at all)
//   - the target user's notification list returns it
//   - one broadcast does not create unintended duplicates per user
//   - user A cannot read user B's notifications (ListByUserID is scoped)
func TestSendBroadcast_PersistsPerTargetUser_NoDuplicates_ReadIsolated(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()

	notifRepo := &fakeNotificationRepo{}
	userRepo := &fakeUserRepoForBroadcast{activeUserIDs: []uuid.UUID{userA, userB}}
	svc := NewNotificationService(notifRepo, userRepo, "", "", zap.NewNop(), nil)

	targetUsers, sent, failed, err := svc.SendBroadcast(context.Background(), "STAGING TEST", "hello", "promotion", nil)
	if err != nil {
		t.Fatalf("SendBroadcast returned an error: %v", err)
	}
	if targetUsers != 2 {
		t.Fatalf("expected 2 target users, got %d", targetUsers)
	}
	if sent != 2 {
		t.Fatalf("expected 2 successful sends (DB persistence, no push tokens needed), got %d", sent)
	}
	if failed != 0 {
		t.Fatalf("expected 0 failures, got %d", failed)
	}

	t.Run("broadcast persists a notification for a target user with zero push tokens", func(t *testing.T) {
		notifsA, err := svc.ListNotifications(context.Background(), userA, 10)
		if err != nil {
			t.Fatalf("ListNotifications failed: %v", err)
		}
		if len(notifsA) != 1 {
			t.Fatalf("expected user A to have exactly 1 persisted notification, got %d", len(notifsA))
		}
		if notifsA[0].Title != "STAGING TEST" {
			t.Fatalf("expected the persisted title to be 'STAGING TEST', got %q", notifsA[0].Title)
		}
	})

	t.Run("target user list returns the broadcast", func(t *testing.T) {
		notifsB, err := svc.ListNotifications(context.Background(), userB, 10)
		if err != nil {
			t.Fatalf("ListNotifications failed: %v", err)
		}
		if len(notifsB) != 1 {
			t.Fatalf("expected user B to have exactly 1 persisted notification, got %d", len(notifsB))
		}
	})

	t.Run("one broadcast does not create unintended duplicates for the same user", func(t *testing.T) {
		if len(notifRepo.created) != 2 {
			t.Fatalf("expected exactly 2 total persisted rows (1 per target user) for a single broadcast, got %d", len(notifRepo.created))
		}
	})

	t.Run("user A cannot read user B's notifications", func(t *testing.T) {
		notifsA, err := svc.ListNotifications(context.Background(), userA, 10)
		if err != nil {
			t.Fatalf("ListNotifications failed: %v", err)
		}
		for _, n := range notifsA {
			if n.UserID == userB {
				t.Fatalf("SECURITY REGRESSION: user A's notification list contained a row belonging to user B (id=%s)", n.ID)
			}
		}
	})
}

// TestSendBroadcast_NoActiveUsers_ReturnsZeroCounts is a defense-in-depth
// check that an empty user base doesn't error and reports zero counts
// rather than a misleading non-zero value.
func TestSendBroadcast_NoActiveUsers_ReturnsZeroCounts(t *testing.T) {
	notifRepo := &fakeNotificationRepo{}
	userRepo := &fakeUserRepoForBroadcast{activeUserIDs: nil}
	svc := NewNotificationService(notifRepo, userRepo, "", "", zap.NewNop(), nil)

	targetUsers, sent, failed, err := svc.SendBroadcast(context.Background(), "t", "b", "promotion", nil)
	if err != nil {
		t.Fatalf("expected no error for zero active users, got %v", err)
	}
	if targetUsers != 0 || sent != 0 || failed != 0 {
		t.Fatalf("expected all counts to be 0, got targetUsers=%d sent=%d failed=%d", targetUsers, sent, failed)
	}
}
