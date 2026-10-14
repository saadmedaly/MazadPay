package services

import (
	"testing"
	"time"

	"github.com/mazadpay/backend/internal/models"
	"github.com/shopspring/decimal"
)

func baseAuctionForEdit() models.Auction {
	fr := "Titre"
	loc := 3
	buyNow := decimal.NewFromInt(500)
	return models.Auction{
		CategoryID:      1,
		LocationID:      &loc,
		TitleAr:         "عنوان",
		TitleFr:         &fr,
		StartPrice:      decimal.NewFromInt(100),
		MinIncrement:    decimal.NewFromInt(10),
		InsuranceAmount: decimal.NewFromInt(20),
		BuyNowPrice:     &buyNow,
		StartTime:       time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC),
		ItemDetails:     models.JSONB{"km": float64(1000), "color": "red"},
	}
}

func TestListingChangedFields_IdenticalSaveHasNoChanges(t *testing.T) {
	before := baseAuctionForEdit()
	after := baseAuctionForEdit()
	// Equivalent representations the app/DB round-trip produces.
	after.StartPrice = decimal.RequireFromString("100.00")
	after.EndTime = before.EndTime.Add(400 * time.Microsecond)
	empty := ""
	after.DescriptionAr = &empty
	after.ItemDetails = models.JSONB{"color": "red", "km": float64(1000)}

	if got := listingChangedFields(&before, &after); len(got) != 0 {
		t.Fatalf("expected no changes for an identical save, got %v", got)
	}
}

func TestListingChangedFields_NilAndEmptyItemDetailsEqual(t *testing.T) {
	before := baseAuctionForEdit()
	after := baseAuctionForEdit()
	before.ItemDetails = nil
	after.ItemDetails = models.JSONB{}
	if got := listingChangedFields(&before, &after); len(got) != 0 {
		t.Fatalf("nil and empty item_details should be equal, got %v", got)
	}
}

func TestListingChangedFields_DetectsEachMaterialField(t *testing.T) {
	cases := map[string]func(a *models.Auction){
		"category_id":      func(a *models.Auction) { a.CategoryID = 2 },
		"location_id":      func(a *models.Auction) { a.LocationID = nil },
		"title_ar":         func(a *models.Auction) { a.TitleAr = "جديد" },
		"title_fr":         func(a *models.Auction) { s := "Autre"; a.TitleFr = &s },
		"title_en":         func(a *models.Auction) { s := "Title"; a.TitleEn = &s },
		"description_ar":   func(a *models.Auction) { s := "وصف"; a.DescriptionAr = &s },
		"description_fr":   func(a *models.Auction) { s := "desc"; a.DescriptionFr = &s },
		"description_en":   func(a *models.Auction) { s := "desc"; a.DescriptionEn = &s },
		"start_price":      func(a *models.Auction) { a.StartPrice = decimal.NewFromInt(101) },
		"min_increment":    func(a *models.Auction) { a.MinIncrement = decimal.NewFromInt(11) },
		"insurance_amount": func(a *models.Auction) { a.InsuranceAmount = decimal.NewFromInt(21) },
		"buy_now_price":    func(a *models.Auction) { a.BuyNowPrice = nil },
		"start_time":       func(a *models.Auction) { a.StartTime = a.StartTime.Add(time.Hour) },
		"end_time":         func(a *models.Auction) { a.EndTime = a.EndTime.Add(time.Hour) },
		"phone_contact":    func(a *models.Auction) { s := "+22200000000"; a.PhoneContact = &s },
		"item_details":     func(a *models.Auction) { a.ItemDetails = models.JSONB{"km": float64(2000), "color": "red"} },
	}
	for field, mutate := range cases {
		before := baseAuctionForEdit()
		after := baseAuctionForEdit()
		mutate(&after)
		got := listingChangedFields(&before, &after)
		if len(got) != 1 || got[0] != field {
			t.Errorf("%s: expected exactly [%s], got %v", field, field, got)
		}
	}
}
