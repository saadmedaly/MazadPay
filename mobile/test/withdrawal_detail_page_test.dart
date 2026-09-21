import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/models/wallet.dart';
import 'package:mezadpay/pages/withdrawal_detail_page.dart';

// MAZADPAY — withdrawal notification routing bug: WithdrawalDetailPage is the
// new correct destination for a tapped withdrawal_processed notification
// (previously routed to DepositPage). These tests cover:
//  - Transaction.fromJson correctly parses the two fields this page depends
//    on (beneficiary_account, admin_attachment_url) that were previously
//    absent from the mobile Transaction model entirely.
//  - the page itself never crashes when the network call it makes in
//    initState fails (no live backend in the test environment), matching the
//    same single-pump async-catch-path convention already used throughout
//    this repo's widget tests (e.g. deposit_page.dart's own tests).
void main() {
  group('Transaction.fromJson: withdrawal-detail fields (MAZADPAY notification routing fix)', () {
    test('4/5/6/7. parses amount, beneficiary_account, admin_notes, and admin_attachment_url', () {
      final txn = Transaction.fromJson({
        'id': 'tx-1',
        'user_id': 'user-1',
        'type': 'withdraw',
        'amount': '750',
        'status': 'completed',
        'beneficiary_account': '22334455',
        'admin_notes': 'تم التحويل بنجاح',
        'admin_attachment_url': 'https://example.com/receipt.jpg',
        'created_at': '2026-01-01T00:00:00Z',
      });

      expect(txn.amount, 750);
      expect(txn.beneficiaryAccount, '22334455');
      expect(txn.adminNotes, 'تم التحويل بنجاح');
      expect(txn.adminAttachmentUrl, 'https://example.com/receipt.jpg');
    });

    test('8. missing admin_attachment_url/admin_notes/beneficiary_account parse as null, never crash', () {
      final txn = Transaction.fromJson({
        'id': 'tx-2',
        'user_id': 'user-1',
        'type': 'withdraw',
        'amount': '100',
        'status': 'pending',
        'created_at': '2026-01-01T00:00:00Z',
      });

      expect(txn.beneficiaryAccount, isNull);
      expect(txn.adminNotes, isNull);
      expect(txn.adminAttachmentUrl, isNull);
    });
  });

  group('WithdrawalDetailPage (MAZADPAY notification routing fix)', () {
    Widget wrap(Widget child) {
      return MaterialApp(
        locale: const Locale('ar'),
        home: child,
      );
    }

    testWidgets('8. renders without crashing even when the transaction fetch fails (no live backend)', (tester) async {
      await tester.pumpWidget(wrap(const WithdrawalDetailPage(transactionId: 'tx-does-not-exist')));
      await tester.pump(); // let the async _load() catch-path settle

      expect(find.text('تفاصيل طلب السحب'), findsOneWidget);
    });
  });
}
