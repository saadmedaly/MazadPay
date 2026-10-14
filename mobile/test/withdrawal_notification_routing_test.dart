import 'package:flutter_test/flutter_test.dart';

// MAZADPAY — withdrawal notification routing bug: a tapped withdrawal_processed
// notification previously opened DepositPage (wrong -- that notification is
// about a withdrawal, not a deposit). Both notification_handler.dart's push-tap
// switch and notifications_page.dart's in-app-list tap switch shared the exact
// same bug: 'withdrawal_processed' was grouped in the same case arm as
// payment_received/deposit_confirmed/deposit_rejected.
//
// Firebase/live-navigator mocking is unavailable in this repo (see
// notification_push_tap_lifecycle_test.dart's file-level comment for why), so
// this test reproduces the exact destination-resolution decision both real
// switches now make, as pure functions -- mirroring the same
// pure-logic-reproduction pattern already used for push-tap safety.

/// Mirrors notification_handler.dart's switch (and notifications_page.dart's
/// equivalent in-app tap switch, which resolves identically): given a
/// notification `type` and its `data` map (containing transaction_id for
/// withdrawal_processed), returns the resolved navigation destination.
String resolveNotificationRoute(String? type, Map<String, dynamic> data) {
  switch (type) {
    case 'payment_received':
    case 'deposit_confirmed':
    case 'deposit_rejected':
      return 'deposit';
    case 'withdrawal_processed':
      final transactionId = data['transaction_id']?.toString();
      if (transactionId != null && transactionId.isNotEmpty) {
        return 'withdrawal_detail:$transactionId';
      }
      return 'deposit'; // documented fallback if transaction_id is unexpectedly missing
    default:
      return 'unchanged';
  }
}

void main() {
  group('MAZADPAY: withdrawal notification routing fix', () {
    test('1. deposit_confirmed notification -> still opens Deposit destination', () {
      expect(resolveNotificationRoute('deposit_confirmed', {'transaction_id': 'tx-1'}), 'deposit');
    });

    test('1b. deposit_rejected and payment_received -> still open Deposit destination', () {
      expect(resolveNotificationRoute('deposit_rejected', {}), 'deposit');
      expect(resolveNotificationRoute('payment_received', {}), 'deposit');
    });

    test('2. withdrawal_processed notification -> opens Withdrawal destination, never Deposit', () {
      final outcome = resolveNotificationRoute('withdrawal_processed', {'transaction_id': 'tx-42'});
      expect(outcome, isNot('deposit'));
      expect(outcome, startsWith('withdrawal_detail:'));
    });

    test('3. the correct transaction/request id is passed through to the destination', () {
      final outcome = resolveNotificationRoute('withdrawal_processed', {'transaction_id': 'tx-abc-123'});
      expect(outcome, 'withdrawal_detail:tx-abc-123');
    });

    test('withdrawal_processed falls back safely to Deposit if transaction_id is unexpectedly missing (never crashes)', () {
      expect(resolveNotificationRoute('withdrawal_processed', {}), 'deposit');
    });

    test('9. unrelated notification types are unchanged', () {
      expect(resolveNotificationRoute('auction_won', {}), 'unchanged');
      expect(resolveNotificationRoute('general', {}), 'unchanged');
      expect(resolveNotificationRoute(null, {}), 'unchanged');
    });
  });
}
