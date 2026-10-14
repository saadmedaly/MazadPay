package services

import (
	"context"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type NotificationService interface {
	SavePushToken(ctx context.Context, userID uuid.UUID, fcmToken, deviceID, platform string) error
	SendPush(ctx context.Context, userID uuid.UUID, title, body string, notifType string, data map[string]string) error
	SendLocalizedPush(ctx context.Context, userID uuid.UUID, notificationType, language string, params map[string]string, data map[string]string) error
	// SendPushWithImage (Customer #22) is SendPush plus an optional imageURL
	// persisted onto the created notification row -- a separate method rather
	// than adding a parameter to SendPush itself, so the 4 existing SendPush
	// call sites (none of which have an image to send) need no changes.
	// imageURL == "" behaves identically to SendPush.
	SendPushWithImage(ctx context.Context, userID uuid.UUID, title, body string, notifType string, data map[string]string, imageURL string) error
	NotifyAdmins(ctx context.Context, title, body string, data map[string]string) error
	NotifyAdminsLocalized(ctx context.Context, notificationType string, params map[string]string, data map[string]string) error
	// SendBroadcast returns (targetUsers, sent, failed, err) so the Admin
	// caller can report a real outcome (client feedback item 16: "add a
	// clear result: sent/failed/persisted") instead of a bare success/error.
	// "sent" here means SendPush succeeded for that user, which always
	// includes DB persistence (see SendPush) -- FCM delivery on top of that
	// is attempted per-user but never fails the persisted count.
	SendBroadcast(ctx context.Context, title, body, notifType string, data map[string]string) (targetUsers, sent, failed int, err error)
	// SendBroadcastWithImage (Customer #22) is SendBroadcast plus an optional
	// imageURL: the SAME URL string is reused across every per-user
	// notification row this broadcast creates (the image itself is uploaded
	// to R2 exactly once by the admin, before this is ever called -- see
	// NotificationHandler.SendNotification) -- never a separate
	// upload/object per recipient. imageURL == "" behaves identically to
	// SendBroadcast.
	SendBroadcastWithImage(ctx context.Context, title, body, notifType string, data map[string]string, imageURL string) (targetUsers, sent, failed int, err error)
	ListNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	CleanupOldNotifications(ctx context.Context) error

	// Admin methods
	AdminListNotifications(ctx context.Context, userID uuid.UUID, status string, limit int) ([]models.Notification, error)
	DeleteNotification(ctx context.Context, id uuid.UUID) error

	// WebSocket real-time notifications
	NotifyNewAuctionRequest(requestID, userID, userName, title string)
	NotifyNewBannerRequest(requestID, userID, userName, title string)
	NotifyRequestReviewed(requestID, requestType, status, updatedBy string)

}

type notificationService struct {
	repo      repository.NotificationRepository
	userRepo  repository.UserRepository
	fcm       *messaging.Client
	adminHub  AdminHub
	logger    *zap.Logger
}

func NewNotificationService(repo repository.NotificationRepository, userRepo repository.UserRepository, serviceAccountPath string, serviceAccountJSON string, logger *zap.Logger, adminHub AdminHub) NotificationService {
	var fcmClient *messaging.Client
	var opt option.ClientOption

	if serviceAccountJSON != "" {
		opt = option.WithCredentialsJSON([]byte(serviceAccountJSON))
	} else if serviceAccountPath != "" {
		if _, err := os.Stat(serviceAccountPath); err == nil {
			opt = option.WithCredentialsFile(serviceAccountPath)
		} else {
			logger.Warn("firebase credentials file not found, skipping fcm initialization", zap.String("path", serviceAccountPath))
		}
	}

	if opt != nil {
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			logger.Error("error initializing firebase app", zap.Error(err))
		} else {
			client, err := app.Messaging(context.Background())
			if err != nil {
				logger.Error("error getting messaging client", zap.Error(err))
			} else {
				fcmClient = client
			}
		}
	} else {
		logger.Warn("no firebase credentials provided (path or json), fcm will be disabled")
	}

	return &notificationService{
		repo:      repo,
		userRepo:  userRepo,
		fcm:       fcmClient,
		adminHub:  adminHub,
		logger:    logger,
	}
}

func (s *notificationService) SavePushToken(ctx context.Context, userID uuid.UUID, fcmToken, deviceID, platform string) error {
	token := &models.PushToken{
		ID:       uuid.New(),
		UserID:   userID,
		FCMToken: fcmToken,
		DeviceID: deviceID,
		Platform: platform,
		IsActive: true,
	}
	return s.repo.SavePushToken(ctx, token)
}

func (s *notificationService) SendPush(ctx context.Context, userID uuid.UUID, title, body string, notifType string, data map[string]string) error {
	return s.sendPush(ctx, userID, title, body, notifType, data, "")
}

