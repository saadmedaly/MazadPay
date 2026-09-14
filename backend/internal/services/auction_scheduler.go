package services

import (
	"context"
	"time"

	"github.com/mazadpay/backend/internal/repository"
	"go.uber.org/zap"
)

// AuctionScheduler runs background jobs for auction-related notifications.
// Customer #23: no longer owns any winner-selection/persistence logic or
// Redis dedup itself -- see checkEndedAuctions's doc comment. It only
// discovers candidate ended auctions and delegates to the canonical
// AuctionService.FinalizeExpiredAuction.
type AuctionScheduler struct {
	auctionRepo repository.AuctionRepository
	auctionSvc  AuctionService
	logger      *zap.Logger
	stopChan    chan struct{}
}

// NewAuctionScheduler creates a new auction scheduler
func NewAuctionScheduler(
	auctionRepo repository.AuctionRepository,
	auctionSvc AuctionService,
	logger *zap.Logger,
) *AuctionScheduler {
	return &AuctionScheduler{
		auctionRepo: auctionRepo,
		auctionSvc:  auctionSvc,
		logger:      logger,
		stopChan:    make(chan struct{}),
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

// checkEndedAuctions checks for auctions that just ended and delegates their
// actual finalization to the canonical AuctionService.FinalizeExpiredAuction.
//
// Customer #23: this used to contain its OWN winner-selection/persistence
// logic, entirely separate from (and racing against) auction_service.go's
// CloseExpiredAuctions, which ran on a faster 30-second tick and never
// called SetWinner at all -- it would almost always flip status to 'ended'
// first, which permanently excluded the auction from this scheduler's own
// status='active' query (FindEndedSince), silently starving the only path
// that persisted winner_id. Both background loops now discover candidate
// auction IDs on their own schedules but delegate the actual decision
// (winner selection, persistence, notifications, realtime event) to the
// same single canonical operation -- see FinalizeExpiredAuction's own doc
// comment for why concurrent calls for the same auction are safe (exactly
// one wins the DB-level race, the other no-ops).
func (s *AuctionScheduler) checkEndedAuctions(ctx context.Context) {
	// Find auctions that ended in the last minute
	endedTime := time.Now().Add(-1 * time.Minute)
	auctions, err := s.auctionRepo.FindEndedSince(ctx, endedTime)
	if err != nil {
		s.logger.Error("Failed to find ended auctions", zap.Error(err))
		return
	}

	for _, auction := range auctions {
		if err := s.auctionSvc.FinalizeExpiredAuction(ctx, auction.ID); err != nil {
			s.logger.Error("checkEndedAuctions: FinalizeExpiredAuction failed", zap.Error(err), zap.String("auction_id", auction.ID.String()))
		}
	}
}
