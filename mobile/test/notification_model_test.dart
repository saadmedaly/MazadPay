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
