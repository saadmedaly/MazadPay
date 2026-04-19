package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
)

type UserService interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, fullName, email, city string) error
	UpdateAvatar(ctx context.Context, id uuid.UUID, url string) error
	UpdateLanguage(ctx context.Context, id uuid.UUID, lang string) error
	UpdateNotificationSettings(ctx context.Context, id uuid.UUID, enabled bool) error
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// Favorites
	AddFavorite(ctx context.Context, userID, auctionID uuid.UUID) error
	RemoveFavorite(ctx context.Context, userID, auctionID uuid.UUID) error
	ListFavorites(ctx context.Context, userID uuid.UUID) ([]models.Auction, error)

	// User Data
	ListMyAuctions(ctx context.Context, userID uuid.UUID) ([]models.Auction, error)
	ListMyBids(ctx context.Context, userID uuid.UUID) ([]models.Auction, error)
	ListMyWinnings(ctx context.Context, userID uuid.UUID) ([]models.Auction, error)

	// KYC
	SubmitKYC(ctx context.Context, kyc *models.KYCVerification) error
	GetKYCStatus(ctx context.Context, userID uuid.UUID) (*models.KYCVerification, error)
}

type userService struct {
	repo         repository.UserRepository
	favoriteRepo repository.FavoriteRepository
	auctionRepo  repository.AuctionRepository
	kycRepo      repository.KYCRepository
}

func NewUserService(
	repo repository.UserRepository,
	favoriteRepo repository.FavoriteRepository,
	auctionRepo repository.AuctionRepository,
	kycRepo repository.KYCRepository,
) UserService {
	return &userService{
		repo:         repo,
		favoriteRepo: favoriteRepo,
		auctionRepo:  auctionRepo,
		kycRepo:      kycRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *userService) UpdateProfile(ctx context.Context, id uuid.UUID, fullName, email, city string) error {
	return s.repo.UpdateProfile(ctx, id, fullName, email, city)
}

func (s *userService) UpdateAvatar(ctx context.Context, id uuid.UUID, url string) error {
	return s.repo.UpdateProfilePic(ctx, id, url)
}

func (s *userService) UpdateLanguage(ctx context.Context, id uuid.UUID, lang string) error {
	return s.repo.UpdateLanguage(ctx, id, lang)
}

func (s *userService) UpdateNotificationSettings(ctx context.Context, id uuid.UUID, enabled bool) error {
	return s.repo.UpdateNotificationSettings(ctx, id, enabled)
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.UpdateStatus(ctx, id, false)
}

func (s *userService) AddFavorite(ctx context.Context, userID, auctionID uuid.UUID) error {
	return s.favoriteRepo.Add(ctx, userID, auctionID)
}

func (s *userService) RemoveFavorite(ctx context.Context, userID, auctionID uuid.UUID) error {
	return s.favoriteRepo.Remove(ctx, userID, auctionID)
}

func (s *userService) ListFavorites(ctx context.Context, userID uuid.UUID) ([]models.Auction, error) {
	return s.favoriteRepo.ListByUserID(ctx, userID)
}

func (s *userService) ListMyAuctions(ctx context.Context, userID uuid.UUID) ([]models.Auction, error) {
	filters := repository.AuctionFilters{SellerID: &userID}
	auctions, _, err := s.auctionRepo.ListPaginated(ctx, 1, 100, filters)
	return auctions, err
}

func (s *userService) ListMyBids(ctx context.Context, userID uuid.UUID) ([]models.Auction, error) {
	// Logic to find auctions where user has bid
	return s.auctionRepo.ListByUserBids(ctx, userID)
}

func (s *userService) ListMyWinnings(ctx context.Context, userID uuid.UUID) ([]models.Auction, error) {
	filters := repository.AuctionFilters{WinnerID: &userID}
	auctions, _, err := s.auctionRepo.ListPaginated(ctx, 1, 100, filters)
	return auctions, err
}

func (s *userService) SubmitKYC(ctx context.Context, kyc *models.KYCVerification) error {
	kyc.Status = "pending"
	return s.kycRepo.Create(ctx, kyc)
}

func (s *userService) GetKYCStatus(ctx context.Context, userID uuid.UUID) (*models.KYCVerification, error) {
	return s.kycRepo.GetByUserID(ctx, userID)
}
