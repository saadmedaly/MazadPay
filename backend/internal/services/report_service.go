package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/mazadpay/backend/internal/repository"
)

type ReportService interface {
	ExportTransactionsCSV(ctx context.Context, status string, startDate, endDate *time.Time) (string, error)
	ExportRevenueCSV(ctx context.Context) (string, error)
}

type reportServiceV2 struct {
	txRepo repository.TransactionRepository
}

func NewReportService(txRepo repository.TransactionRepository) ReportService {
	return &reportServiceV2{txRepo: txRepo}
}

func (s *reportServiceV2) ExportTransactionsCSV(ctx context.Context, status string, startDate, endDate *time.Time) (string, error) {
	// For now, we list all transactions. 
	// We might need a more specialized query in the repo for date ranges, 
	// but let's reuse ListPaginated with a large perPage for the MVP.
	txs, _, err := s.txRepo.ListPaginated(ctx, 1, 10000, status, nil)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// Header
	// Currency (migration 000046, Phase 2): each row carries its own currency_code
	// (EffectiveCurrencyCode fallback for legacy NULL rows) -- rows can legitimately
	// span multiple currencies (MRU/TND/MAD/XOF/...), so the CSV never sums Amount
	// across rows into one fake blended total; a reader must group by Currency
	// themselves if they need per-currency subtotals.
	writer.Write([]string{"ID", "User ID", "Auction ID", "Type", "Amount", "Currency", "Gateway", "Status", "Reference", "Date"})

	for _, tx := range txs {
		// Filter by date if provided (ideally done in SQL)
		if startDate != nil && tx.CreatedAt.Before(*startDate) {
			continue
		}
		if endDate != nil && tx.CreatedAt.After(*endDate) {
			continue
		}

		gateway := ""
		if tx.Gateway != nil {
			gateway = *tx.Gateway
		}

		reference := ""
		if tx.Reference != nil {
			reference = *tx.Reference
		}

		row := []string{
			tx.ID.String(),
			tx.UserID.String(),
			func() string {
				if tx.AuctionID != nil {
					return tx.AuctionID.String()
				}
				return "N/A"
			}(),
			tx.Type,
			tx.Amount.String(),
			tx.EffectiveCurrencyCode(),
			gateway,
			tx.Status,
			reference,
			tx.CreatedAt.Format(time.RFC3339),
		}
		writer.Write(row)
	}

	writer.Flush()
	return sb.String(), nil
}

func (s *reportServiceV2) ExportRevenueCSV(ctx context.Context) (string, error) {
	// Currency-grouped (migration 000046, Phase 2): revenue rows can legitimately
	// span multiple currencies, so this export lists a separate row per
	// (date, currency) instead of a single "Revenue (MRU)" column that would
	// silently blend e.g. MRU and TND amounts into one fake figure.
	data, err := s.txRepo.GetDailyRevenueChartByCurrency(ctx)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	writer.Write([]string{"Date", "Currency", "Revenue"})

	for _, d := range data {
		writer.Write([]string{
			fmt.Sprintf("%v", d["date"]),
			fmt.Sprintf("%v", d["currency"]),
			fmt.Sprintf("%v", d["amount"]),
		})
	}

	writer.Flush()
	return sb.String(), nil
}
