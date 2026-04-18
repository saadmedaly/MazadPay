package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *models.Notification) error
	ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	
	// Push Tokens
	SavePushToken(ctx context.Context, token *models.PushToken) error
	GetPushTokens(ctx context.Context, userID uuid.UUID) ([]string, error)
	DeactivateToken(ctx context.Context, fcmToken string) error
}

type notificationRepo struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(ctx context.Context, n *models.Notification) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO notifications (id, user_id, type, title, body, reference_id, reference_type, data)
		VALUES (:id, :user_id, :type, :title, :body, :reference_id, :reference_type, :data)
	`, n)
	return err
}

func (r *notificationRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.SelectContext(ctx, &notifications, `
		SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2
	`, userID, limit)
	return notifications, err
}

func (r *notificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET is_read = true WHERE user_id = $1`, userID)
	return err
}

func (r *notificationRepo) SavePushToken(ctx context.Context, token *models.PushToken) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO push_tokens (id, user_id, fcm_token, device_id, platform, is_active)
		VALUES (:id, :user_id, :fcm_token, :device_id, :platform, :is_active)
		ON CONFLICT (fcm_token) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			is_active = true,
			updated_at = now()
	`, token)
	return err
}

func (r *notificationRepo) GetPushTokens(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var tokens []string
	err := r.db.SelectContext(ctx, &tokens, `SELECT fcm_token FROM push_tokens WHERE user_id = $1 AND is_active = true`, userID)
	return tokens, err
}

func (r *notificationRepo) DeactivateToken(ctx context.Context, fcmToken string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE push_tokens SET is_active = false WHERE fcm_token = $1`, fcmToken)
	return err
}
