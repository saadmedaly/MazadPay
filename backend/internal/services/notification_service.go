package services

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"google.golang.org/api/option"
)

type NotificationService interface {
	SavePushToken(ctx context.Context, userID uuid.UUID, fcmToken, deviceID, platform string) error
	SendPush(ctx context.Context, userID uuid.UUID, title, body string, data map[string]string) error
	NotifyAdmins(ctx context.Context, title, body string, data map[string]string) error
	ListNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	CleanupOldNotifications(ctx context.Context) error
}

type notificationService struct {
	repo     repository.NotificationRepository
	userRepo repository.UserRepository
	fcm      *messaging.Client
}

func NewNotificationService(repo repository.NotificationRepository, userRepo repository.UserRepository, serviceAccountPath string) NotificationService {
	var fcmClient *messaging.Client

	if serviceAccountPath != "" {
		opt := option.WithCredentialsFile(serviceAccountPath)
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			log.Printf("error initializing firebase app: %v", err)
		} else {
			client, err := app.Messaging(context.Background())
			if err != nil {
				log.Printf("error getting messaging client: %v", err)
			} else {
				fcmClient = client
			}
		}
	}

	return &notificationService{
		repo:     repo,
		userRepo: userRepo,
		fcm:      fcmClient,
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

func (s *notificationService) SendPush(ctx context.Context, userID uuid.UUID, title, body string, data map[string]string) error {
	// 1. Log in database
	notification := &models.Notification{
		ID:     uuid.New(),
		UserID: userID,
		Type:   "push",
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
	_ = s.repo.Create(ctx, notification)

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
		_ = s.SendPush(ctx, admin.ID, title, body, data)
	}
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
