package services

import "testing"

// TestImagesUpdateIncludesNewURLs is the regression test for targeted test A
// (client feedback, R2 deletion-path audit): AuctionService.Update and
// UpdateAuction (PUT /auctions/:id -- the mobile app's own edit-auction
// endpoint) used to unconditionally wipe existing auction_images rows and
// delete the corresponding R2 objects on every call, including a plain
// field edit (title/price/description/duration only). Root cause:
// AuctionHandler.Update's own request DTO never parses an "images" field
// from the body at all, so input.Images was ALWAYS empty on every real call
// reaching these functions. imagesUpdateIncludesNewURLs is the exact guard
// now shared by both functions -- this test locks down its decision logic
// directly (real code, no database needed, unlike Update/UpdateAuction
// themselves which require a live *sqlx.DB via BeginTxx and cannot be
// driven end-to-end without one).
func TestImagesUpdateIncludesNewURLs(t *testing.T) {
	t.Run("nil slice (a plain field-only PUT /auctions/:id body) does not trigger an image wipe", func(t *testing.T) {
		if imagesUpdateIncludesNewURLs(nil) {
			t.Fatal("SECURITY/DATA-LOSS REGRESSION: a nil Images slice (the exact shape of every real PUT /auctions/:id call, whose DTO has no images field) must never be treated as \"caller supplied new images\"")
		}
	})

	t.Run("empty slice does not trigger an image wipe", func(t *testing.T) {
		if imagesUpdateIncludesNewURLs([]string{}) {
			t.Fatal("expected an empty Images slice to not be treated as a real image update")
		}
	})

	t.Run("slice of only empty strings does not trigger an image wipe", func(t *testing.T) {
		if imagesUpdateIncludesNewURLs([]string{"", "", ""}) {
			t.Fatal("expected a slice of only empty-string entries to not be treated as a real image update")
		}
	})

	t.Run("a genuinely supplied URL does trigger the image-replace path", func(t *testing.T) {
		if !imagesUpdateIncludesNewURLs([]string{"https://pub-example.r2.dev/auctions/x/y.jpg"}) {
			t.Fatal("expected a real supplied URL to trigger the image-replace path")
		}
	})

	t.Run("a mix of empty and real URLs still triggers the image-replace path", func(t *testing.T) {
		if !imagesUpdateIncludesNewURLs([]string{"", "https://pub-example.r2.dev/auctions/x/y.jpg", ""}) {
			t.Fatal("expected at least one non-empty URL among empties to still trigger the image-replace path")
		}
	})
}
