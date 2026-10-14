import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/utils/auction_image.dart';

// Client feedback: Bug H. No generic/default product image (e.g. a car
// photo) may ever stand in for an auction's real image. Auction "حرامي"
// (whose real R2 image was deleted, see Bug G) showed a Toyota Corolla
// stock photo on Home/Live Auction and Auction History cards -- misleading
// users into thinking that was the auction's actual photo. Root cause:
// several screens defaulted a missing image URL to 'assets/corolla.png',
// a real bundled asset that rendered successfully, so the neutral
// missing-image placeholder those same screens already had was never
// reached. resolveAuctionImageUrl is the extracted, pure decision helper
// now actually wired into home_page.dart, auction_history_page.dart,
// my_winnings_page.dart and auction_details_page.dart's related-auction
// card (not just documented in isolation) -- these tests exercise the
// exact function each of those screens calls.
void main() {
  group('resolveAuctionImageUrl (Bug H)', () {
    test('a valid image URL is selected', () {
      expect(resolveAuctionImageUrl(['https://example.com/photo.jpg']), 'https://example.com/photo.jpg');
    });

    test('an empty URL list resolves to null, never a fake stock asset', () {
      expect(resolveAuctionImageUrl([]), isNull);
    });

    test('null/empty is never resolved to assets/corolla.png', () {
      final result = resolveAuctionImageUrl([]);
      expect(result, isNot('assets/corolla.png'));
      expect(result, isNull);
    });

    test('a list containing only an empty string resolves to null', () {
      expect(resolveAuctionImageUrl(['']), isNull);
    });

    test('a list containing only whitespace resolves to null', () {
      expect(resolveAuctionImageUrl(['   ']), isNull);
    });

    test('multiple images: the first non-empty URL is selected', () {
      expect(
        resolveAuctionImageUrl(['https://example.com/a.jpg', 'https://example.com/b.jpg']),
        'https://example.com/a.jpg',
      );
    });

    test('a leading empty entry is skipped in favor of the next real URL', () {
      expect(
        resolveAuctionImageUrl(['', 'https://example.com/real.jpg']),
        'https://example.com/real.jpg',
      );
    });
  });

  group('resolveAuctionImageUrl -- no-corolla-fallback guarantee (Bug H)', () {
    test('the decision is null-or-real-URL only -- there is no branch that can return a stock asset path', () {
      // resolveAuctionImageUrl has exactly one return path for "no real
      // image": null. It never returns a literal asset path, so no caller
      // can receive 'assets/corolla.png' (or any other hardcoded asset)
      // from this function by construction.
      expect(resolveAuctionImageUrl([]), isNull);
      expect(resolveAuctionImageUrl(['']), isNull);
      expect(resolveAuctionImageUrl(['assets/corolla.png']), 'assets/corolla.png');
      // The one case where the literal string appears is when a caller's
      // OWN data contains it -- proving the function is a pure pass-through
      // with no independent corolla-injecting branch of its own.
    });
  });
}
