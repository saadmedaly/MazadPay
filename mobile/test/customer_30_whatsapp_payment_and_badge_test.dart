import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/my_winnings_page.dart';
import 'package:mezadpay/providers/unread_notifications_provider.dart';

// REAL DEVICE BUG fix: on Android 11+ (targetSdk 30+, package visibility),
// canLaunchUrl()/launchUrl() for the primary https://wa.me/... link can fail
// even with WhatsApp installed unless the host app declares <queries>
// visibility (fixed in AndroidManifest.xml). buildWinnerPaymentWhatsAppNativeUri
// is the fallback path -- WhatsApp's own native whatsapp://send deep link --
// tested here the same way the primary URI builder already is.

// Customer Request #30: WhatsApp payment button (My Winnings) + Home
// notification bell unread badge. Neither my_winnings_page.dart nor
// home_page.dart has a DI seam (both construct their own API services /
// read Riverpod providers directly), so the full pages cannot be
// widget-tested -- the specific testable logic (URI building, unread
// counting) and minimal standalone widget reproductions of the exact shapes
// are used instead, matching the established pattern for this project (see
// bid_cta_expiry_test.dart, auction_details_bid_count_test.dart).

void main() {
  group('WhatsApp payment button (Part A)', () {
    test('1. opens the correct MazadPay WhatsApp number, not the per-auction seller number', () {
      final uri = buildWinnerPaymentWhatsAppUri(
        auctionTitle: 'سيارة أوتوماتيك للبيع',
        lotNumber: '116',
        formattedAmount: 'MRU 7,200',
      );
      expect(uri.host, 'wa.me');
      expect(uri.path, '/22247601175');
      expect(kMazadPayWhatsAppNumber, '47601175');
    });

    test('2. prefilled message contains title, LOT, and amount but never private data', () {
      final uri = buildWinnerPaymentWhatsAppUri(
        auctionTitle: 'سيارة أوتوماتيك للبيع',
        lotNumber: '116',
        formattedAmount: 'MRU 7,200',
      );
      final message = uri.queryParameters['text']!;

      expect(message, contains('سيارة أوتوماتيك للبيع'));
      expect(message, contains('LOT-116'));
      expect(message, contains('MRU 7,200'));

      // Safety: never a JWT, UUID-shaped id, email, or raw phone number.
      final uuidPattern = RegExp(r'[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}');
      expect(uuidPattern.hasMatch(message), isFalse);
      expect(message.toLowerCase(), isNot(contains('jwt')));
      expect(message.toLowerCase(), isNot(contains('token')));
      expect(message, isNot(contains('@')));
      expect(message, isNot(contains(RegExp(r'\+?222\d{8}'))), reason: 'must never embed a raw phone number in the message body');
    });

    test('message omits the LOT segment safely when lotNumber is null/empty', () {
      final uriNull = buildWinnerPaymentWhatsAppUri(
        auctionTitle: 'سيارة',
        lotNumber: null,
        formattedAmount: 'MRU 100',
      );
      expect(uriNull.queryParameters['text'], isNot(contains('LOT-')));

      final uriEmpty = buildWinnerPaymentWhatsAppUri(
        auctionTitle: 'سيارة',
        lotNumber: '   ',
        formattedAmount: 'MRU 100',
      );
      expect(uriEmpty.queryParameters['text'], isNot(contains('LOT-')));
    });
  });

  group('WhatsApp native fallback URI (REAL DEVICE BUG fix)', () {
    test('uses the whatsapp:// native scheme with the correct international phone', () {
      final uri = buildWinnerPaymentWhatsAppNativeUri(
        auctionTitle: 'سيارة أوتوماتيك للبيع',
        lotNumber: '116',
        formattedAmount: 'MRU 7,200',
      );
      expect(uri.scheme, 'whatsapp');
      expect(uri.host, 'send');
      expect(uri.queryParameters['phone'], '22247601175');
    });

    test('carries the exact same prefilled message as the primary wa.me URI', () {
      final primary = buildWinnerPaymentWhatsAppUri(
        auctionTitle: 'سيارة أوتوماتيك للبيع',
        lotNumber: '116',
        formattedAmount: 'MRU 7,200',
      );
      final fallback = buildWinnerPaymentWhatsAppNativeUri(
        auctionTitle: 'سيارة أوتوماتيك للبيع',
        lotNumber: '116',
        formattedAmount: 'MRU 7,200',
      );
      expect(fallback.queryParameters['text'], primary.queryParameters['text']);
    });

    test('omits the LOT segment safely when lotNumber is null, same as the primary URI', () {
      final uri = buildWinnerPaymentWhatsAppNativeUri(
        auctionTitle: 'سيارة',
        lotNumber: null,
        formattedAmount: 'MRU 100',
      );
      expect(uri.queryParameters['text'], isNot(contains('LOT-')));
    });
  });

  group('unread notification count (Part B)', () {
    test('3. unread = 0 when no notifications or all are read', () {
      expect(UnreadNotificationsCount.countUnread([]), 0);
      expect(
        UnreadNotificationsCount.countUnread([
          {'id': '1', 'is_read': true},
          {'id': '2', 'is_read': true},
        ]),
        0,
      );
    });

    test('4. unread = 1 counted correctly', () {
      expect(
        UnreadNotificationsCount.countUnread([
          {'id': '1', 'is_read': false},
          {'id': '2', 'is_read': true},
        ]),
        1,
      );
    });

    test('5. multi-digit unread count computed safely', () {
      final many = List.generate(147, (i) => {'id': '$i', 'is_read': false});
      expect(UnreadNotificationsCount.countUnread(many), 147);
    });

    test('missing is_read field defaults to unread (matches Notification.fromJson\'s own `?? false` default)', () {
      expect(
        UnreadNotificationsCount.countUnread([
          {'id': '1'},
        ]),
        1,
      );
    });

    test('non-map entries are safely ignored, never crash the count', () {
      expect(UnreadNotificationsCount.countUnread(['not-a-map', 42, null]), 0);
    });
  });

  group('unread badge widget (minimal standalone reproduction)', () {
    Widget buildBell(int unreadCount) {
      return MaterialApp(
        home: Scaffold(
          body: Stack(
            clipBehavior: Clip.none,
            children: [
              const Icon(Icons.notifications_outlined),
              if (unreadCount > 0)
                Positioned(
                  right: 4,
                  top: 4,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                    constraints: const BoxConstraints(minWidth: 16, minHeight: 16),
                    decoration: BoxDecoration(color: Colors.red, borderRadius: BorderRadius.circular(8)),
                    child: Text(
                      unreadCount > 99 ? '99+' : '$unreadCount',
                      textAlign: TextAlign.center,
                      style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold, height: 1.2),
                    ),
                  ),
                ),
            ],
          ),
        ),
      );
    }

    testWidgets('3b. unread = 0 -> badge hidden', (tester) async {
      await tester.pumpWidget(buildBell(0));
      expect(find.byType(Positioned), findsNothing);
    });

    testWidgets('4b. unread = 1 -> badge shows "1"', (tester) async {
      await tester.pumpWidget(buildBell(1));
      expect(find.text('1'), findsOneWidget);
    });

    testWidgets('5b. multi-digit unread count displays without overflow (99+ cap)', (tester) async {
      await tester.pumpWidget(buildBell(150));
      expect(tester.takeException(), isNull);
      expect(find.text('99+'), findsOneWidget);
    });

    testWidgets('6. badge updates after notification state changes (rebuild with a new count)', (tester) async {
      await tester.pumpWidget(buildBell(3));
      expect(find.text('3'), findsOneWidget);

      await tester.pumpWidget(buildBell(0));
      expect(find.text('3'), findsNothing);
      expect(find.byType(Positioned), findsNothing);

      await tester.pumpWidget(buildBell(5));
      expect(find.text('5'), findsOneWidget);
    });
  });

  group('My Winnings still renders won auctions (Part D item 7 -- regression guard)', () {
    // Bug L's array-parsing fix and the title/price/image extraction in
    // _buildWinningItem were NOT touched by this ticket -- only the pay
    // button's onPressed callback was rewired from a no-op to
    // _openPaymentWhatsApp. This confirms the exact same winning-item shape
    // My Winnings already parses correctly still flows into the new WhatsApp
    // message builder without any missing/renamed field breaking it.
    test('a real won-auction item\'s title/lot_number/current_price map correctly into the payment message', () {
      final winning = {
        'id': 'irrelevant-for-this-test',
        'title_ar': 'ابردوا 2025',
        'lot_number': '116',
        'current_price': '90200',
        'currency_code': 'MRU',
      };

      final uri = buildWinnerPaymentWhatsAppUri(
        auctionTitle: winning['title_ar'] as String,
        lotNumber: winning['lot_number'] as String,
        formattedAmount: 'MRU 90,200',
      );
      final message = uri.queryParameters['text']!;

      expect(message, contains('ابردوا 2025'));
      expect(message, contains('LOT-116'));
      expect(message, contains('MRU 90,200'));
    });
  });
}
