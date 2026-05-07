package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/models"
)

type SponsorRepository interface {
	Create(ctx context.Context, sponsor *models.Sponsor) error
	Update(ctx context.Context, sponsor *models.Sponsor) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Sponsor, error)
	ListActive(ctx context.Context, limit, offset int) ([]models.Sponsor, int, error)
	ListAll(ctx context.Context, page, perPage int) ([]models.Sponsor, int, error)
	ToggleStatus(ctx context.Context, id uuid.UUID) error
}

type sponsorRepo struct {
	db *sqlx.DB
}

func NewSponsorRepository(db *sqlx.DB) SponsorRepository {
	return &sponsorRepo{db: db}
}

func (r *sponsorRepo) Create(ctx context.Context, sponsor *models.Sponsor) error {
	query := `
		INSERT INTO sponsors (id, name,phone, image_url, link_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(ctx, query,
		sponsor.ID, sponsor.Name, sponsor.Phone, sponsor.ImageURL, sponsor.LinkURL,
		sponsor.IsActive, sponsor.CreatedAt, sponsor.UpdatedAt,
	)
	return err
}

func (r *sponsorRepo) Update(ctx context.Context, sponsor *models.Sponsor) error {
	query := `
		UPDATE sponsors 
		SET name = $1, phone = $2, image_url = $3, link_url = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.ExecContext(ctx, query,
		sponsor.Name, sponsor.Phone, sponsor.ImageURL, sponsor.LinkURL, sponsor.IsActive, sponsor.ID,
	)
	return err
}

func (r *sponsorRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sponsors WHERE id = $1", id)
	return err
}

func (r *sponsorRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Sponsor, error) {
	var sponsor models.Sponsor
	err := r.db.GetContext(ctx, &sponsor, "SELECT * FROM sponsors WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &sponsor, nil
}

func (r *sponsorRepo) ListActive(ctx context.Context, limit, offset int) ([]models.Sponsor, int, error) {
	var total int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM sponsors WHERE is_active = TRUE")
	if err != nil {
		return nil, 0, err
	}

	sponsors := []models.Sponsor{}
	query := "SELECT * FROM sponsors WHERE is_active = TRUE ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	err = r.db.SelectContext(ctx, &sponsors, query, limit, offset)
	
	return sponsors, total, err
}

func (r *sponsorRepo) ListAll(ctx context.Context, page, perPage int) ([]models.Sponsor, int, error) {
	var total int
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM sponsors")
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	sponsors := []models.Sponsor{}
	query := "SELECT * FROM sponsors ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	err = r.db.SelectContext(ctx, &sponsors, query, perPage, offset)
	
	return sponsors, total, err
}

func (r *sponsorRepo) ToggleStatus(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE sponsors SET is_active = NOT is_active, updated_at = NOW() WHERE id = $1", id)
	return err
}
