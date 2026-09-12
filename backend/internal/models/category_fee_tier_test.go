package models

import "testing"

// Client feedback #4 (100/500 MRU subscription fee): Category.SubscriptionFee
// must derive the fee ONLY from FeeTier -- never from NameAr/NameFr/NameEn,
// ID, or IconName, none of which are stable identifiers (confirmed: on a
// real local database, category id=1 was "Phones", not the seed migration's
// intended "Real Estate", and icon_name is a free, non-unique, admin-editable
// string). Every existing/legacy category (FeeTier="" or "standard") must
// stay at the cheaper fee unless an admin explicitly set FeeTier="premium".
func TestCategory_SubscriptionFee(t *testing.T) {
	t.Run("an explicit standard tier returns the 100 MRU fee", func(t *testing.T) {
		c := &Category{FeeTier: FeeTierStandard}
		if got := c.SubscriptionFee(); !got.Equal(StandardSubscriptionFee) {
			t.Fatalf("expected %s, got %s", StandardSubscriptionFee, got)
		}
	})

	t.Run("an explicit premium tier returns the 500 MRU fee", func(t *testing.T) {
		c := &Category{FeeTier: FeeTierPremium}
		if got := c.SubscriptionFee(); !got.Equal(PremiumSubscriptionFee) {
			t.Fatalf("expected %s, got %s", PremiumSubscriptionFee, got)
		}
	})

	t.Run("an empty/legacy FeeTier (pre-migration row) defaults to standard, never premium", func(t *testing.T) {
		c := &Category{FeeTier: ""}
		if got := c.SubscriptionFee(); !got.Equal(StandardSubscriptionFee) {
			t.Fatalf("expected legacy category to default to standard fee %s, got %s", StandardSubscriptionFee, got)
		}
	})

	t.Run("an unrecognized/garbage FeeTier value fails safe to standard, never overcharges", func(t *testing.T) {
		c := &Category{FeeTier: "unexpected_value"}
		if got := c.SubscriptionFee(); !got.Equal(StandardSubscriptionFee) {
			t.Fatalf("expected an unrecognized tier to default to standard fee %s, got %s", StandardSubscriptionFee, got)
		}
	})

	t.Run("fee is never derived from category name, regardless of what the name says", func(t *testing.T) {
		// "سيارات" (cars) in NameAr, but FeeTier left at the default -- must
		// still be the standard fee. The fee comes only from an explicit
		// admin FeeTier decision, never from guessing at the name.
		c := &Category{NameAr: "سيارات", NameFr: "Voitures", FeeTier: FeeTierStandard}
		if got := c.SubscriptionFee(); !got.Equal(StandardSubscriptionFee) {
			t.Fatalf("expected name-based category to NOT influence fee (still standard), got %s", got)
		}
	})

	t.Run("fee is never derived from category id", func(t *testing.T) {
		// id=1 does not imply "real estate" or any particular fee -- only
		// FeeTier does.
		c := &Category{ID: 1, FeeTier: FeeTierPremium}
		if got := c.SubscriptionFee(); !got.Equal(PremiumSubscriptionFee) {
			t.Fatalf("expected fee to follow FeeTier regardless of id, got %s", got)
		}
	})
}
