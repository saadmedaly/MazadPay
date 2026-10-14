package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/mazadpay/backend/internal/models"
)

// Customer Request #22 hardening round: notificationRepo.Create was modified
// to add image_url/action_url/action_label to its INSERT. This proves that
// change did not regress the other pre-existing supported fields
// (reference_id, reference_type, data, priority) -- a real Postgres
// round-trip against the local dev/test container (mazadpay_postgres,
// DB_PORT=5433 per backend/.env), never Production. Self-skips if that
// container isn't reachable, so it never blocks environments without it.

func connectNotificationTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5433")
	user := getEnvOrDefault("DB_USER", "mazadpay")
	pass := getEnvOrDefault("DB_PASSWORD", "mazadpay_secret")
	name := getEnvOrDefault("DB_NAME", "mazadpay")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable client_encoding=UTF8",
		host, port, user, pass, name)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		t.Skipf("skipping: could not open connection to local test Postgres: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Skipf("skipping: local test Postgres not reachable at %s:%s (%v) -- this test requires the mazadpay_postgres dev container", host, port, err)
	}
	return db
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// createTestUser inserts a minimal throwaway user row and registers cleanup
// to delete it (and, via ON DELETE CASCADE on notifications.user_id, any
// notification rows created for it) when the test finishes.
func createTestUser(t *testing.T, db *sqlx.DB) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	phone := "+22200000" + userID.String()[0:4]

	_, err := db.Exec(
		`INSERT INTO users (id, phone, password_hash, role) VALUES ($1, $2, $3, 'user')`,
		userID, phone, "test-hash-not-a-real-password",
	)
	if err != nil {
		t.Fatalf("failed to insert test user fixture: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	})
	return userID
}

// TestNotificationRepo_Create_RoundTrips_AllSupportedFields is the primary
// hardening-round proof: image_url, action_url, action_label, reference_id,
// reference_type, and data (JSONB) all survive a real Create ->
// SELECT-back round trip. priority is asserted separately below since
// Create relies on the column's own DB default rather than setting it.
func TestNotificationRepo_Create_RoundTrips_AllSupportedFields(t *testing.T) {
	db := connectNotificationTestDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)
	userID := createTestUser(t, db)

	refID := uuid.New()
	refType := "auction"
	imageURL := "https://cdn.example.com/notifications/round-trip.jpg"
	actionURL := "/auction/" + refID.String()
	actionLabel := "View auction"
	body := "round-trip test body"

	n := &models.Notification{
		ID:            uuid.New(),
		UserID:        userID,
		Type:          "general",
		Title:         "Round-trip test",
		Body:          &body,
		IsRead:        false,
		ReferenceID:   &refID,
		ReferenceType: &refType,
		Data:          models.JSONB{"auction_id": refID.String(), "extra": "value"},
		ImageURL:      &imageURL,
		ActionURL:     &actionURL,
		ActionLabel:   &actionLabel,
	}

	if err := repo.Create(context.Background(), n); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var got models.Notification
	err := db.Get(&got, `SELECT * FROM notifications WHERE id = $1`, n.ID)
	if err != nil {
		t.Fatalf("failed to read back the created notification: %v", err)
	}

	t.Run("IMAGE_URL_ROUNDTRIP", func(t *testing.T) {
		if got.ImageURL == nil || *got.ImageURL != imageURL {
			t.Fatalf("expected image_url %q, got %v", imageURL, got.ImageURL)
		}
	})

	t.Run("ACTION_URL_ROUNDTRIP", func(t *testing.T) {
		if got.ActionURL == nil || *got.ActionURL != actionURL {
			t.Fatalf("expected action_url %q, got %v", actionURL, got.ActionURL)
		}
	})

	t.Run("ACTION_LABEL_ROUNDTRIP", func(t *testing.T) {
		if got.ActionLabel == nil || *got.ActionLabel != actionLabel {
			t.Fatalf("expected action_label %q, got %v", actionLabel, got.ActionLabel)
		}
	})

	t.Run("REFERENCE_FIELDS_PRESERVED", func(t *testing.T) {
		if got.ReferenceID == nil || *got.ReferenceID != refID {
			t.Fatalf("expected reference_id %s, got %v", refID, got.ReferenceID)
		}
		if got.ReferenceType == nil || *got.ReferenceType != refType {
			t.Fatalf("expected reference_type %q, got %v", refType, got.ReferenceType)
		}
	})

	t.Run("DATA_JSON_PRESERVED", func(t *testing.T) {
		if got.Data == nil {
			t.Fatalf("expected data JSONB to be persisted, got nil")
		}
		if got.Data["auction_id"] != refID.String() {
			t.Fatalf("expected data.auction_id %q, got %v", refID.String(), got.Data["auction_id"])
		}
		if got.Data["extra"] != "value" {
			t.Fatalf("expected data.extra %q, got %v", "value", got.Data["extra"])
		}
	})

	t.Run("PRIORITY_PRESERVED", func(t *testing.T) {
		// Create does not set priority explicitly (not invented behavior --
		// this proves the column's own DB default ('normal') still applies
		// and the column is not left NULL/broken by the Create change).
		if got.Priority == nil {
			t.Fatalf("expected priority to have a value via the column default, got nil")
		}
		if *got.Priority != "normal" {
			t.Fatalf("expected default priority 'normal', got %q", *got.Priority)
		}
	})
}

// TestNotificationRepo_Create_NoImage_NoActionFields_RemainNull proves the
// pre-existing text-only-notification shape (nil ImageURL/ActionURL/
// ActionLabel/ReferenceID/ReferenceType) still round-trips to NULL columns,
// not empty strings or zero UUIDs -- matching every pre-Customer-#22
// notification (auction/payment/wallet) which never set these fields.
func TestNotificationRepo_Create_NoImage_NoActionFields_RemainNull(t *testing.T) {
	db := connectNotificationTestDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)
	userID := createTestUser(t, db)

	body := "plain text-only notification"
	n := &models.Notification{
		ID:     uuid.New(),
		UserID: userID,
		Type:   "auction_won",
		Title:  "You won!",
		Body:   &body,
	}

	if err := repo.Create(context.Background(), n); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	var got models.Notification
	if err := db.Get(&got, `SELECT * FROM notifications WHERE id = $1`, n.ID); err != nil {
		t.Fatalf("failed to read back: %v", err)
	}

	if got.ImageURL != nil {
		t.Fatalf("expected image_url NULL for a text-only notification, got %q", *got.ImageURL)
	}
	if got.ActionURL != nil {
		t.Fatalf("expected action_url NULL, got %q", *got.ActionURL)
	}
	if got.ActionLabel != nil {
		t.Fatalf("expected action_label NULL, got %q", *got.ActionLabel)
	}
	if got.ReferenceID != nil {
		t.Fatalf("expected reference_id NULL, got %v", *got.ReferenceID)
	}
	if got.ReferenceType != nil {
		t.Fatalf("expected reference_type NULL, got %q", *got.ReferenceType)
	}
}
