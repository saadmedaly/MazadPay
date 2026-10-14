import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/services/favorites_service.dart';

// Client feedback: Bug C. The Favorites page showed "مزاد #<UUID>" /
// "البيانات غير متوفرة في وضع عدم الاتصال" (data not available offline)
// for every favorite, even while the device was online and other endpoints
// worked. Root cause: GET /users/me/favorites returns a bare JSON array of
// full auction objects (already JOINed server-side), e.g.
// {"success":true,"data":[{"id":"...","title_ar":"...",...}, ...]} -- but
// FavoritesService's old code expected a {"favorites": [...]} wrapper that
// never existed, crashed with a type error the moment it tried to read that
// key off what was actually a List, and silently fell back to an empty/
// local-only favorites cache. This is not a real connectivity problem --
// extractFavoriteAuctionIds/extractFavoriteAuctions are the extracted, pure
// parsing functions now used by FavoritesService, tested directly here
// against the actual backend response shape.
void main() {
  group('extractFavoriteAuctionIds (Bug C contract fix)', () {
    test('extracts real auction IDs from the backend\'s bare array of full auction objects', () {
      final data = [
        {'id': 'auction-1', 'title_ar': 'مزاد أول', 'current_price': '100'},
        {'id': 'auction-2', 'title_ar': 'مزاد ثاني', 'current_price': '200'},
      ];

      final ids = extractFavoriteAuctionIds(data);

      expect(ids, ['auction-1', 'auction-2']);
    });

    test('the old {"favorites": [...]} wrapper is never required -- IDs come from "id" directly', () {
      final data = [
        {'id': 'auction-1'},
      ];

      final ids = extractFavoriteAuctionIds(data);

      expect(ids, ['auction-1']);
    });

    test('an empty array (no favorites) extracts to an empty list, not an error', () {
      expect(extractFavoriteAuctionIds([]), isEmpty);
    });

    test('a non-object entry falls back to its string form rather than crashing', () {
      final ids = extractFavoriteAuctionIds(['bare-id-string']);
      expect(ids, ['bare-id-string']);
    });
  });

  group('extractFavoriteAuctions (Bug C contract fix)', () {
    test('returns the full auction objects directly -- no per-favorite re-fetch needed', () {
      final data = [
        {'id': 'auction-1', 'title_ar': 'مزاد اختبار', 'current_price': '150', 'status': 'active'},
      ];

      final auctions = extractFavoriteAuctions(data);

      expect(auctions, hasLength(1));
      expect(auctions.first['title_ar'], 'مزاد اختبار');
      expect(auctions.first['current_price'], '150');
      expect(auctions.first['status'], 'active');
    });

    test('title is the real auction title, never a UUID fallback', () {
      final data = [
        {'id': '11111111-1111-1111-1111-111111111111', 'title_ar': 'مزاد حقيقي', 'current_price': '50'},
      ];

      final auctions = extractFavoriteAuctions(data);

      expect(auctions.first['title_ar'], 'مزاد حقيقي');
      expect(auctions.first['title_ar'], isNot(contains('#')));
    });

    test('image data survives extraction unchanged for the page to render', () {
      final data = [
        {
          'id': 'auction-1',
          'title_ar': 'مزاد',
          'current_price': '10',
          'images': ['https://example.com/photo.jpg'],
        },
      ];

      final auctions = extractFavoriteAuctions(data);

      expect(auctions.first['images'], ['https://example.com/photo.jpg']);
    });

    test('price is the real auction current_price, not a placeholder', () {
      final data = [
        {'id': 'auction-1', 'title_ar': 'مزاد', 'current_price': '999.50'},
      ];

      final auctions = extractFavoriteAuctions(data);

      expect(auctions.first['current_price'], '999.50');
    });

    test('an ended auction is still returned with its real status, not dropped', () {
      final data = [
        {'id': 'auction-1', 'title_ar': 'مزاد منتهي', 'current_price': '75', 'status': 'ended'},
      ];

      final auctions = extractFavoriteAuctions(data);

      expect(auctions, hasLength(1));
      expect(auctions.first['status'], 'ended');
    });

    test('a non-object entry in the array is skipped, never crashes extraction', () {
      final data = [
        {'id': 'auction-1', 'title_ar': 'مزاد'},
        'unexpected-bare-string',
      ];

      final auctions = extractFavoriteAuctions(data);

      expect(auctions, hasLength(1));
      expect(auctions.first['id'], 'auction-1');
    });

    test('an empty array extracts to an empty list -- this is the true empty-favorites case, not an error', () {
      expect(extractFavoriteAuctions([]), isEmpty);
    });
  });
}
