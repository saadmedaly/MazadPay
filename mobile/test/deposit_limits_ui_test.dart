import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/deposit_page.dart';

// Client feedback (deposit min/max limits): the deposit request screen's
// "تعليمات الدفع" (payment instructions) box previously showed no allowed
// amount range at all. Per the client's exact reference values (MRU 100 /
// MRU 100,000), these tests confirm the two required lines actually render
// inside the existing instructions box -- and that the same bounds are
// authoritatively enforced server-side (see backend/internal/services/
// integrationtest/deposit_limits_test.go for the real
// WalletService.InitiateDeposit coverage: min accepted, max accepted, below
// min rejected, above max rejected).
void main() {
  Widget wrap(Widget child) {
    return MaterialApp(
      locale: const Locale('ar'),
      home: child,
    );
  }

  group('DepositPage payment instructions (Note: deposit min/max limits)', () {
    testWidgets('the minimum deposit amount line renders inside تعليمات الدفع', (tester) async {
      await tester.pumpWidget(wrap(const DepositPage()));
      await tester.pump(); // let the async _loadMethods() catch-path settle

      expect(find.text('تعليمات الدفع'), findsOneWidget);
      expect(find.textContaining('المبلغ الحد الأدنى'), findsOneWidget);
      expect(find.textContaining('MRU 100'), findsWidgets);
    });

    testWidgets('the maximum deposit amount line renders inside تعليمات الدفع', (tester) async {
      await tester.pumpWidget(wrap(const DepositPage()));
      await tester.pump();

      expect(find.textContaining('المبلغ الأقصى'), findsOneWidget);
      expect(find.textContaining('MRU 100000'), findsOneWidget);
    });

    testWidgets('the limits render for the generic wallet top-up entry point (no auction request)', (tester) async {
      await tester.pumpWidget(wrap(const DepositPage()));
      await tester.pump();

      // Same instructions box, same limits, regardless of which entry
      // point reached this page -- the client requirement applies to the
      // deposit screen itself, not a specific flow into it.
      expect(find.textContaining('المبلغ الحد الأدنى'), findsOneWidget);
      expect(find.textContaining('المبلغ الأقصى'), findsOneWidget);
    });

    testWidgets('the limits render for the auction-subscription entry point too', (tester) async {
      await tester.pumpWidget(wrap(const DepositPage(auctionRequestId: 'req-1', subscriptionFee: 100.0)));
      await tester.pump();

      expect(find.textContaining('المبلغ الحد الأدنى'), findsOneWidget);
      expect(find.textContaining('المبلغ الأقصى'), findsOneWidget);
    });

    testWidgets('the existing payment-instructions text and account number are preserved unchanged', (tester) async {
      await tester.pumpWidget(wrap(const DepositPage()));
      await tester.pump();

      expect(find.textContaining('36601175'), findsWidgets);
      expect(find.textContaining('يرجى تحويل مبلغ الاشتراك'), findsOneWidget);
    });

    testWidgets('receipt upload, phone number, and notes fields are preserved unchanged', (tester) async {
      await tester.pumpWidget(wrap(const DepositPage()));
      await tester.pump();

      expect(find.text('رقم الهاتف *'), findsOneWidget);
      expect(find.text('إيصال التحويل *'), findsOneWidget);
      expect(find.text('ملاحظة (اختياري)'), findsOneWidget);
      expect(find.text('إتمام الدفع'), findsOneWidget);
    });
  });
}
