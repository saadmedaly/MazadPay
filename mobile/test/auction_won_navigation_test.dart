import 'package:flutter_test/flutter_test.dart';

// Customer feedback #11: tapping an auction_won notification (from a push,
// or from the in-app notification list) used to route to plain
// AuctionDetailsPage, identically to every other auction event -- so the
// customer's required congratulations experience (AuctionWinnerPage, which
// already existed but was never wired to anything) never appeared. Both
// notification_handler.dart's _navigateFromNotification and
// notifications_page.dart's _onNotificationTap now special-case auction_won
// to route to the winner page. This reproduces that routing decision as a
// pure predicate -- proving auction_won is isolated from every other
// auction-event type, which must keep going to plain auction details.
enum NavigationTarget { auctionDetails, auctionWinner, wallet, home, none }

NavigationTarget routeForNotificationType(String? type, String? auctionId) {
  switch (type) {
    case 'auction_won':
      return auctionId != null ? NavigationTarget.auctionWinner : NavigationTarget.none;
    case 'auction_pending':
    case 'auction_approved':
    case 'auction_rejected':
    case 'auction_ended':
    case 'bid_outbid':
      return auctionId != null ? NavigationTarget.auctionDetails : NavigationTarget.none;
    case 'payment_received':
    case 'deposit_confirmed':
    case 'deposit_rejected':
    case 'withdrawal_processed':
      return NavigationTarget.wallet;
    default:
      return NavigationTarget.home;
  }
}

void main() {
  group('auction_won routes to the congratulations page, not plain auction details', () {
    test('auction_won with an auctionId routes to AuctionWinnerPage', () {
      expect(routeForNotificationType('auction_won', 'a1'), NavigationTarget.auctionWinner);
    });

    test('auction_won without an auctionId does not navigate (no route invented)', () {
      expect(routeForNotificationType('auction_won', null), NavigationTarget.none);
    });

    test('every other auction event still routes to plain auction details, unchanged', () {
      for (final type in ['auction_pending', 'auction_approved', 'auction_rejected', 'auction_ended', 'bid_outbid']) {
        expect(routeForNotificationType(type, 'a1'), NavigationTarget.auctionDetails,
            reason: 'expected $type to still route to auction details');
      }
    });

    test('wallet-related notification types are unaffected by the auction_won change', () {
      for (final type in ['payment_received', 'deposit_confirmed', 'deposit_rejected', 'withdrawal_processed']) {
        expect(routeForNotificationType(type, null), NavigationTarget.wallet,
            reason: 'expected $type to still route to wallet');
      }
    });

    test('an unknown notification type still falls back to home, no regression', () {
      expect(routeForNotificationType('some_future_type', 'a1'), NavigationTarget.home);
    });
  });
}
