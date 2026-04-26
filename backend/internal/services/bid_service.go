package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/database"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
	ws "github.com/mazadpay/backend/internal/websocket"
	"github.com/shopspring/decimal"
)

type BidService interface {
	PlaceBid(ctx context.Context, auctionID, userID uuid.UUID, amount decimal.Decimal) (*models.Bid, error)
	GetHistory(ctx context.Context, auctionID uuid.UUID) ([]models.BidHistoryEntry, error)
}

type bidService struct {
	db          *sqlx.DB
	auctionRepo repository.AuctionRepository
	bidRepo     repository.BidRepository
	walletRepo  repository.WalletRepository
	hub         *ws.Hub
}

func NewBidService(
	db *sqlx.DB,
	auctionRepo repository.AuctionRepository,
	bidRepo repository.BidRepository,
	walletRepo repository.WalletRepository,
	hub *ws.Hub,
) BidService {
	return &bidService{db: db, auctionRepo: auctionRepo, bidRepo: bidRepo, walletRepo: walletRepo, hub: hub}
}

// PlaceBid — LOGIQUE CRITIQUE avec verrouillage optimiste
// Ordre : SELECT FOR UPDATE wallet → vérifications → UPDATE auctions (version) → INSERT bid → COMMIT → Broadcast
func (s *bidService) PlaceBid(ctx context.Context, auctionID, userID uuid.UUID, amount decimal.Decimal) (*models.Bid, error) {
	var createdBid *models.Bid

	err := database.WithTransaction(s.db, func(tx *sqlx.Tx) error {
		// 1. Charger l'enchère avec la transaction pour garantir la cohérence
		auction, err := s.auctionRepo.FindByIDTx(ctx, tx, auctionID)
		if err != nil {
			fmt.Printf("FindByIDTx error for auction %s: %v\n", auctionID, err)
			return apperr.ErrNotFound
		}

		// 2. Vérifications métier
		if auction.Status != "active" {
			return apperr.ErrAuctionNotActive
		}
		if time.Now().After(auction.EndTime) {
			return apperr.ErrAuctionEnded
		}
		if auction.SellerID == userID {
			return apperr.ErrSelfBid
		}

		minRequired := auction.CurrentPrice.Add(auction.MinIncrement)
		if amount.LessThan(minRequired) {
			return apperr.ErrBidTooLow
		}

		// 3. [SUPPRIMÉ] Pas de vérification wallet/caution lors du bid
		// Les utilisateurs peuvent enchérir librement sans bloquer de fonds
		// Le paiement se fera après la fin de l'enchère uniquement pour le gagnant

		// 4. Marquer les anciens bids comme non-gagnants
		if err := s.bidRepo.SetAllNotWinning(ctx, tx, auctionID); err != nil {
			return err
		}

		// 5. Mettre à jour le prix de l'enchère avec verrouillage optimiste
		ok, err := s.auctionRepo.UpdatePrice(ctx, tx, auctionID, amount, auction.Version)
		if err != nil {
			return err
		}
		if !ok {
			return apperr.ErrBidConflict // Le client doit retry
		}

		// 6. Insérer le nouveau bid
		bid := &models.Bid{
			ID:            uuid.New(),
			AuctionID:     auctionID,
			UserID:        userID,
			Amount:        amount,
			PreviousPrice: &auction.CurrentPrice,
			IsWinning:     true,
		}
		if err := s.bidRepo.Create(ctx, tx, bid); err != nil {
			return err
		}

		// 7. Broadcast WebSocket en temps réel
		go func() {
			// Récupérer les infos utilisateur pour le masquage
			var userPhone string
			err := s.db.QueryRowContext(ctx, "SELECT phone FROM users WHERE id = $1", userID).Scan(&userPhone)
			if err != nil {
				return
			}
			// Masquer le numéro (garder 4 derniers chiffres)
			if len(userPhone) >= 4 {
				userPhone = "####" + userPhone[len(userPhone)-4:]
			}

 			payload := ws.WSEvent{
				Type: "bid_placed",
				Payload: ws.BidPlacedPayload{
					AuctionID:    auctionID.String(),
					NewPrice:     amount.InexactFloat64(),
					BidderMasked: userPhone,
					BidCount:     0, // À calculer depuis la base
					SecondsLeft:  int64(auction.EndTime.Sub(time.Now()).Seconds()),
				},
			}

			// Envoyer à tous les clients de l'enchère
			s.hub.Broadcast(auctionID, payload)
		}()

		createdBid = bid
		return nil
	})

	if err != nil {
		return nil, err
	}

 	auction, _ := s.auctionRepo.FindByID(ctx, auctionID)
	secsLeft := int64(0)
	if auction != nil {
		secsLeft = int64(time.Until(auction.EndTime).Seconds())
	}

	s.hub.Broadcast(auctionID, ws.WSEvent{
		Type: "bid_placed",
		Payload: ws.BidPlacedPayload{
			AuctionID:    auctionID.String(),
			NewPrice:     amount.InexactFloat64(),
			BidderMasked: "####" + userID.String()[len(userID.String())-4:],
			SecondsLeft:  secsLeft,
		},
	})

	return createdBid, nil
}

func (s *bidService) GetHistory(ctx context.Context, auctionID uuid.UUID) ([]models.BidHistoryEntry, error) {
	bids, err := s.bidRepo.FindHistoryByAuction(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	// Masquage des numéros de téléphone (####xxxx)
	for i := range bids {
		if bids[i].BidderPhone != "" {
			p := bids[i].BidderPhone
			if len(p) > 4 {
				bids[i].BidderPhone = "####" + p[len(p)-4:]
			} else {
				bids[i].BidderPhone = "####"
			}
		}
	}

	return bids, nil
}
