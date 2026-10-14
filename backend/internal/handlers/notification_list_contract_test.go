package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/services"
	"go.uber.org/zap"
)

// fakeNotificationServiceForList is a minimal services.NotificationService fake
// used only to drive NotificationHandler.List through the real Fiber
// request/response cycle, proving the exact wire-level JSON contract mobile
// consumes -- not a stand-in that skips the real serialization path. Embeds
// the real interface (nil) so only ListNotifications, the one method List()
// actually calls, needs an override.
type fakeNotificationServiceForList struct {
	services.NotificationService
	listResult []models.Notification
}

func (f *fakeNotificationServiceForList) ListNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error) {
	return f.listResult, nil
}

// TestNotificationList_HTTP_ResponseContractMatchesMobileExpectation is the
// regression test for the real-device Staging failure (client feedback
// NOTIFICATIONS root cause): GET /notifications used to respond with
// {"data": [...]} and no "success" field, unlike every other endpoint. Since
// mobile/lib/models/api_response.dart's ApiResponse.fromJson defaults
// `success` to false whenever the key is absent, and notifications_page.dart
// only populates its list `if (response.success && response.data != null)`,
// a broadcast that was genuinely persisted (54/54 per Admin) still rendered
// as "لا توجد إشعارات" on the device. This test asserts the wire JSON shape
// directly: top-level "success" must be true, and "data" must be the bare
// array mobile's `responseData is List` branch expects.
func TestNotificationList_HTTP_ResponseContractMatchesMobileExpectation(t *testing.T) {
	userID := uuid.New()
	notifID := uuid.New()
	svc := &fakeNotificationServiceForList{
		listResult: []models.Notification{
			{
				ID:        notifID,
				UserID:    userID,
				Type:      "general",
				Title:     "STAGING TEST 2",
				IsRead:    false,
				CreatedAt: time.Now(),
			},
		},
	}
	h := NewNotificationHandler(svc, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/notifications", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return h.List(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/api/notifications", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("failed to parse response JSON: %v, body: %s", err, body)
	}

	// This is the exact assertion mobile's ApiResponse.fromJson makes:
	// json['success'] as bool? ?? false.
	success, ok := parsed["success"].(bool)
	if !ok || !success {
		t.Fatalf(`REGRESSION: expected top-level "success": true (mobile ApiResponse.fromJson defaults to false when this key is absent, discarding a valid response), got: %v`, parsed["success"])
	}

	dataField, ok := parsed["data"]
	if !ok {
		t.Fatal(`expected top-level "data" field, none found`)
	}
	dataArray, ok := dataField.([]interface{})
	if !ok {
		t.Fatalf(`expected top-level "data" to be a bare JSON array (matching notifications_page.dart's "responseData is List" branch), got %T: %v`, dataField, dataField)
	}
	if len(dataArray) != 1 {
		t.Fatalf("expected 1 notification in the response, got %d", len(dataArray))
	}

	item, ok := dataArray[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected notification item to be an object, got %T", dataArray[0])
	}
	if item["title"] != "STAGING TEST 2" {
		t.Fatalf(`expected notification title "STAGING TEST 2", got %v`, item["title"])
	}
}

// TestNotificationList_HTTP_UnauthenticatedRejected proves the fix didn't
// weaken the existing 401 behavior for a request with no authenticated user.
func TestNotificationList_HTTP_UnauthenticatedRejected(t *testing.T) {
	svc := &fakeNotificationServiceForList{}
	h := NewNotificationHandler(svc, zap.NewNop())

	app := fiber.New()
	app.Get("/v1/api/notifications", h.List)

	req := httptest.NewRequest(http.MethodGet, "/v1/api/notifications", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected HTTP 401 for an unauthenticated request, got %d", resp.StatusCode)
	}
}
