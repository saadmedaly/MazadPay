package services

import (
	"context"

	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
)

type ContentService interface {
	GetFAQ(ctx context.Context) ([]models.FAQItem, error)
	GetTutorials(ctx context.Context) ([]models.Tutorial, error)
	GetBanners(ctx context.Context, onlyActive bool) ([]models.Banner, error)

	// Admin CRUD
	CreateFAQ(ctx context.Context, item *models.FAQItem) error
	UpdateFAQ(ctx context.Context, item *models.FAQItem) error
	DeleteFAQ(ctx context.Context, id int) error
	CreateTutorial(ctx context.Context, tutorial *models.Tutorial) error
	UpdateTutorial(ctx context.Context, tutorial *models.Tutorial) error
	DeleteTutorial(ctx context.Context, id int) error
	CreateBanner(ctx context.Context, banner *models.Banner) error
	RequestBanner(ctx context.Context, banner *models.Banner) error
	ToggleBanner(ctx context.Context, id int, active bool) error
	UpdateBanner(ctx context.Context, banner *models.Banner) error
	DeleteBanner(ctx context.Context, id int) error
}

type contentService struct {
	repo     repository.ContentRepository
	notifSvc NotificationService
}

func NewContentService(repo repository.ContentRepository, notifSvc NotificationService) ContentService {
	return &contentService{
		repo:     repo,
		notifSvc: notifSvc,
	}
}

func (s *contentService) GetFAQ(ctx context.Context) ([]models.FAQItem, error) {
	return s.repo.ListFAQ(ctx)
}

func (s *contentService) GetTutorials(ctx context.Context) ([]models.Tutorial, error) {
	return s.repo.ListTutorials(ctx)
}

func (s *contentService) GetBanners(ctx context.Context, onlyActive bool) ([]models.Banner, error) {
	return s.repo.ListBanners(ctx, onlyActive)
}

func (s *contentService) CreateFAQ(ctx context.Context, item *models.FAQItem) error {
	return s.repo.CreateFAQ(ctx, item)
}

func (s *contentService) UpdateFAQ(ctx context.Context, item *models.FAQItem) error {
	return s.repo.UpdateFAQ(ctx, item)
}

func (s *contentService) DeleteFAQ(ctx context.Context, id int) error {
	return s.repo.DeleteFAQ(ctx, id)
}

func (s *contentService) CreateTutorial(ctx context.Context, tutorial *models.Tutorial) error {
	return s.repo.CreateTutorial(ctx, tutorial)
}

func (s *contentService) UpdateTutorial(ctx context.Context, tutorial *models.Tutorial) error {
	return s.repo.UpdateTutorial(ctx, tutorial)
}

func (s *contentService) DeleteTutorial(ctx context.Context, id int) error {
	return s.repo.DeleteTutorial(ctx, id)
}

func (s *contentService) CreateBanner(ctx context.Context, banner *models.Banner) error {
	return s.repo.CreateBanner(ctx, banner)
}

func (s *contentService) RequestBanner(ctx context.Context, banner *models.Banner) error {
	banner.IsActive = false // Les demandes sont inactives par défaut
	if err := s.repo.CreateBanner(ctx, banner); err != nil {
		return err
	}

	// Notifier les admins
	if s.notifSvc != nil {
		go func() {
			_ = s.notifSvc.NotifyAdmins(context.Background(),
				"💰 طلب إعلان جديد (Banner Request)",
				"إليك طلب جديد لإضافة إعلان على المنصة.",
				map[string]string{
					"type":  "banner_request",
					"title": banner.TitleAr,
				},
			)
		}()
	}

	return nil
}

func (s *contentService) ToggleBanner(ctx context.Context, id int, active bool) error {
	return s.repo.UpdateBannerStatus(ctx, id, active)
}

func (s *contentService) UpdateBanner(ctx context.Context, banner *models.Banner) error {
	return s.repo.UpdateBanner(ctx, banner)
}

func (s *contentService) DeleteBanner(ctx context.Context, id int) error {
	return s.repo.DeleteBanner(ctx, id)
}
