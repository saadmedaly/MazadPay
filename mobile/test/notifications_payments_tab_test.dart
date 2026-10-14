import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/notifications_page.dart';

// Customer Request #21 hardening round: deposit_submitted/withdrawal_submitted
// (the new submission-acknowledgment types) rendered correctly under the "All"
// tab but were missing from the "Payments" tab's whitelist -- these tests lock
// down matchesNotificationTabFilter, the exact pure decision
// _filteredNotifications now delegates to, so the Payments tab shows every
// wallet-related notification type.
void main() {
  group('matchesNotificationTabFilter -- payments tab (Customer #21)', () {
    test('deposit_submitted is classified as payment', () {
      expect(matchesNotificationTabFilter('deposit_submitted', 'payments'), isTrue);
    });

    test('deposit_confirmed is classified as payment', () {
      expect(matchesNotificationTabFilter('deposit_confirmed', 'payments'), isTrue);
    });

    test('deposit_rejected is classified as payment', () {
      expect(matchesNotificationTabFilter('deposit_rejected', 'payments'), isTrue);
    });

    test('withdrawal_submitted is classified as payment', () {
      expect(matchesNotificationTabFilter('withdrawal_submitted', 'payments'), isTrue);
    });

    test('withdrawal_processed is classified as payment', () {
      expect(matchesNotificationTabFilter('withdrawal_processed', 'payments'), isTrue);
    });

    test('an unrelated auction notification is not classified as payment', () {
      expect(matchesNotificationTabFilter('auction_won', 'payments'), isFalse);
      expect(matchesNotificationTabFilter('auction_approved', 'payments'), isFalse);
    });

    test('payment_received (pre-existing type) remains classified as payment', () {
      expect(matchesNotificationTabFilter('payment_received', 'payments'), isTrue);
    });
  });

  group('matchesNotificationTabFilter -- auctions tab unaffected', () {
    test('a wallet notification is not classified as an auction notification', () {
      expect(matchesNotificationTabFilter('deposit_submitted', 'auctions'), isFalse);
      expect(matchesNotificationTabFilter('withdrawal_processed', 'auctions'), isFalse);
    });

    test('auction_won remains classified as an auction notification', () {
      expect(matchesNotificationTabFilter('auction_won', 'auctions'), isTrue);
    });
  });

  group('matchesNotificationTabFilter -- "all" tab shows every type', () {
    test('every wallet notification type matches the default/all filter', () {
      for (final type in [
        'deposit_submitted',
        'deposit_confirmed',
        'deposit_rejected',
        'withdrawal_submitted',
        'withdrawal_processed',
        'auction_won',
        'bid_outbid',
        'system',
        'something_unrecognized',
      ]) {
        expect(matchesNotificationTabFilter(type, 'all'), isTrue, reason: '$type must match the "all" filter key');
      }
    });
  });
}
