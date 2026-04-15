package services

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    apperr "github.com/mazadpay/backend/internal/errors"
    "github.com/mazadpay/backend/internal/models"
    "github.com/mazadpay/backend/internal/repository"
    ws "github.com/mazadpay/backend/internal/websocket"
    "github.com/shopspring/decimal"
)

type BidService interface {
    PlaceBid(ctx context.Context, auctionID, userID uuid.UUID, amount decimal.Decimal) (*models.Bid, error)
    GetHistory(ctx context.Context, auctionID uuid.UUID) ([]models.Bid, error)
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

    err := repository.WithTransaction(s.db, func(tx *sqlx.Tx) error {
        // 1. Charger l'enchère (sans FOR UPDATE car on utilise version)
        auction, err := s.auctionRepo.FindByID(ctx, auctionID)
        if err != nil {
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

        // 3. Vérifier et geler le solde (SELECT FOR UPDATE sur wallet)
        wallet, err := s.walletRepo.FindForUpdate(ctx, tx, userID)
        if err != nil {
            return apperr.ErrNotFound
        }

        // Vérifier si l'utilisateur a déjà un hold pour cette enchère
        existingHold, _ := s.walletRepo.FindActiveHold(ctx, tx, userID, auctionID)

        if existingHold == nil {
            // Première mise : vérifier que le solde couvre la caution
            if wallet.Balance.LessThan(auction.InsuranceAmount) {
                return apperr.ErrInsufficientBalance
            }
            // Geler la caution
            if err := s.walletRepo.CreateHold(ctx, tx, userID, auctionID, auction.InsuranceAmount); err != nil {
                return err
            }
            if err := s.walletRepo.DebitFreezeBalance(ctx, tx, userID, auction.InsuranceAmount, wallet.Version); err != nil {
                return err
            }
        }
        // Si mise précédente existante → le hold est déjà actif, pas de nouveau débit

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

        createdBid = bid
        return nil
    })

    if err != nil {
        return nil, err
    }

    // 7. Broadcast WebSocket APRÈS le commit (critique : jamais avant)
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

func (s *bidService) GetHistory(ctx context.Context, auctionID uuid.UUID) ([]models.Bid, error) {
    bids, err := s.bidRepo.FindByAuction(ctx, auctionID)
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

