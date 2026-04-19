package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type ContentRepository interface {
	// FAQ
	ListFAQ(ctx context.Context) ([]models.FAQItem, error)
	CreateFAQ(ctx context.Context, item *models.FAQItem) error
	DeleteFAQ(ctx context.Context, id int) error
	UpdateFAQ(ctx context.Context, item *models.FAQItem) error

	// Tutorials
	ListTutorials(ctx context.Context) ([]models.Tutorial, error)
	CreateTutorial(ctx context.Context, tutorial *models.Tutorial) error
	DeleteTutorial(ctx context.Context, id int) error
	UpdateTutorial(ctx context.Context, tutorial *models.Tutorial) error

	// Banners
	ListBanners(ctx context.Context, onlyActive bool) ([]models.Banner, error)
	CreateBanner(ctx context.Context, banner *models.Banner) error
	UpdateBannerStatus(ctx context.Context, id int, isActive bool) error
	UpdateBanner(ctx context.Context, banner *models.Banner) error
	DeleteBanner(ctx context.Context, id int) error
}

type contentRepo struct {
	db *sqlx.DB
}

func NewContentRepository(db *sqlx.DB) ContentRepository {
	return &contentRepo{db: db}
}

func (r *contentRepo) ListFAQ(ctx context.Context) ([]models.FAQItem, error) {
	var items []models.FAQItem
	err := r.db.SelectContext(ctx, &items, `SELECT * FROM faq_items ORDER BY display_order ASC`)
	return items, err
}

func (r *contentRepo) CreateFAQ(ctx context.Context, item *models.FAQItem) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO faq_items (question_ar, question_fr, answer_ar, answer_fr, display_order)
		VALUES (:question_ar, :question_fr, :answer_ar, :answer_fr, :display_order)
	`, item)
	return err
}

func (r *contentRepo) DeleteFAQ(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM faq_items WHERE id = $1`, id)
	return err
}

func (r *contentRepo) UpdateFAQ(ctx context.Context, item *models.FAQItem) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE faq_items SET 
			question_ar = :question_ar, 
			question_fr = :question_fr, 
			answer_ar = :answer_ar, 
			answer_fr = :answer_fr, 
			display_order = :display_order
		WHERE id = :id
	`, item)
	return err
}

func (r *contentRepo) ListTutorials(ctx context.Context) ([]models.Tutorial, error) {
	var items []models.Tutorial
	err := r.db.SelectContext(ctx, &items, `SELECT * FROM tutorials ORDER BY display_order ASC`)
	return items, err
}

func (r *contentRepo) CreateTutorial(ctx context.Context, tutorial *models.Tutorial) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO tutorials (title_ar, title_fr, video_url, thumbnail_url, category, display_order)
		VALUES (:title_ar, :title_fr, :video_url, :thumbnail_url, :category, :display_order)
	`, tutorial)
	return err
}

func (r *contentRepo) DeleteTutorial(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tutorials WHERE id = $1`, id)
	return err
}

func (r *contentRepo) UpdateTutorial(ctx context.Context, tutorial *models.Tutorial) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE tutorials SET 
			title_ar = :title_ar, 
			title_fr = :title_fr, 
			video_url = :video_url, 
			thumbnail_url = :thumbnail_url, 
			category = :category, 
			display_order = :display_order
		WHERE id = :id
	`, tutorial)
	return err
}

func (r *contentRepo) ListBanners(ctx context.Context, onlyActive bool) ([]models.Banner, error) {
	query := `SELECT id, title_ar, title_fr, title_en, image_url, target_url, is_active, starts_at, ends_at, display_order FROM banners`
	if onlyActive {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY display_order ASC`

	var banners []models.Banner
	err := r.db.SelectContext(ctx, &banners, query)
	if err != nil {
		return nil, err
	}
	if banners == nil {
		banners = []models.Banner{}
	}
	return banners, nil
}

func (r *contentRepo) CreateBanner(ctx context.Context, banner *models.Banner) error {
	query := `
		INSERT INTO banners (title_ar, title_fr, title_en, image_url, target_url, is_active, starts_at, ends_at, display_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`
	err := r.db.GetContext(ctx, banner, query,
		banner.TitleAr, banner.TitleFr, banner.TitleEn,
		banner.ImageURL, banner.TargetURL, banner.IsActive,
		banner.StartsAt, banner.EndsAt, banner.DisplayOrder)
	return err
}

func (r *contentRepo) UpdateBannerStatus(ctx context.Context, id int, isActive bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE banners SET is_active = $1 WHERE id = $2`, isActive, id)
	return err
}

func (r *contentRepo) UpdateBanner(ctx context.Context, banner *models.Banner) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE banners SET
			title_ar = :title_ar,
			title_fr = :title_fr,
			title_en = :title_en,
			image_url = :image_url,
			target_url = :target_url,
			is_active = :is_active,
			starts_at = :starts_at,
			ends_at = :ends_at,
			display_order = :display_order
		WHERE id = :id
	`, banner)
	return err
}

func (r *contentRepo) DeleteBanner(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM banners WHERE id = $1`, id)
	return err
}
