package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
)

type SponsorService interface {
	CreateSponsor(ctx context.Context, req *models.Sponsor) error
	UpdateSponsor(ctx context.Context, req *models.Sponsor) error
	DeleteSponsor(ctx context.Context, id uuid.UUID) error
	GetSponsorByID(ctx context.Context, id uuid.UUID) (*models.Sponsor, error)
	ListActiveSponsors(ctx context.Context, page, perPage int) ([]models.Sponsor, int, error)
	ListAllSponsors(ctx context.Context, page, perPage int) ([]models.Sponsor, int, error)
	ToggleSponsorStatus(ctx context.Context, id uuid.UUID) error
}

type sponsorService struct {
	repo repository.SponsorRepository
}

func NewSponsorService(repo repository.SponsorRepository) SponsorService {
	return &sponsorService{repo: repo}
}

func (s *sponsorService) CreateSponsor(ctx context.Context, req *models.Sponsor) error {
	req.ID = uuid.New()
	return s.repo.Create(ctx, req)
}

func (s *sponsorService) UpdateSponsor(ctx context.Context, req *models.Sponsor) error {
	return s.repo.Update(ctx, req)
}

func (s *sponsorService) DeleteSponsor(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *sponsorService) GetSponsorByID(ctx context.Context, id uuid.UUID) (*models.Sponsor, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *sponsorService) ListActiveSponsors(ctx context.Context, page, perPage int) ([]models.Sponsor, int, error) {
	offset := (page - 1) * perPage
	return s.repo.ListActive(ctx, perPage, offset)
}

func (s *sponsorService) ListAllSponsors(ctx context.Context, page, perPage int) ([]models.Sponsor, int, error) {
	return s.repo.ListAll(ctx, page, perPage)
}

func (s *sponsorService) ToggleSponsorStatus(ctx context.Context, id uuid.UUID) error {
	return s.repo.ToggleStatus(ctx, id)
}
