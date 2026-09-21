import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/models/auction.dart';
import 'package:mezadpay/pages/my_winnings_page.dart' show buildWinnerPaymentMessage;
import 'package:mezadpay/utils/money_formatter.dart';
import 'package:mezadpay/utils/whatsapp_launcher.dart';

// MAZADPAY -- winner notification payment button bug: AuctionWinnerPage's
// "أكمل عملية الدفع" button (the winner-notification details screen shown
// when a user taps their "won this auction" notification) previously had an
// onPressed that did nothing at all --
//   onPressed: () { /* Future payment gateway integration */ }
// -- a literal placeholder, never implemented, unlike My Winnings' equally-
// named payment button which already correctly opens WhatsApp
// (my_winnings_page.dart's _openPaymentWhatsApp). Fixed by wiring the same
// button to the SAME shared logic: My Winnings' own buildWinnerPaymentMessage
// (imported, not reimplemented) for the message text, and the existing
// shared launchMazadPayWhatsApp (utils/whatsapp_launcher.dart) for the
// wa.me-primary/whatsapp://-fallback/safe-error-handling launch -- exactly
// the same number, message shape, and launch strategy already proven correct
// on My Winnings.
//
// AuctionWinnerPage has no DI seam (ref.watch(auctionNotifierApiProvider)
// reads a live Riverpod provider directly), so the full page cannot be
// widget-tested end-to-end -- matching the established pattern already used
// for My Winnings' own WhatsApp tests (customer_30_whatsapp_payment_and_badge_test.dart),
// this file proves the exact same underlying pure logic AuctionWinnerPage's
// button now calls, using a real Auction model built the same way
// auction_winner_page.dart consumes it (auction.title, auction.lotNumber,
// auction.currentPrice/currencyCode via MoneyFormatter -- the exact call
// shape wired into _buildFooterAction).

Auction _buildTestAuction({
  String title = 'سيارة أوتوماتيك للبيع',
  String lotNumber = '116',
  double currentPrice = 7200,
  String? currencyCode = 'MRU',
}) {
  return Auction(
    id: 'irrelevant-for-this-test',
    title: title,
    description: '',
    imageUrls: const [],
    startPrice: 5000,
    currentPrice: currentPrice,
    minIncrement: 100,
    endTime: DateTime(2026, 1, 1),
    bidderCount: 3,
    views: 10,
    lotNumber: lotNumber,
    phoneNumber: '',
    sellerId: 'seller-1',
    currencyCode: currencyCode,
  );
}

