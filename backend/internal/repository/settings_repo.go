package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type SettingsRepository interface {
	Get(ctx context.Context, key string) (*models.SystemSettings, error)
	Set(ctx context.Context, key, value, settingType string) error
	List(ctx context.Context) ([]models.SystemSettings, error)
}

type settingsRepo struct {
	db *sqlx.DB
}

func NewSettingsRepository(db *sqlx.DB) SettingsRepository {
	return &settingsRepo{db: db}
}

func (r *settingsRepo) Get(ctx context.Context, key string) (*models.SystemSettings, error) {
	var s models.SystemSettings
	err := r.db.GetContext(ctx, &s, "SELECT * FROM system_settings WHERE key = $1", key)
	return &s, err
}

func (r *settingsRepo) Set(ctx context.Context, key, value, settingType string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO system_settings (key, value, type, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (key) DO UPDATE SET value = $2, type = $3, updated_at = now()
	`, key, value, settingType)
	return err
}

func (r *settingsRepo) List(ctx context.Context) ([]models.SystemSettings, error) {
	var settings []models.SystemSettings
	err := r.db.SelectContext(ctx, &settings, "SELECT * FROM system_settings ORDER BY key")
	return settings, err
}