func (s *notificationService) SendPushWithImage(ctx context.Context, userID uuid.UUID, title, body string, notifType string, data map[string]string, imageURL string) error {
	return s.sendPush(ctx, userID, title, body, notifType, data, imageURL)
}

// buildFCMData (Customer #22, Phase 5, extracted during final hardening for
// direct unit testing without a live *messaging.Client) builds the FCM data
// payload for a single recipient's push: a copy of the caller's own data map
// (never mutated -- SendBroadcast's per-user loop passes the same map to
// every recipient) plus notification_id, the just-created DB row's own ID.
// Because each call site passes that recipient's own freshly-created
// notification.ID, a broadcast to N users naturally produces N distinct
// notification_id values -- one real DB row ID per recipient, never a
// single ID shared across the broadcast -- letting a push tap look up its
// own row via the already-user-scoped GET /notifications.
func buildFCMData(data map[string]string, notificationID uuid.UUID) map[string]string {
	fcmData := make(map[string]string, len(data)+1)
	for k, v := range data {
		fcmData[k] = v
	}
	fcmData["notification_id"] = notificationID.String()
	return fcmData
}

// sendPush is the shared implementation behind SendPush/SendPushWithImage --
// imageURL == "" is the exact pre-Customer-#22 SendPush behavior (no image
// column set, no FCM ImageURL), so every existing caller of SendPush is
// unaffected.
func (s *notificationService) sendPush(ctx context.Context, userID uuid.UUID, title, body string, notifType string, data map[string]string, imageURL string) error {
	// 1. Log in database
	notification := &models.Notification{
		ID:     uuid.New(),
		UserID: userID,
		Type:   notifType,
		Title:  title,
		Body:   &body,
		IsRead: false,
	}
	if imageURL != "" {
		notification.ImageURL = &imageURL
	}
	if data != nil {
		notification.Data = make(models.JSONB)
		for k, v := range data {
			notification.Data[k] = v
		}
	}
	if err := s.repo.Create(ctx, notification); err != nil {
		s.logger.Error("error saving notification to db", zap.Error(err), zap.String("userID", userID.String()))
		// Continue anyway - don't fail the whole operation (FCM is still sent)
	}

	// Client feedback #10: users.notifications_enabled (set via PUT
	// /users/me/notification-prefs, client feedback #6) was written and read
	// back for display, but never actually consulted before sending a push --
	// the in-app notification row above is always created (the account
	// preference toggle only ever claimed to control "notifications", i.e.
	// push, and the in-app inbox is a separate concern the user never asked
	// to suppress), but the FCM push itself must be skipped when the user has
	// disabled it.
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.Warn("SendPush: could not load user to check notification preference, sending push anyway", zap.Error(err), zap.String("userID", userID.String()))
	} else if !user.NotificationsEnabled {
		return nil
	}

	// 2. Send via FCM
	if s.fcm == nil {
		return nil // FCM not configured
	}

	tokens, err := s.repo.GetPushTokens(ctx, userID)
	if err != nil || len(tokens) == 0 {
		return nil
	}

	fcmData := buildFCMData(data, notification.ID)

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: fcmData,
	}

	response, err := s.fcm.SendMulticast(ctx, message)
	if err != nil {
		return err
	}

	// Deactivate invalid tokens
	if response.FailureCount > 0 {
		for idx, resp := range response.Responses {
			if !resp.Success {
				// Optionally deactivate the token if it's invalid
				_ = s.repo.DeactivateToken(ctx, tokens[idx])
			}
		}
	}

	return nil
}

func (s *notificationService) NotifyAdmins(ctx context.Context, title, body string, data map[string]string) error {
	admins, err := s.userRepo.FindAllAdmins(ctx)
	if err != nil {
		return err
	}

	for _, admin := range admins {
		_ = s.SendPush(ctx, admin.ID, title, body, "system", data)
	}
	return nil
}

func (s *notificationService) SendBroadcast(ctx context.Context, title, body, notifType string, data map[string]string) (int, int, int, error) {
	return s.sendBroadcast(ctx, title, body, notifType, data, "")
}

func (s *notificationService) SendBroadcastWithImage(ctx context.Context, title, body, notifType string, data map[string]string, imageURL string) (int, int, int, error) {
	return s.sendBroadcast(ctx, title, body, notifType, data, imageURL)
}

