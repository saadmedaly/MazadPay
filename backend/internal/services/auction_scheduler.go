package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AuctionScheduler runs background jobs for auction-related notifications
type AuctionScheduler struct {
	db                *sqlx.DB
	auctionRepo       repository.AuctionRepository
	bidRepo           repository.BidRepository
	notificationSvc   NotificationService
	userRepo          repository.UserRepository
	redis             *redis.Client
	logger            *zap.Logger
	stopChan          chan struct{}
}

// NewAuctionScheduler creates a new auction scheduler
func NewAuctionScheduler(
	db *sqlx.DB,
	auctionRepo repository.AuctionRepository,
	bidRepo repository.BidRepository,
	notificationSvc NotificationService,
	userRepo repository.UserRepository,
	redis *redis.Client,
	logger *zap.Logger,
) *AuctionScheduler {
	return &AuctionScheduler{
		db:              db,
		auctionRepo:     auctionRepo,
		bidRepo:         bidRepo,
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
		
		// Item 11 fix (client feedback Phase B): the scheduler used to only send
		// notifications and never persisted a winner -- auctions.winner_id stayed
		// NULL forever for every real bidding win (SetWinner was defined in the
		// repo but never called from anywhere), which meant "My Winnings"
		// (ListMyWinnings -> WHERE winner_id = $userID) was permanently empty for
		// legitimate winners even though the push notification fired correctly.
		// Also: without persisting status='ended', FindEndedSince kept matching
		// the same 'active' auctions on every 1-minute tick until they aged past
		// its lookback window, relying solely on the Redis dedup key above to
		// avoid renotifying -- a second latent issue this fix also closes.
		topBid, bidErr := s.bidRepo.FindTopBid(ctx, auction.ID)
		hasWinner := bidErr == nil && topBid != nil && auction.CurrentPrice.GreaterThan(auction.StartPrice) && auction.BidderCount > 0

		if hasWinner {
			if err := s.setAuctionWinner(ctx, auction.ID, topBid.UserID, topBid.ID); err != nil {
				s.logger.Error("Failed to persist auction winner", zap.Error(err), zap.String("auction_id", auction.ID.String()))
			}

			winner, err := s.userRepo.FindByID(ctx, topBid.UserID)
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

				err = s.notificationSvc.SendLocalizedPush(ctx, topBid.UserID, "auction_won", winnerLang, winnerParams, winnerData)
				if err != nil {
					s.logger.Error("Failed to send winner notification", zap.Error(err))
				}
			}
		} else {
			// No qualifying bid: still transition out of 'active' so
			// FindEndedSince stops matching this auction on every tick.
			if err := s.auctionRepo.UpdateStatus(ctx, auction.ID, "ended"); err != nil {
				s.logger.Error("Failed to mark unsold auction as ended", zap.Error(err), zap.String("auction_id", auction.ID.String()))
			}
		}
	}
}

// setAuctionWinner runs SetWinner (winner_id/winning_bid_id/status='ended') in
// its own short transaction -- SetWinner requires a *sqlx.Tx (see auction_repo.go),
// and the scheduler has no other transactional caller, so it opens/commits one here.
func (s *AuctionScheduler) setAuctionWinner(ctx context.Context, auctionID, winnerID, winningBidID uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.auctionRepo.SetWinner(ctx, tx, auctionID, winnerID, winningBidID); err != nil {
		return err
	}

	return tx.Commit()
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
