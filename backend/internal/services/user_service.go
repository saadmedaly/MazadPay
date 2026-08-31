package services

import (
	"context"

	"github.com/google/uuid"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type UserService interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, fullName, email, city string) error
	UpdateProfileExtended(ctx context.Context, id uuid.UUID, fullName, email, city, countryCode, address, postalCode, dateOfBirth, gender string) error
	UpdateAvatar(ctx context.Context, id uuid.UUID, url string) error
	UpdateLanguage(ctx context.Context, id uuid.UUID, lang string) error
	UpdateNotificationSettings(ctx context.Context, id uuid.UUID, enabled bool) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	DeactivateOwnAccount(ctx context.Context, userID uuid.UUID) error

	// Favorites
	AddFavorite(ctx context.Context, userID, auctionID uuid.UUID) error
	RemoveFavorite(ctx context.Context, userID, auctionID uuid.UUID) error
	ListFavorites(ctx context.Context, userID uuid.UUID) ([]models.Auction, error)

	// User Data
	ListMyAuctions(ctx context.Context, userID uuid.UUID) ([]models.Auction, error)
	ListMyAuctionsByStatus(ctx context.Context, userID uuid.UUID, status string) ([]models.Auction, error)
	ListMyBids(ctx context.Context, userID uuid.UUID) ([]models.Auction, error)
	ListMyBidsByStatus(ctx context.Context, userID uuid.UUID, status string) ([]models.Auction, error)
	ListMyWinnings(ctx context.Context, userID uuid.UUID) ([]models.Auction, error)

	// KYC
	SubmitKYC(ctx context.Context, kyc *models.KYCVerification) error
	GetKYCStatus(ctx context.Context, userID uuid.UUID) (*models.KYCVerification, error)
	UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status string) error

	// User Settings (new)
	GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error)
	UpdateUserSettings(ctx context.Context, userID uuid.UUID, settings interface{}) error
	Search(ctx context.Context, query string, page, perPage int) ([]models.User, int, error)
}

type userService struct {
	repo         repository.UserRepository
	favoriteRepo repository.FavoriteRepository
	auctionRepo  repository.AuctionRepository
	kycRepo      repository.KYCRepository
	auditSvc     AuditService
	rdb          *redis.Client
	logger       *zap.Logger
	jwtExpiry    int
}