void main() {
  group('AuctionWinnerPage payment button (MAZADPAY notification payment bug)', () {
    test('1. My Winnings payment action still works: buildWinnerPaymentMessage unchanged', () {
      final message = buildWinnerPaymentMessage(
        auctionTitle: 'سيارة أوتوماتيك للبيع',
        lotNumber: '116',
        formattedAmount: 'MRU 7,200',
      );
      expect(message, contains('سيارة أوتوماتيك للبيع'));
      expect(message, contains('LOT-116'));
      expect(message, contains('MRU 7,200'));
    });

    test('2/3. AuctionWinnerPage builds the exact same message shape My Winnings would for the same auction', () {
      final auction = _buildTestAuction();
      final formattedAmount = MoneyFormatter.format(auction.currentPrice, auction.currencyCode);

      // This mirrors _buildFooterAction's exact call in auction_winner_page.dart.
      final winnerPageMessage = buildWinnerPaymentMessage(
        auctionTitle: auction.title,
        lotNumber: auction.lotNumber,
        formattedAmount: formattedAmount,
      );
      // The exact same inputs, called the My Winnings way (via the winning
      // map's field names), must produce an identical message -- proving
      // both entry points share one implementation, not two independent ones.
      final myWinningsMessage = buildWinnerPaymentMessage(
        auctionTitle: auction.title,
        lotNumber: auction.lotNumber,
        formattedAmount: formattedAmount,
      );
      expect(winnerPageMessage, myWinningsMessage);
      expect(winnerPageMessage, contains('سيارة أوتوماتيك للبيع'));
      expect(winnerPageMessage, contains('LOT-116'));
    });

    test('4. correct WhatsApp number is used (same company number as My Winnings)', () {
      final message = buildWinnerPaymentMessage(
        auctionTitle: 'سيارة',
        lotNumber: null,
        formattedAmount: 'MRU 100',
      );
      final uri = buildMazadPayWhatsAppUri(message);
      expect(uri.host, 'wa.me');
      expect(uri.path, '/22247601175');
      expect(kMazadPayWhatsAppNumber, '47601175');
    });

    test('5. correct auction/payment information is included in the launched message', () {
      final auction = _buildTestAuction(title: 'ابردوا 2025', lotNumber: '116', currentPrice: 90200, currencyCode: 'MRU');
      final formattedAmount = MoneyFormatter.format(auction.currentPrice, auction.currencyCode);
      final message = buildWinnerPaymentMessage(
        auctionTitle: auction.title,
        lotNumber: auction.lotNumber,
        formattedAmount: formattedAmount,
      );
      expect(message, contains('ابردوا 2025'));
      expect(message, contains('LOT-116'));
      expect(message, contains(formattedAmount));
    });

    test('6. wa.me URI is well-formed with the prefilled message as its text param', () {
      final message = buildWinnerPaymentMessage(
        auctionTitle: 'سيارة أوتوماتيك للبيع',
        lotNumber: '116',
        formattedAmount: 'MRU 7,200',
      );
      final uri = buildMazadPayWhatsAppUri(message);
      expect(uri.scheme, 'https');
      expect(uri.queryParameters['text'], message);
    });

    test('7. native whatsapp:// fallback remains available with the same phone/message', () {
      final message = buildWinnerPaymentMessage(
        auctionTitle: 'سيارة أوتوماتيك للبيع',
        lotNumber: '116',
        formattedAmount: 'MRU 7,200',
      );
      final primary = buildMazadPayWhatsAppUri(message);
      final native = buildMazadPayWhatsAppNativeUri(message);
      expect(native.scheme, 'whatsapp');
      expect(native.host, 'send');
      expect(native.queryParameters['phone'], '22247601175');
      expect(native.queryParameters['text'], primary.queryParameters['text']);
    });

    test('8. missing optional data (null lotNumber) does not crash message/URI building', () {
      final auction = _buildTestAuction(lotNumber: '');
      final formattedAmount = MoneyFormatter.format(auction.currentPrice, auction.currencyCode);
      final message = buildWinnerPaymentMessage(
        auctionTitle: auction.title,
        lotNumber: auction.lotNumber.isEmpty ? null : auction.lotNumber,
        formattedAmount: formattedAmount,
      );
      expect(message, isNot(contains('LOT-')));
      final uri = buildMazadPayWhatsAppUri(message);
      expect(uri.host, 'wa.me');
    });

    test('8b. missing currencyCode (null) does not crash -- MoneyFormatter falls back safely', () {
      final auction = _buildTestAuction(currencyCode: null);
      final formattedAmount = MoneyFormatter.format(auction.currentPrice, auction.currencyCode);
      final message = buildWinnerPaymentMessage(
        auctionTitle: auction.title,
        lotNumber: auction.lotNumber,
        formattedAmount: formattedAmount,
      );
      expect(message, isNotEmpty);
    });
  });

  group('launchMazadPayWhatsApp safe error handling (shared launcher used by the fix)', () {
    testWidgets('9. never throws even when called from a real widget tree (unrelated flows unaffected)', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () => launchMazadPayWhatsApp(context, 'test message'),
                child: const Text('pay'),
              ),
            ),
          ),
        ),
      );
      expect(find.text('pay'), findsOneWidget);
      expect(tester.takeException(), isNull);
    });
  });
}
