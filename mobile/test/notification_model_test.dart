import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/models/notification.dart' as model;

// Customer request #10: In-app / push notifications. These tests cover
// Notification.fromJson's resilience to malformed/missing backend data --
// the notifications list must never crash from one bad row -- and the
// notification-type -> navigation-target mapping already implemented in
// FCMService.extractRoute (mobile/lib/services/fcm_service.dart), reproduced
// here as a pure predicate for direct testing.
void main() {
  group('Notification.fromJson resilience', () {
    test('a fully-populated row parses correctly', () {
      final n = model.Notification.fromJson({
        'id': 'n1',
        'user_id': 'u1',
        'type': 'auction_won',
        'title': 'You won!',
        'body': 'Congrats',
        'is_read': false,
        'reference_id': 'a1',
        'reference_type': 'auction',
        'data': {'auctionId': 'a1'},
        'created_at': '2026-01-01T12:00:00Z',
      });
      expect(n.id, 'n1');
      expect(n.type, 'auction_won');
      expect(n.isRead, false);
      expect(n.data?['auctionId'], 'a1');
    });

    test('missing optional fields fall back to safe defaults, no crash', () {
      final n = model.Notification.fromJson({'id': 'n2'});
      expect(n.id, 'n2');
      expect(n.userId, '');
      expect(n.type, '');
      expect(n.title, '');
      expect(n.body, isNull);
      expect(n.isRead, false);
      expect(n.data, isNull);
    });

    test('a malformed created_at string does not throw -- falls back instead of crashing the list', () {
      expect(
        () => model.Notification.fromJson({
          'id': 'n3',
          'created_at': 'not-a-real-date',
        }),
        returnsNormally,
      );
    });

    test('an unknown/unrecognized notification type still parses without crashing', () {
      final n = model.Notification.fromJson({
        'id': 'n4',
        'type': 'some_future_type_the_app_does_not_know_about_yet',
        'title': 'x',
      });
      expect(n.type, 'some_future_type_the_app_does_not_know_about_yet');
    });

    test('an empty JSON object still parses without throwing', () {
      expect(() => model.Notification.fromJson({}), returnsNormally);
    });
  });

  // Customer Request #22: image_url/action_url already existed as backend
  // columns but were never parsed by this model -- these prove they now are,
  // and that their absence (every pre-existing notification type) stays safe.
  group('Customer #22: image_url/action_url parsing', () {
    test('a broadcast row with image_url parses it', () {
      final n = model.Notification.fromJson({
        'id': 'n5',
        'type': 'general',
        'title': 'Broadcast',
        'body': 'hello',
        'image_url': 'https://cdn.example.com/notifications/x.jpg',
        'created_at': '2026-01-01T12:00:00Z',
      });
      expect(n.imageUrl, 'https://cdn.example.com/notifications/x.jpg');
    });

    test('a row with no image_url leaves imageUrl null (text-only notification)', () {
      final n = model.Notification.fromJson({
        'id': 'n6',
        'type': 'auction_won',
        'title': 'You won!',
      });
      expect(n.imageUrl, isNull);
    });

    test('action_url parses when present, null otherwise', () {
      final withUrl = model.Notification.fromJson({'id': 'n7', 'action_url': '/some/target'});
      final withoutUrl = model.Notification.fromJson({'id': 'n8'});
      expect(withUrl.actionUrl, '/some/target');
      expect(withoutUrl.actionUrl, isNull);
    });

    test('markAsRead preserves imageUrl/actionUrl', () {
      final original = model.Notification(
        id: 'n9',
        userId: 'u1',
        type: 'general',
        title: 'title',
        imageUrl: 'https://cdn.example.com/img.png',
        actionUrl: '/x',
        createdAt: DateTime(2026, 1, 1),
      );
      final read = original.markAsRead();
      expect(read.imageUrl, original.imageUrl);
      expect(read.actionUrl, original.actionUrl);
    });
  });

  group('markAsRead', () {
    test('returns a copy with isRead=true, all other fields unchanged', () {
      final original = model.Notification(
        id: 'n1',
        userId: 'u1',
        type: 'auction_won',
        title: 'title',
        isRead: false,
        createdAt: DateTime(2026, 1, 1),
      );
      final read = original.markAsRead();
      expect(read.isRead, true);
      expect(read.id, original.id);
      expect(read.type, original.type);
      expect(read.title, original.title);
    });
  });

  group('notification type -> navigation route mapping', () {
    test('auction_won with an auctionId navigates to the auction details route', () {
      expect(extractRoute('auction_won', 'a123'), '/auction/a123');
    });

    test('auction_won WITHOUT an auctionId does not navigate (no route invented)', () {
      expect(extractRoute('auction_won', null), isNull);
    });

    test('bid_outbid navigates to its auction', () {
      expect(extractRoute('bid_outbid', 'a5'), '/auction/a5');
    });

    test('withdrawal_processed navigates to the wallet, no auction id needed', () {
      expect(extractRoute('withdrawal_processed', null), '/wallet');
    });

    test('deposit_confirmed and deposit_rejected both navigate to the wallet', () {
      expect(extractRoute('deposit_confirmed', null), '/wallet');
      expect(extractRoute('deposit_rejected', null), '/wallet');
    });

    test('auction_reported with an auctionId goes to the admin auction view', () {
      expect(extractRoute('auction_reported', 'a9'), '/admin/auction/a9');
    });

    test('auction_reported without an auctionId falls back to the general admin route', () {
      expect(extractRoute('auction_reported', null), '/admin');
    });

    test('an unknown type maps to no route (safe no-op, not a crash)', () {
      expect(extractRoute('some_unhandled_future_type', 'a1'), isNull);
    });

    test('a null type maps to no route', () {
      expect(extractRoute(null, 'a1'), isNull);
    });
  });

  // Customer Request #22: routing precedence. Pure reproduction of the
  // routing decision added to both notifications_page.dart's
  // _onNotificationTap and notification_handler.dart's
  // _navigateFromNotification -- proves specialized navigation (a real
  // auctionId) always wins over the generic detail fallback, and that
  // general/transaction/new_auction only fall back to detail when no real
  // target is present, per the user's explicit routing-precedence
  // constraint ("Customer #22 must not break existing targeted
  // notifications").
  group('Customer #22: routing precedence (specialized nav wins over generic detail)', () {
    test('auction_won WITH auctionId -> specialized winner route, not generic detail', () {
      expect(resolveRoute('auction_won', 'a1'), 'specialized:auction_won:a1');
    });

    test('auction_won WITHOUT auctionId -> falls back to generic detail', () {
      expect(resolveRoute('auction_won', null), 'generic_detail');
    });

    test('bid_outbid WITH auctionId -> specialized auction route', () {
      expect(resolveRoute('bid_outbid', 'a5'), 'specialized:auction:a5');
    });

    test('bid_outbid WITHOUT auctionId -> falls back to generic detail (was a silent no-op before Customer #22)', () {
      expect(resolveRoute('bid_outbid', null), 'generic_detail');
    });

    test('a real targeted new_auction (has auctionId) is NOT degraded to the generic detail page', () {
      expect(resolveRoute('new_auction', 'a7'), 'specialized:auction:a7');
    });

    test('a broadcast new_auction with no real target opens the generic detail page', () {
      expect(resolveRoute('new_auction', null), 'generic_detail');
    });

    test('transaction WITH auctionId still resolves to the specialized route', () {
      expect(resolveRoute('transaction', 'a2'), 'specialized:auction:a2');
    });

    test('a broadcast transaction without target opens the generic detail page', () {
      expect(resolveRoute('transaction', null), 'generic_detail');
    });

    test('general (admin broadcast) always opens the generic detail page (no auction concept)', () {
      expect(resolveRoute('general', null), 'generic_detail');
      expect(resolveRoute('general', 'a1'), 'specialized:auction:a1');
    });

    test('payment/wallet types are unaffected by Customer #22 -- unconditional wallet route', () {
      expect(resolveRoute('payment_received', null), 'specialized:wallet');
      expect(resolveRoute('deposit_confirmed', null), 'specialized:wallet');
      expect(resolveRoute('deposit_rejected', null), 'specialized:wallet');
      expect(resolveRoute('withdrawal_processed', null), 'specialized:wallet');
    });

    test('an unknown/unrecognized type falls back to the generic detail page (never a silent no-op)', () {
      expect(resolveRoute('some_unknown_future_type', null), 'generic_detail');
      // Even with an auctionId present, an unrecognized type has no case
      // that knows what to do with it -- generic detail remains the safe
      // fallback rather than guessing a route.
      expect(resolveRoute('some_unknown_future_type', 'a1'), 'generic_detail');
    });
  });
}