func NewUserService(
	repo repository.UserRepository,
	favoriteRepo repository.FavoriteRepository,
	auctionRepo repository.AuctionRepository,
	kycRepo repository.KYCRepository,
	auditSvc AuditService,
	rdb *redis.Client,
	logger *zap.Logger,
	jwtExpiry int,
) UserService {
	return &userService{
		repo:         repo,
		favoriteRepo: favoriteRepo,
		auctionRepo:  auctionRepo,
		kycRepo:      kycRepo,
		auditSvc:     auditSvc,
		rdb:          rdb,
		logger:       logger,
		jwtExpiry:    jwtExpiry,
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

// DeactivateOwnAccount supprime le compte de l'utilisateur authentifié
// lui-même — jamais un autre utilisateur, userID vient exclusivement du JWT côté
// handler (Release Phase 1B — App Store account deletion, guideline 5.1.1(v),
// qui exige une suppression réelle et pas seulement une désactivation). Un hard
// delete de la ligne users (comme AdminService.DeleteUser / userRepo.Delete)
// reste délibérément évité : auctions/bids/wallet/transactions référencent
// users(id) par FK, et les retirer casserait ces relations ou nécessiterait des
// CASCADE dangereux sur des données financières/légales à conserver. À la
// place : is_active=false + anonymisation des champs personnels directs
// (userRepo.AnonymizeAndDeactivate — voir son commentaire pour le détail exact
// des colonnes effacées/remplacées). Révoque aussi toutes les sessions JWT
// actives (même mécanisme que BlockUser, Session Security Phase 1).
func (s *userService) DeactivateOwnAccount(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.AnonymizeAndDeactivate(ctx, userID); err != nil {
		return err
	}

	// Un compte supprimé ne doit pas rester utilisable via un JWT déjà émis.
	RevokeUserSessions(ctx, s.rdb, s.logger, userID, s.jwtExpiry)

	if s.auditSvc != nil {
		// Jamais l'ancien phone/email/token — uniquement l'identifiant de
		// l'utilisateur et des indicateurs non sensibles (Release Phase 1B).
		if auditErr := s.auditSvc.Log(ctx, userID, "user_account_deleted", "user", &userID,
			"Account deletion requested by owner (anonymized, financial/legal records retained)",
			WithActorType("user"),
			WithDetailsJSON(models.JSONB{
				"target_user_id":   userID.String(),
				"anonymized":       true,
				"retained_records": []string{"transactions", "auctions", "bids"},
			}),
		); auditErr != nil {
			if s.logger != nil {
				s.logger.Error("DeactivateOwnAccount: failed to write audit log", zap.String("user_id", userID.String()), zap.Error(auditErr))
			}
		}
	}

	return nil
}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// AddFavorite enforces country-scoped market isolation (migration 000046, V1): a user
// may only favorite an auction in their own effective market -- otherwise a cross-market
// auction ID could be added to favorites and surfaced back to the user despite the list/
// detail endpoints excluding it. 404 (via apperr.ErrNotFound), not a "wrong market"
// disclosure, matching auction_handler.go GetByID's convention for inaccessible auctions.
func (s *userService) AddFavorite(ctx context.Context, userID, auctionID uuid.UUID) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	auction, err := s.auctionRepo.FindByID(ctx, auctionID)
	if err != nil {
		return apperr.ErrNotFound
	}
	if user.EffectiveAccountCountryISO() != auction.EffectiveMarketCountryISO() {
		return apperr.ErrNotFound
	}
	return s.favoriteRepo.Add(ctx, userID, auctionID)
}

// RemoveFavorite is intentionally NOT market-scoped: a user must always be able to
// remove their own existing favorite relation (e.g. cleanup of a stale row from before
// this check existed), and removal reveals no auction details -- only that a
// (user_id, auction_id) pair no longer exists.
func (s *userService) RemoveFavorite(ctx context.Context, userID, auctionID uuid.UUID) error {
	return s.favoriteRepo.Remove(ctx, userID, auctionID)
}

// ListFavorites filters out any favorited auction outside the caller's effective market
// (migration 000046, V1) -- defense-in-depth against stale rows predating the AddFavorite
// guard above, or any other path that could otherwise have created a cross-market
// favorite. Never discloses the excluded auction's existence, just omits it.
func (s *userService) ListFavorites(ctx context.Context, userID uuid.UUID) ([]models.Auction, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	auctions, err := s.favoriteRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	market := user.EffectiveAccountCountryISO()
	filtered := make([]models.Auction, 0, len(auctions))
	for _, a := range auctions {
		if a.EffectiveMarketCountryISO() == market {
			filtered = append(filtered, a)
		}
	}
	return filtered, nil
}

func (s *userService) ListMyAuctions(ctx context.Context, userID uuid.UUID) ([]models.Auction, error) {
	filters := repository.AuctionFilters{SellerID: &userID}
	auctions, _, err := s.auctionRepo.ListPaginated(ctx, 1, 100, filters)
	return auctions, err
}

func (s *userService) ListMyAuctionsByStatus(ctx context.Context, userID uuid.UUID, status string) ([]models.Auction, error) {
	filters := repository.AuctionFilters{SellerID: &userID, Status: status}
	auctions, _, err := s.auctionRepo.ListPaginated(ctx, 1, 100, filters)
	return auctions, err
}

func (s *userService) ListMyBids(ctx context.Context, userID uuid.UUID) ([]models.Auction, error) {
	// Logic to find auctions where user has bid
	return s.auctionRepo.ListByUserBids(ctx, userID)
}

func (s *userService) ListMyBidsByStatus(ctx context.Context, userID uuid.UUID, status string) ([]models.Auction, error) {
	// Logic to find auctions where user has bid by status
	// For now, return all bids and filter in application layer
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

// New methods for extended user functionality

func (s *userService) UpdateProfileExtended(ctx context.Context, id uuid.UUID, fullName, email, city, countryCode, address, postalCode, dateOfBirth, gender string) error {
	return s.repo.UpdateProfileExtended(ctx, id, fullName, email, city, countryCode, address, postalCode, dateOfBirth, gender)
}

func (s *userService) UpdateKYCStatus(ctx context.Context, userID uuid.UUID, status string) error {
	return s.repo.UpdateKYCStatus(ctx, userID, status)
}

func (s *userService) GetUserSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	return s.repo.GetUserSettings(ctx, userID)
}

func (s *userService) UpdateUserSettings(ctx context.Context, userID uuid.UUID, settings interface{}) error {
	return s.repo.UpdateUserSettings(ctx, userID, settings)
}

func (s *userService) Search(ctx context.Context, query string, page, perPage int) ([]models.User, int, error) {
	return s.repo.ListPaginated(ctx, page, perPage, query)
}

