package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mazadpay/backend/internal/database"
	apperr "github.com/mazadpay/backend/internal/errors"
	"github.com/mazadpay/backend/internal/models"
	"github.com/mazadpay/backend/internal/repository"
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
	userRepo    repository.UserRepository
	notifSvc    NotificationService
	hub         AuctionHub
}

func NewBidService(
	db *sqlx.DB,
	auctionRepo repository.AuctionRepository,
	bidRepo repository.BidRepository,
	walletRepo repository.WalletRepository,
	userRepo repository.UserRepository,
	notifSvc NotificationService,
	hub AuctionHub,
) BidService {
	return &bidService{db: db, auctionRepo: auctionRepo, bidRepo: bidRepo, walletRepo: walletRepo, userRepo: userRepo, notifSvc: notifSvc, hub: hub}
}

// PlaceBid — LOGIQUE CRITIQUE avec verrouillage optimiste
// Ordre : SELECT FOR UPDATE wallet → vérifications → UPDATE auctions (version) → INSERT bid → COMMIT → Broadcast
func (s *bidService) PlaceBid(ctx context.Context, auctionID, userID uuid.UUID, amount decimal.Decimal) (*models.Bid, error) {
	var createdBid *models.Bid

	// Trouver l'ancien meilleur enchérisseur avant la transaction
	prevTopBid, _ := s.bidRepo.FindTopBid(ctx, auctionID)

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

		// 2b. Country-scoped market check (migration 000046, V1) — the bidder and
		// the auction MUST belong to the same market (account_country_iso ==
		// market_country_iso). Checked by COUNTRY, never by currency alone: two
		// countries can share a currency (e.g. SN/CI both use XOF) but must
		// remain separate markets — see errors.ErrCrossMarketBid.
		bidder, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			return apperr.ErrNotFound
		}
		if bidder.EffectiveAccountCountryISO() != auction.EffectiveMarketCountryISO() {
			return apperr.ErrCrossMarketBid
		}

		// 3. Vérification et gel de la caution (insurance_amount) — audit de sécurité V03
		// (durci suite à un contournement constaté en production : tous les auctions actifs
		// avaient insurance_amount = 0, ce qui désactivait complètement la protection).
		//
		// Règle stricte inchangée pour toute enchère "required" (politique par défaut,
		// voir Auction.InsuranceRequired -- défaut à true pour toute valeur vide/legacy) :
		// aucune mise sans caution définie. Si insurance_amount <= 0, l'enchère est
		// refusée. Il n'y a toujours aucune "compatibilité silencieuse" avec une caution
		// nulle par accident ou absence de configuration.
		//
		// Insurance policy (migration 000048) : SEULE une politique "not_required"
		// explicitement choisie par un admin (jamais un défaut, jamais un état
		// accidentel -- voir AdminUpdateAuctionRequest/ReviewAuctionRequest) permet de
		// sauter entièrement ce bloc. Dans ce cas, la mise est acceptée sans aucune
		// caution gelée : la protection financière V03 (avoir "quelque chose à perdre")
		// est délibérément absente pour cette enchère précise, par décision explicite de
		// l'admin -- ce n'est PAS le même invariant de sécurité qu'une enchère "required",
		// et ne doit jamais être présenté comme équivalent.
		if auction.InsuranceRequired() {
			if !auction.InsuranceAmount.GreaterThan(decimal.Zero) {
				return apperr.ErrInsuranceNotSet
			}

			existingHold, err := s.walletRepo.FindActiveHold(ctx, tx, userID, auctionID)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
			if existingHold == nil {
				wallet, err := s.walletRepo.FindForUpdate(ctx, tx, userID)
				if err != nil {
					return err
				}
				// Defense-in-depth (migration 000046): the wallet's currency must
				// match the auction's currency before any freeze/comparison. Should
				// never trigger in practice given the ErrCrossMarketBid check above
				// (a wallet is always denominated in its owner's account market's
				// currency), but verified explicitly per financial-safety
				// requirements rather than assumed.
				if wallet.EffectiveCurrencyCode() != auction.EffectiveCurrencyCode() {
					return apperr.ErrWalletCurrencyMismatch
				}
				if wallet.Balance.LessThan(auction.InsuranceAmount) {
					return apperr.ErrInsufficientForInsurance
				}
				if err := s.walletRepo.DebitFreezeBalance(ctx, tx, userID, auction.InsuranceAmount, wallet.Version); err != nil {
					return err
				}
				if err := s.walletRepo.CreateHold(ctx, tx, userID, auctionID, auction.InsuranceAmount); err != nil {
					return err
				}
			}
		}

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
			var userPhone string
			err := s.db.QueryRowContext(ctx, "SELECT phone FROM users WHERE id = $1", userID).Scan(&userPhone)
			if err != nil {
				return
			}
			if len(userPhone) >= 4 {
				userPhone = "####" + userPhone[len(userPhone)-4:]
			}

 			payload := models.WSEvent{
				Type: "bid_placed",
				Payload: models.BidPlacedPayload{
					AuctionID:    auctionID.String(),
					NewPrice:     amount.InexactFloat64(),
					BidderMasked: userPhone,
					BidCount:     0,
					SecondsLeft:  int64(auction.EndTime.Sub(time.Now()).Seconds()),
				},
			}

			s.hub.Broadcast(auctionID, payload)
		}()

		createdBid = bid
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 8. Notifier l'ancien enchérisseur qu'il a été dépassé
	go func() {
		if prevTopBid == nil || prevTopBid.UserID == userID {
			return
		}
		auction, err := s.auctionRepo.FindByID(ctx, auctionID)
		if err != nil {
			return
		}
		previousBidder, err := s.userRepo.FindByID(ctx, prevTopBid.UserID)
		if err != nil {
			return
		}
		language := "ar"
		if previousBidder.LanguagePref != "" {
			language = previousBidder.LanguagePref
		}
		title := auction.TitleAr
		if language == "fr" && auction.TitleFr != nil && *auction.TitleFr != "" {
			title = *auction.TitleFr
		} else if language == "en" && auction.TitleEn != nil && *auction.TitleEn != "" {
			title = *auction.TitleEn
		}
		_ = s.notifSvc.SendLocalizedPush(ctx, prevTopBid.UserID, "bid_outbid", language, map[string]string{
			"auctionTitle": title,
			"newPrice":     amount.String(),
			"currency":     auction.EffectiveCurrencyCode(),
		}, map[string]string{
			"type":      "bid_outbid",
			"auctionId": auctionID.String(),
		})
	}()

 	auction, _ := s.auctionRepo.FindByID(ctx, auctionID)
	secsLeft := int64(0)
	if auction != nil {
		secsLeft = int64(time.Until(auction.EndTime).Seconds())
	}

	s.hub.Broadcast(auctionID, models.WSEvent{
		Type: "bid_placed",
		Payload: models.BidPlacedPayload{
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
