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
	NotifyAdmins(ctx context.Context, title, body string, data map[string]string) error
	NotifyAdminsLocalized(ctx context.Context, notificationType string, params map[string]string, data map[string]string) error
	// SendBroadcast returns (targetUsers, sent, failed, err) so the Admin
	// caller can report a real outcome (client feedback item 16: "add a
	// clear result: sent/failed/persisted") instead of a bare success/error.
	// "sent" here means SendPush succeeded for that user, which always
	// includes DB persistence (see SendPush) -- FCM delivery on top of that
	// is attempted per-user but never fails the persisted count.
	SendBroadcast(ctx context.Context, title, body, notifType string, data map[string]string) (targetUsers, sent, failed int, err error)
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
	// 1. Log in database
	notification := &models.Notification{
		ID:     uuid.New(),
		UserID: userID,
		Type:   notifType,
		Title:  title,
		Body:   &body,
		IsRead: false,
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

	// 2. Send via FCM
	if s.fcm == nil {
		return nil // FCM not configured
	}

	tokens, err := s.repo.GetPushTokens(ctx, userID)
	if err != nil || len(tokens) == 0 {
		return nil
	}

	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
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
		if err := s.SendPush(ctx, userID, title, body, notifType, data); err != nil {
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


