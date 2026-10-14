package services

import (
	"encoding/json"
	"time"

	"github.com/mazadpay/backend/internal/models"
	"github.com/shopspring/decimal"
)

// listingChangedFields compares a seller edit against the stored auction and
// returns the names of the public listing fields that actually changed. The
// mobile app resends every field on each save, so an identical save must
// yield no changes (and must not send a live listing back to review).
//
// Only fields persisted by auctionRepo.UpdateListingTx are compared.
func listingChangedFields(before, after *models.Auction) []string {
	var changed []string
	add := func(name string, differs bool) {
		if differs {
			changed = append(changed, name)
		}
	}

	add("category_id", before.CategoryID != after.CategoryID)
	add("location_id", !intPtrEqual(before.LocationID, after.LocationID))
	add("title_ar", before.TitleAr != after.TitleAr)
	add("title_fr", !strPtrEqual(before.TitleFr, after.TitleFr))
	add("title_en", !strPtrEqual(before.TitleEn, after.TitleEn))
	add("description_ar", !strPtrEqual(before.DescriptionAr, after.DescriptionAr))
	add("description_fr", !strPtrEqual(before.DescriptionFr, after.DescriptionFr))
	add("description_en", !strPtrEqual(before.DescriptionEn, after.DescriptionEn))
	add("start_price", !before.StartPrice.Equal(after.StartPrice))
	add("min_increment", !before.MinIncrement.Equal(after.MinIncrement))
	add("insurance_amount", !before.InsuranceAmount.Equal(after.InsuranceAmount))
	add("buy_now_price", !decimalPtrEqual(before.BuyNowPrice, after.BuyNowPrice))
	add("start_time", !secondsEqual(before.StartTime, after.StartTime))
	add("end_time", !secondsEqual(before.EndTime, after.EndTime))
	add("phone_contact", !strPtrEqual(before.PhoneContact, after.PhoneContact))
	add("item_details", !jsonbEqual(before.ItemDetails, after.ItemDetails))

	return changed
}

// strPtrEqual treats nil and "" as the same value: the service stores an
// empty optional text field as NULL.
func strPtrEqual(a, b *string) bool {
	var av, bv string
	if a != nil {
		av = *a
	}
	if b != nil {
		bv = *b
	}
	return av == bv
}

func intPtrEqual(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func decimalPtrEqual(a, b *decimal.Decimal) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

// secondsEqual ignores sub-second precision: the app sends RFC3339 times
// while Postgres keeps microseconds.
func secondsEqual(a, b time.Time) bool {
	return a.Truncate(time.Second).Equal(b.Truncate(time.Second))
}

// jsonbEqual compares by canonical JSON (encoding/json sorts map keys) and
// treats a nil map and an empty map as equal.
func jsonbEqual(a, b models.JSONB) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	aj, errA := json.Marshal(a)
	bj, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	return string(aj) == string(bj)
}
