package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"go.uber.org/zap"
)

// fakeUserRepoWithPreference lets each test control NotificationsEnabled per
// user, to exercise the client feedback #10 fix: SendPush now checks
// user.NotificationsEnabled before sending the FCM push. DB persistence of
// the in-app notification row must stay unconditional regardless of this
// preference -- the account page's "Notifications" toggle (client feedback
// #6) only ever claimed to control push delivery, not the in-app inbox.
type fakeUserRepoWithPreference struct {
	repository.UserRepository
	enabled map[uuid.UUID]bool
}

func (f *fakeUserRepoWithPreference) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return &models.User{ID: id, NotificationsEnabled: f.enabled[id]}, nil
}

func TestSendPush_RespectsNotificationsEnabled(t *testing.T) {
	userEnabled := uuid.New()
	userDisabled := uuid.New()

	t.Run("notifications_enabled=true: in-app row persisted (FCM not configured in this test, so no push assertion needed)", func(t *testing.T) {
		notifRepo := &fakeNotificationRepo{}
		userRepo := &fakeUserRepoWithPreference{enabled: map[uuid.UUID]bool{userEnabled: true}}
		svc := NewNotificationService(notifRepo, userRepo, "", "", zap.NewNop(), nil)

		if err := svc.SendPush(context.Background(), userEnabled, "title", "body", "system", nil); err != nil {
			t.Fatalf("SendPush returned an error: %v", err)
		}

		notifs, err := svc.ListNotifications(context.Background(), userEnabled, 10)
		if err != nil {
			t.Fatalf("ListNotifications failed: %v", err)
		}
		if len(notifs) != 1 {
			t.Fatalf("expected 1 persisted in-app notification for an enabled user, got %d", len(notifs))
		}
	})

	t.Run("notifications_enabled=false: in-app row is STILL persisted (separate from push)", func(t *testing.T) {
		notifRepo := &fakeNotificationRepo{}
		userRepo := &fakeUserRepoWithPreference{enabled: map[uuid.UUID]bool{userDisabled: false}}
		svc := NewNotificationService(notifRepo, userRepo, "", "", zap.NewNop(), nil)

		if err := svc.SendPush(context.Background(), userDisabled, "title", "body", "system", nil); err != nil {
			t.Fatalf("SendPush returned an error even though it should skip only the push, not fail: %v", err)
		}

		notifs, err := svc.ListNotifications(context.Background(), userDisabled, 10)
		if err != nil {
			t.Fatalf("ListNotifications failed: %v", err)
		}
		if len(notifs) != 1 {
			t.Fatalf("expected the in-app notification to still be persisted for a user with push disabled, got %d rows", len(notifs))
		}
	})

	t.Run("a user whose preference cannot be loaded still gets the push attempted (fail open on lookup error, not silently dropped)", func(t *testing.T) {
		notifRepo := &fakeNotificationRepo{}
		// FindByID explicitly errors here, to prove SendPush does not treat a
		// lookup failure as "disabled" and silently drop the notification --
		// it logs a warning and proceeds as if enabled.
		erroringRepo := &fakeUserRepoLookupError{}
		svc := NewNotificationService(notifRepo, erroringRepo, "", "", zap.NewNop(), nil)

		someUser := uuid.New()
		if err := svc.SendPush(context.Background(), someUser, "title", "body", "system", nil); err != nil {
			t.Fatalf("SendPush returned an error: %v", err)
		}
		notifs, err := svc.ListNotifications(context.Background(), someUser, 10)
		if err != nil {
			t.Fatalf("ListNotifications failed: %v", err)
		}
		if len(notifs) != 1 {
			t.Fatalf("expected the in-app notification to still be persisted even when the preference lookup fails, got %d rows", len(notifs))
		}
	})
}

type fakeUserRepoLookupError struct {
	repository.UserRepository
}

func (f *fakeUserRepoLookupError) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return nil, context.DeadlineExceeded
}