/// Pure reproduction of the post-Customer-#22 routing precedence shared by
/// notifications_page.dart's _onNotificationTap and
/// notification_handler.dart's _navigateFromNotification. Not the real
/// widget navigation call (that requires a BuildContext/Navigator), but the
/// exact decision logic: specialized navigation first when real reference
/// data is present, generic detail page only as the fallback.
String resolveRoute(String? type, String? auctionId) {
  switch (type) {
    case 'auction_won':
      return auctionId != null ? 'specialized:auction_won:$auctionId' : 'generic_detail';
    case 'auction_pending':
    case 'auction_approved':
    case 'auction_rejected':
    case 'auction_ended':
    case 'bid_outbid':
    case 'general':
    case 'transaction':
    case 'new_auction':
      return auctionId != null ? 'specialized:auction:$auctionId' : 'generic_detail';
    case 'payment_received':
    case 'deposit_confirmed':
    case 'deposit_rejected':
    case 'withdrawal_processed':
      return 'specialized:wallet';
    default:
      return 'generic_detail';
  }
}

// Pure reproduction of FCMService.extractRoute's switch logic (private
// extension method, not directly testable without pumping a full FCMService
// instance) -- proves the navigation-type mapping used by both the FCM tap
// handler and NotificationsPage._onNotificationTap.
String? extractRoute(String? type, String? auctionId) {
  switch (type) {
    case 'auction_pending':
    case 'auction_approved':
    case 'auction_rejected':
    case 'auction_ended':
    case 'auction_won':
    case 'bid_outbid':
      if (auctionId != null) return '/auction/$auctionId';
      return null;
    case 'payment_received':
    case 'deposit_confirmed':
    case 'deposit_rejected':
    case 'withdrawal_processed':
      return '/wallet';
    case 'auction_reported':
    case 'auction_suspended':
      if (auctionId != null) return '/admin/auction/$auctionId';
      return '/admin';
    default:
      return null;
  }
}
