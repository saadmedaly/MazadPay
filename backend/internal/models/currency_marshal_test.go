package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// These tests cover Phase 2 (migration 000046, country-scoped currency) API
// contract exposure: any handler that serializes Auction/Wallet/Transaction/
// AuctionRequest directly (not via a hand-built fiber.Map) must still emit
// currency_code/market_country_iso, falling back to Default*  for legacy rows
// where the raw pointer field is nil. Without the MarshalJSON overrides in
// auction.go/wallet.go/request.go, the `omitempty` tag on the raw pointer
// field would silently drop the key for legacy NULL rows instead of falling
// back -- these tests would fail on that regression.

func TestAuction_MarshalJSON_LegacyFallback(t *testing.T) {
	a := Auction{
		ID:       uuid.New(),
		SellerID: uuid.New(),
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Hour),
		// CurrencyCode/MarketCountryISO left nil: simulates a pre-migration-000046 row.
	}
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if out["currency_code"] != DefaultCurrencyCode {
		t.Fatalf("expected currency_code=%s fallback for legacy nil row, got %v", DefaultCurrencyCode, out["currency_code"])
	}
	if out["market_country_iso"] != DefaultAccountCountryISO {
		t.Fatalf("expected market_country_iso=%s fallback for legacy nil row, got %v", DefaultAccountCountryISO, out["market_country_iso"])
	}
}

func TestAuction_MarshalJSON_NonLegacyPreserved(t *testing.T) {
	currency := "TND"
	market := "TN"
	a := Auction{
		ID:               uuid.New(),
		SellerID:         uuid.New(),
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(time.Hour),
		CurrencyCode:     &currency,
		MarketCountryISO: &market,
	}
	b, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if out["currency_code"] != "TND" {
		t.Fatalf("expected currency_code=TND to be preserved, got %v", out["currency_code"])
	}
	if out["market_country_iso"] != "TN" {
		t.Fatalf("expected market_country_iso=TN to be preserved, got %v", out["market_country_iso"])
	}
}

func TestWallet_MarshalJSON_LegacyFallback(t *testing.T) {
	w := Wallet{
		UserID:       uuid.New(),
		Balance:      decimal.NewFromInt(100),
		FrozenAmount: decimal.Zero,
		UpdatedAt:    time.Now(),
		// CurrencyCode nil: pre-migration wallet.
	}
	b, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if out["currency_code"] != DefaultCurrencyCode {
		t.Fatalf("expected currency_code=%s fallback for legacy nil wallet, got %v", DefaultCurrencyCode, out["currency_code"])
	}
}

func TestTransaction_MarshalJSON_LegacyFallback(t *testing.T) {
	tx := Transaction{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Type:      "deposit",
		Amount:    decimal.NewFromInt(50),
		Status:    "approved",
		CreatedAt: time.Now(),
		// CurrencyCode nil: pre-migration transaction.
	}
	b, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if out["currency_code"] != DefaultCurrencyCode {
		t.Fatalf("expected currency_code=%s fallback for legacy nil transaction, got %v", DefaultCurrencyCode, out["currency_code"])
	}
}

func TestAuctionRequest_MarshalJSON_LegacyFallback(t *testing.T) {
	r := AuctionRequest{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		CategoryID:   1,
		TitleAr:      "test",
		StartPrice:   decimal.NewFromInt(10),
		MinIncrement: decimal.NewFromInt(1),
		StartDate:    time.Now(),
		EndDate:      time.Now().Add(time.Hour),
		// CurrencyCode/MarketCountryISO nil: pre-migration request.
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if out["currency_code"] != DefaultCurrencyCode {
		t.Fatalf("expected currency_code=%s fallback for legacy nil request, got %v", DefaultCurrencyCode, out["currency_code"])
	}
	if out["market_country_iso"] != DefaultAccountCountryISO {
		t.Fatalf("expected market_country_iso=%s fallback for legacy nil request, got %v", DefaultAccountCountryISO, out["market_country_iso"])
	}
}

// Sanity: MRU must be the actual fallback ISO code used across the system --
// never the stale/unsupported CLDR "MRO".
func TestDefaultCurrencyCode_IsMRU_NotStaleMRO(t *testing.T) {
	if DefaultCurrencyCode != "MRU" {
		t.Fatalf("expected DefaultCurrencyCode=MRU, got %s (stale MRO code regression?)", DefaultCurrencyCode)
	}
}
