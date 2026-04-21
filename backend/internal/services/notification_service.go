package services

import (
	"context"

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
	NotifyAdmins(ctx context.Context, title, body string, data map[string]string) error
	SendBroadcast(ctx context.Context, title, body, notifType string, data map[string]string) error
	ListNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	CleanupOldNotifications(ctx context.Context) error

	// Admin methods
	AdminListNotifications(ctx context.Context, status string, limit int) ([]models.Notification, error)
	DeleteNotification(ctx context.Context, id uuid.UUID) error
}

type notificationService struct {
	repo     repository.NotificationRepository
	userRepo repository.UserRepository
	fcm      *messaging.Client
	logger   *zap.Logger
}

func NewNotificationService(repo repository.NotificationRepository, userRepo repository.UserRepository, serviceAccountPath string, logger *zap.Logger) NotificationService {
	var fcmClient *messaging.Client

	if serviceAccountPath != "" {
		opt := option.WithCredentialsFile(serviceAccountPath)
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
	}

	return &notificationService{
		repo:     repo,
		userRepo: userRepo,
		fcm:      fcmClient,
		logger:   logger,
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

func (s *notificationService) SendBroadcast(ctx context.Context, title, body, notifType string, data map[string]string) error {
	// Get all users with active push tokens
	tokens, err := s.repo.GetAllActiveTokens(ctx)
	if err != nil {
		s.logger.Error("failed to get active tokens", zap.Error(err))
		return err
	}

	if len(tokens) == 0 {
		s.logger.Info("no active tokens found for broadcast")
		return nil
	}

	// Send to all tokens in batches
	for _, token := range tokens {
		_ = s.SendPush(ctx, token.UserID, title, body, notifType, data)
	}

	s.logger.Info("broadcast sent", zap.Int("count", len(tokens)))
	return nil
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

func (s *notificationService) AdminListNotifications(ctx context.Context, status string, limit int) ([]models.Notification, error) {
	return s.repo.AdminList(ctx, status, limit)
}

func (s *notificationService) DeleteNotification(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
