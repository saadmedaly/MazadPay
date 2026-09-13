package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
	repo      repository.ContentRepository
	notifSvc  NotificationService
	mediaSvc  MediaService
	globalHub GlobalHub
}

func NewContentService(repo repository.ContentRepository, notifSvc NotificationService, mediaSvc MediaService, globalHub GlobalHub) ContentService {
	return &contentService{
		repo:      repo,
		notifSvc:  notifSvc,
		mediaSvc:  mediaSvc,
		globalHub: globalHub,
	}
}

// emitContentEvent broadcasts a Customer #20 global invalidation event for
// FAQ/tutorial/banner content -- these are global (not market-scoped) so
// they use globalHub.Broadcast (all connected clients), unlike auction
// events which are filtered by market (see auctionService.emitAuctionEvent).
// Called only after the triggering write already succeeded.
func (s *contentService) emitContentEvent(eventType, entityType string, entityID int) {
	if s.globalHub == nil {
		return
	}
	s.globalHub.Broadcast(models.GlobalWSEvent{
		Type:       eventType,
		EntityType: entityType,
		EntityID:   strconv.Itoa(entityID),
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	})
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
	if err := s.repo.CreateFAQ(ctx, item); err != nil {
		return err
	}
	s.emitContentEvent(models.EventFAQUpdated, "faq", item.ID)
	return nil
}

func (s *contentService) UpdateFAQ(ctx context.Context, item *models.FAQItem) error {
	if err := s.repo.UpdateFAQ(ctx, item); err != nil {
		return err
	}
	s.emitContentEvent(models.EventFAQUpdated, "faq", item.ID)
	return nil
}

func (s *contentService) DeleteFAQ(ctx context.Context, id int) error {
	if err := s.repo.DeleteFAQ(ctx, id); err != nil {
		return err
	}
	s.emitContentEvent(models.EventFAQUpdated, "faq", id)
	return nil
}

func (s *contentService) CreateTutorial(ctx context.Context, tutorial *models.Tutorial) error {
	return s.repo.CreateTutorial(ctx, tutorial)
}

func (s *contentService) UpdateTutorial(ctx context.Context, tutorial *models.Tutorial) error {
	return s.repo.UpdateTutorial(ctx, tutorial)
}

func (s *contentService) DeleteTutorial(ctx context.Context, id int) error {
	// Get tutorial before deletion (for R2 cleanup)
	tutorial, err := s.repo.GetTutorialByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get tutorial: %w", err)
	}

	// Delete from DB
	if err := s.repo.DeleteTutorial(ctx, id); err != nil {
		return err
	}

	// After successful DB deletion, delete files from R2 (best effort)
	if s.mediaSvc != nil {
		if tutorial.VideoURL != "" {
			if err := s.mediaSvc.DeleteFile(ctx, tutorial.VideoURL); err != nil {
				fmt.Printf("[DeleteTutorial] Warning: failed to delete video from R2: %s, error: %v\n", tutorial.VideoURL, err)
			}
		}
		if tutorial.ThumbnailURL != nil {
			if err := s.mediaSvc.DeleteFile(ctx, *tutorial.ThumbnailURL); err != nil {
				fmt.Printf("[DeleteTutorial] Warning: failed to delete thumbnail from R2: %s, error: %v\n", *tutorial.ThumbnailURL, err)
			}
		}
	}

	return nil
}

func (s *contentService) CreateBanner(ctx context.Context, banner *models.Banner) error {
	if err := s.repo.CreateBanner(ctx, banner); err != nil {
		return err
	}
	s.emitContentEvent(models.EventBannerUpdated, "banner", banner.ID)
	return nil
}

func (s *contentService) RequestBanner(ctx context.Context, banner *models.Banner) error {
	banner.IsActive = false // Les demandes sont inactives par défaut
	if err := s.repo.CreateBanner(ctx, banner); err != nil {
		return err
	}

	// Notifier les admins (localized)
	if s.notifSvc != nil {
		go func() {
			_ = s.notifSvc.NotifyAdminsLocalized(context.Background(), "banner_request", map[string]string{
				"bannerTitle": banner.TitleAr,
			}, map[string]string{
				"type":  "banner_request",
				"title": banner.TitleAr,
			})
		}()
	}

	// Not broadcast: an inactive/pending banner request isn't visible to
	// mobile yet -- see UpdateBanner/ToggleBanner, which emit once an admin
	// actually approves/activates it.
	return nil
}

func (s *contentService) ToggleBanner(ctx context.Context, id int, active bool) error {
	if err := s.repo.UpdateBannerStatus(ctx, id, active); err != nil {
		return err
	}
	s.emitContentEvent(models.EventBannerUpdated, "banner", id)
	return nil
}

func (s *contentService) UpdateBanner(ctx context.Context, banner *models.Banner) error {
	if err := s.repo.UpdateBanner(ctx, banner); err != nil {
		return err
	}
	s.emitContentEvent(models.EventBannerUpdated, "banner", banner.ID)
	return nil
}

func (s *contentService) DeleteBanner(ctx context.Context, id int) error {
	// Get banner before deletion (for R2 cleanup)
	banner, err := s.repo.GetBannerByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get banner: %w", err)
	}

	// Delete from DB
	if err := s.repo.DeleteBanner(ctx, id); err != nil {
		return err
	}

	// After successful DB deletion, delete image from R2 (best effort)
	if s.mediaSvc != nil && banner.ImageURL != "" {
		if err := s.mediaSvc.DeleteFile(ctx, banner.ImageURL); err != nil {
			fmt.Printf("[DeleteBanner] Warning: failed to delete image from R2: %s, error: %v\n", banner.ImageURL, err)
		}
	}

	s.emitContentEvent(models.EventBannerUpdated, "banner", id)
	return nil
}
