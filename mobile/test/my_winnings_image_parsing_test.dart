import 'package:flutter_test/flutter_test.dart';

// Customer feedback #11: /users/me/winnings returns auctions via
// AuctionRepository.ListPaginated, whose image_urls field is a
// comma-separated STRING (models.Auction.ImageURLs), not a JSON list. The
// image-resolution logic in my_winnings_page.dart's _buildWinningItem only
// checked `is List` for image_urls, which always failed for this shape, so
// every won item silently fell back to the generic placeholder asset even
// when the auction had real images. This reproduces the fixed resolution
// logic (mirroring the same comma-separated-string-or-List handling already
// established in my_auctions_page.dart for the identical field) as a pure,
// widget-free predicate.
String resolveWinningImageUrl(Map<String, dynamic> winning) {
  String imageUrl = 'assets/corolla.png';
  final rawImageUrls = winning['image_urls'];
  if (rawImageUrls != null && rawImageUrls.toString().isNotEmpty) {
    if (rawImageUrls is List && rawImageUrls.isNotEmpty) {
      imageUrl = rawImageUrls[0].toString();
    } else {
      imageUrl = rawImageUrls.toString().split(',').first.trim();
    }
  } else if (winning['images'] != null && winning['images'] is List && (winning['images'] as List).isNotEmpty) {
    imageUrl = (winning['images'] as List)[0].toString();
  } else if (winning['image_url'] != null) {
    imageUrl = winning['image_url'].toString();
  } else if (winning['image'] != null) {
    imageUrl = winning['image'].toString();
  }
  return imageUrl;
}

void main() {
  group('My Winnings image resolution', () {
    test('a comma-separated image_urls string (the real backend shape) resolves to its first URL', () {
      final url = resolveWinningImageUrl({
        'id': 'a1',
        'image_urls': 'https://cdn.example.com/1.jpg,https://cdn.example.com/2.jpg',
      });
      expect(url, 'https://cdn.example.com/1.jpg');
    });

    test('a single-URL image_urls string with no comma resolves correctly', () {
      final url = resolveWinningImageUrl({
        'id': 'a2',
        'image_urls': 'https://cdn.example.com/only.jpg',
      });
      expect(url, 'https://cdn.example.com/only.jpg');
    });

    test('a List-shaped image_urls (defensive, in case a future endpoint returns one) still works', () {
      final url = resolveWinningImageUrl({
        'id': 'a3',
        'image_urls': ['https://cdn.example.com/list1.jpg'],
      });
      expect(url, 'https://cdn.example.com/list1.jpg');
    });

    test('an empty image_urls string falls through to the images list field', () {
      final url = resolveWinningImageUrl({
        'id': 'a4',
        'image_urls': '',
        'images': ['https://cdn.example.com/fallback.jpg'],
      });
      expect(url, 'https://cdn.example.com/fallback.jpg');
    });

    test('no image fields at all falls back to the local placeholder asset, no crash', () {
      final url = resolveWinningImageUrl({'id': 'a5'});
      expect(url, 'assets/corolla.png');
    });
  });
}
