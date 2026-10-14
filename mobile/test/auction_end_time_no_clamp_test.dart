import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/models/auction.dart';

// Customer feedback #15 (allow auction duration longer than 24 hours).
// Auction.fromJson used to clamp end_time to start_time + 24h client-side
// (commit e984cfd, "cap auction end time to 24h in API and Flutter"),
// mirroring a matching backend cap that was later fully removed ("client
// feedback Phase B item 15: the 24h max-duration cap is removed -- multi-day
// auctions (48h, 72h, several days) must now be allowed", auction_handler.go)
// -- but the mobile clamp was never cleaned up, silently truncating every
// real multi-day auction back down to 24h on every screen that reads
// Auction.endTime (countdown, active/ended display, bidding UI). This proves
// the real server end_time is now trusted as-is, exactly like start_time.
void main() {
  Map<String, dynamic> auctionJson({required String startTime, required String endTime}) {
    return {
      'id': 'a1',
      'title_ar': 'مزاد اختبار',
      'description_ar': '',
      'images': <dynamic>[],
      'start_price': '100',
      'current_price': '100',
      'min_increment': '10',
      'start_time': startTime,
      'end_time': endTime,
      'bidder_count': 0,
      'views': 0,
      'lot_number': 'LOT-1',
      'seller_phone': '20000000',
      'seller_id': 'seller-1',
    };
  }

  group('Auction.fromJson end_time is never clamped to 24h', () {
    test('a real 48-hour auction keeps its true 48h end_time', () {
      final start = DateTime(2026, 1, 1, 0, 0, 0);
      final end = start.add(const Duration(hours: 48));
      final auction = Auction.fromJson(auctionJson(
        startTime: start.toIso8601String(),
        endTime: end.toIso8601String(),
      ));
      expect(auction.endTime, end);
      expect(auction.endTime.difference(start), const Duration(hours: 48));
    });

    test('a real 72-hour auction keeps its true 72h end_time', () {
      final start = DateTime(2026, 1, 1, 0, 0, 0);
      final end = start.add(const Duration(hours: 72));
      final auction = Auction.fromJson(auctionJson(
        startTime: start.toIso8601String(),
        endTime: end.toIso8601String(),
      ));
      expect(auction.endTime, end);
    });

    test('a real 7-day auction keeps its true 7-day end_time', () {
      final start = DateTime(2026, 1, 1, 0, 0, 0);
      final end = start.add(const Duration(days: 7));
      final auction = Auction.fromJson(auctionJson(
        startTime: start.toIso8601String(),
        endTime: end.toIso8601String(),
      ));
      expect(auction.endTime, end);
      expect(auction.endTime.difference(start), const Duration(days: 7));
    });

    test('a real 23-hour (under 24h) auction is still parsed correctly', () {
      final start = DateTime(2026, 1, 1, 0, 0, 0);
      final end = start.add(const Duration(hours: 23));
      final auction = Auction.fromJson(auctionJson(
        startTime: start.toIso8601String(),
        endTime: end.toIso8601String(),
      ));
      expect(auction.endTime, end);
    });

    test('an exact 24-hour auction is preserved, not shortened', () {
      final start = DateTime(2026, 1, 1, 0, 0, 0);
      final end = start.add(const Duration(hours: 24));
      final auction = Auction.fromJson(auctionJson(
        startTime: start.toIso8601String(),
        endTime: end.toIso8601String(),
      ));
      expect(auction.endTime, end);
    });

    test('missing end_time falls back to now(), not a crash', () {
      final json = auctionJson(startTime: DateTime.now().toIso8601String(), endTime: '');
      json.remove('end_time');
      expect(() => Auction.fromJson(json), returnsNormally);
    });
  });
}