// sendBroadcast is the shared implementation behind SendBroadcast/
// SendBroadcastWithImage -- imageURL == "" is the exact pre-Customer-#22
// SendBroadcast behavior.
func (s *notificationService) sendBroadcast(ctx context.Context, title, body, notifType string, data map[string]string, imageURL string) (int, int, int, error) {
	// Staging blocker fix (client feedback item 10/16 follow-up): this used
	// to iterate GetAllActiveTokens (push_tokens rows) instead of actual
	// users -- a user with no push token registered (or a deactivated one)
	// never got an in-app notification row created at all, since SendPush
	// was only ever invoked for token owners. That meant a broadcast that
	// showed as "sent" in Admin could be completely invisible in a target
	// user's mobile Notifications list whenever their device had no live
	// FCM token, which is exactly what Staging testing observed. Persistence
	// must not depend on FCM: iterate every active user once (so one
	// broadcast never creates more than one row per user even if they have
	// multiple push-token rows across devices), and let SendPush's own
	// token lookup handle FCM delivery as a secondary, best-effort channel.
	userIDs, err := s.userRepo.ListAllActiveUserIDs(ctx)
	if err != nil {
		s.logger.Error("failed to list active users for broadcast", zap.Error(err))
		return 0, 0, 0, err
	}

	if len(userIDs) == 0 {
		s.logger.Info("no active users found for broadcast")
		return 0, 0, 0, nil
	}

	sent := 0
	failed := 0
	for _, userID := range userIDs {
		if err := s.sendPush(ctx, userID, title, body, notifType, data, imageURL); err != nil {
			failed++
			s.logger.Warn("broadcast: failed to deliver to user", zap.String("user_id", userID.String()), zap.Error(err))
			continue
		}
		sent++
	}

	s.logger.Info("broadcast persisted/sent", zap.Int("target_users", len(userIDs)), zap.Int("sent", sent), zap.Int("failed", failed))
	return len(userIDs), sent, failed, nil
}

func (s *notificationService) ListNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error) {
	return s.repo.ListByUserID(ctx, userID, limit)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}

func (s *notificationService) CleanupOldNotifications(ctx context.Context) error {
	return s.repo.DeleteOld(ctx, 30)
}

func (s *notificationService) AdminListNotifications(ctx context.Context, userID uuid.UUID, status string, limit int) ([]models.Notification, error) {
	return s.repo.AdminList(ctx, userID, status, limit)
}

func (s *notificationService) DeleteNotification(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// WebSocket Real-time Notifications

func (s *notificationService) BroadcastNewRequest(requestType string, payload models.NewRequestPayload) {
	if s.adminHub == nil {
		s.logger.Warn("AdminHub not initialized, skipping real-time notification",
			zap.String("request_type", requestType),
			zap.String("request_id", payload.RequestID))
		return
	}
	s.adminHub.BroadcastNewRequest(requestType, payload)
}

func (s *notificationService) BroadcastRequestUpdated(payload models.RequestUpdatedPayload) {
	if s.adminHub == nil {
		s.logger.Warn("AdminHub not initialized, skipping real-time notification",
			zap.String("request_type", payload.RequestType),
			zap.String("request_id", payload.RequestID),
			zap.String("status", payload.Status))
		return
	}
	s.adminHub.BroadcastRequestUpdated(payload)
}

func (s *notificationService) NotifyNewAuctionRequest(requestID, userID, userName, title string) {
	s.BroadcastNewRequest("auction", models.NewRequestPayload{
		RequestID:   requestID,
		RequestType: "auction",
		UserID:      userID,
		UserName:    userName,
		Title:       title,
		CreatedAt:   time.Now().Format(time.RFC3339),
	})
}

func (s *notificationService) NotifyNewBannerRequest(requestID, userID, userName, title string) {
	s.BroadcastNewRequest("banner", models.NewRequestPayload{
		RequestID:   requestID,
		RequestType: "banner",
		UserID:      userID,
		UserName:    userName,
		Title:       title,
		CreatedAt:   time.Now().Format(time.RFC3339),
	})
}

func (s *notificationService) NotifyRequestReviewed(requestID, requestType, status, updatedBy string) {
	s.BroadcastRequestUpdated(models.RequestUpdatedPayload{
		RequestID:   requestID,
		RequestType: requestType,
		Status:      status,
		UpdatedBy:   updatedBy,
		UpdatedAt:   time.Now().Format(time.RFC3339),
	})
}

// SendLocalizedPush sends a notification with localized title and body
func (s *notificationService) SendLocalizedPush(ctx context.Context, userID uuid.UUID, notificationType, language string, params map[string]string, data map[string]string) error {
	title, body := GetLocalizedNotification(notificationType, language, params)
	if title == "" || body == "" {
		// Fallback to English if localization not found
		title, body = GetLocalizedNotification(notificationType, "en", params)
	}
	return s.SendPush(ctx, userID, title, body, notificationType, data)
}

// NotifyAdminsLocalized sends a localized notification to all admins
func (s *notificationService) NotifyAdminsLocalized(ctx context.Context, notificationType string, params map[string]string, data map[string]string) error {
	// Get admins
	admins, err := s.userRepo.FindAllAdmins(ctx)
	if err != nil {
		return err
	}

	// Send to each admin in their preferred language
	for _, admin := range admins {
		language := "ar" // Default to Arabic
		if admin.LanguagePref != "" {
			language = admin.LanguagePref
		}
		_ = s.SendLocalizedPush(ctx, admin.ID, notificationType, language, params, data)
	}
	return nil
}


