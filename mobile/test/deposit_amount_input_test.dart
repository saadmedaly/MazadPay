import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/deposit_page.dart';

// MAZADPAY — admin credit 500 issue: the generic wallet top-up entry point
// (no auction request) previously had NO amount input at all and always sent
// a fixed fallback of 500 as transactions.amount, regardless of how much the
// user actually transferred -- that stored amount is exactly what admin
// approval later credits (see backend AdminService.ValidateTransaction,
// which credits tx.Amount, never a value re-entered at approval time).
//
// These tests confirm: the generic flow now shows a real amount input field
// (so the user can enter the exact amount they transferred), while the
// auction-subscription flow keeps its server-stamped, non-editable
// subscriptionFee display exactly as before.
void main() {
  Widget wrap(Widget child) {
    return MaterialApp(
      locale: const Locale('ar'),
      home: child,
    );
  }

  group('DepositPage amount input (MAZADPAY admin credit 500 issue)', () {
    testWidgets('generic wallet top-up entry point shows an amount input field', (tester) async {
      await tester.pumpWidget(wrap(const DepositPage()));
      await tester.pump();

      expect(find.text('المبلغ الذي قمت بتحويله *'), findsOneWidget);
      expect(find.byType(TextField), findsWidgets);
    });

    testWidgets('generic entry point does not show the auction-subscription fixed fee display', (tester) async {
      await tester.pumpWidget(wrap(const DepositPage()));
      await tester.pump();

      expect(find.text('رسوم النشر المطلوبة'), findsNothing);
    });

    testWidgets('auction-subscription entry point does NOT show the free-text amount input', (tester) async {
      await tester.pumpWidget(wrap(const DepositPage(auctionRequestId: 'req-1', subscriptionFee: 100.0)));
      await tester.pump();

      expect(find.text('المبلغ الذي قمت بتحويله *'), findsNothing);
      // The server-stamped fee is still shown, unchanged, as a fixed display.
      expect(find.text('رسوم النشر المطلوبة'), findsOneWidget);
      expect(find.textContaining('100'), findsWidgets);
    });

    testWidgets('entering an amount in the generic flow updates the controller text', (tester) async {
      await tester.pumpWidget(wrap(const DepositPage()));
      await tester.pump();

      // The amount field is the first TextField in the generic (non-
      // subscription) flow, rendered before phone/notes.
      await tester.enterText(find.byType(TextField).first, '750');
      await tester.pump();

      final widget = tester.widget<TextField>(find.byType(TextField).first);
      expect(widget.controller?.text, '750');
    });
  });
}
