package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AuctionScheduler runs background jobs for auction-related notifications
type AuctionScheduler struct {
	auctionRepo       repository.AuctionRepository
	notificationSvc   NotificationService
	userRepo          repository.UserRepository
	redis             *redis.Client
	logger            *zap.Logger
	stopChan          chan struct{}
}

// NewAuctionScheduler creates a new auction scheduler
func NewAuctionScheduler(
	auctionRepo repository.AuctionRepository,
	notificationSvc NotificationService,
	userRepo repository.UserRepository,
	redis *redis.Client,
	logger *zap.Logger,
) *AuctionScheduler {
	return &AuctionScheduler{
		auctionRepo:     auctionRepo,
		notificationSvc: notificationSvc,
		userRepo:        userRepo,
		redis:           redis,
		logger:          logger,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the scheduler loop
func (s *AuctionScheduler) Start() {
	s.logger.Info("Starting auction scheduler")
	
	// Run immediately, then every minute
	s.runChecks()
	
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.runChecks()
			case <-s.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the scheduler
func (s *AuctionScheduler) Stop() {
	close(s.stopChan)
	s.logger.Info("Auction scheduler stopped")
}

// runChecks runs all scheduled checks
func (s *AuctionScheduler) runChecks() {
	ctx := context.Background()
	
	// Check for just-ended auctions
	s.checkEndedAuctions(ctx)
}

// checkEndedAuctions checks for auctions that just ended and sends notifications
func (s *AuctionScheduler) checkEndedAuctions(ctx context.Context) {
	// Find auctions that ended in the last minute
	endedTime := time.Now().Add(-1 * time.Minute)
	auctions, err := s.auctionRepo.FindEndedSince(ctx, endedTime)
	if err != nil {
		s.logger.Error("Failed to find ended auctions", zap.Error(err))
		return
	}
	
	for _, auction := range auctions {
		if s.hasNotificationBeenSent(ctx, auction.ID, "ended") {
			continue
		}
		
		seller, err := s.userRepo.FindByID(ctx, auction.SellerID)
		if err != nil {
			s.logger.Warn("Failed to get seller info", zap.Error(err))
			continue
		}
		
		language := "ar"
		if seller.LanguagePref != "" {
			language = seller.LanguagePref
		}
		
		params := map[string]string{
			"auctionTitle": auction.TitleAr,
			"finalPrice":  auction.CurrentPrice.String(),
			"currency":    auction.EffectiveCurrencyCode(),
		}
		data := map[string]string{
			"type":      "auction_ended",
			"auctionId": auction.ID.String(),
			"finalPrice": auction.CurrentPrice.String(),
		}
		
		// Notify seller
		err = s.notificationSvc.SendLocalizedPush(ctx, auction.SellerID, "auction_ended", language, params, data)
		if err != nil {
			s.logger.Error("Failed to send ended notification to seller", zap.Error(err))
		}
		
		// Notify winner if there is one
		if auction.CurrentPrice.GreaterThan(auction.StartPrice) && auction.BidderCount > 0 {
			// Get the highest bidder
			winnerID, err := s.auctionRepo.GetHighestBidder(ctx, auction.ID)
			if err == nil && winnerID != uuid.Nil {
				winner, err := s.userRepo.FindByID(ctx, winnerID)
				if err == nil {
					winnerLang := "ar"
					if winner.LanguagePref != "" {
						winnerLang = winner.LanguagePref
					}
					
					winnerParams := map[string]string{
						"auctionTitle": auction.TitleAr,
						"finalPrice":  auction.CurrentPrice.String(),
						"currency":    auction.EffectiveCurrencyCode(),
					}
					winnerData := map[string]string{
						"type":      "auction_won",
						"auctionId": auction.ID.String(),
						"finalPrice": auction.CurrentPrice.String(),
					}
					
					err = s.notificationSvc.SendLocalizedPush(ctx, winnerID, "auction_won", winnerLang, winnerParams, winnerData)
					if err != nil {
						s.logger.Error("Failed to send winner notification", zap.Error(err))
					}
				}
			}
		}
	}
}

// hasNotificationBeenSent checks if a notification has already been sent for this auction/event
// Uses Redis SETNX with 24h TTL for deduplication
func (s *AuctionScheduler) hasNotificationBeenSent(ctx context.Context, auctionID uuid.UUID, eventType string) bool {
	if s.redis == nil {
		return false
	}

	key := fmt.Sprintf("notif:%s:%s", auctionID.String(), eventType)

	// Try to set the key with NX (only if not exists) and 24h TTL
	set, err := s.redis.SetNX(ctx, key, "1", 24*time.Hour).Result()
	if err != nil {
		s.logger.Warn("Failed to check notification dedup in Redis",
			zap.Error(err),
			zap.String("key", key),
		)
		return false
	}

	// If set is false, key already exists (notification was already sent)
	return !set
}
