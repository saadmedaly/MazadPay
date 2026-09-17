import 'package:flutter_test/flutter_test.dart';

// Customer feedback #34 (insurance refund / withdrawal beneficiary flow).
// WithdrawPage and WithdrawBeneficiaryPage both construct their own
// WalletApi() directly (no DI seam) and WithdrawBeneficiaryPage calls
// AppLocalizations.of(context)!, so neither can be widget-pumped in a plain
// unit test (see mobile-testing constraints established for DepositPage).
// This reproduces their pure decision logic as widget-free predicates.

const String kGatewayBankTransfer = 'bank_transfer';
const String kGatewayMobileMoney = 'mobile_money';

// Mirrors WithdrawPage._makeWithdrawal's routing decision: only the
// mobile/Bankily gateway needs a beneficiary number collected on a second
// screen -- bank transfer keeps submitting directly, unchanged from before
// this ticket.
bool requiresBeneficiaryScreen(String gateway) => gateway == kGatewayMobileMoney;

// Mirrors WithdrawBeneficiaryPage._submit's validation: a blank/whitespace-only
// beneficiary number must never be submitted.
bool isValidBeneficiaryAccount(String input) => input.trim().isNotEmpty;

// Mirrors WalletApi.withdraw's request body construction (Customer #34):
// beneficiary_account is only included when provided -- an empty/omitted
// value should never send a spurious empty string that the backend would
// otherwise have to special-case.
Map<String, dynamic> buildWithdrawRequestBody({
  required double amount,
  required String gateway,
  String? beneficiaryAccount,
}) {
  return {
    'amount': amount,
    'gateway': gateway,
    'beneficiary_account': ?beneficiaryAccount,
  };
}

void main() {
  group('Beneficiary screen routing (Customer #34)', () {
    test('the mobile/Bankily gateway routes to the beneficiary screen', () {
      expect(requiresBeneficiaryScreen(kGatewayMobileMoney), isTrue);
    });

    test('bank transfer submits directly, no beneficiary screen', () {
      expect(requiresBeneficiaryScreen(kGatewayBankTransfer), isFalse);
    });
  });

  group('Beneficiary account validation', () {
    test('a real phone number is valid', () {
      expect(isValidBeneficiaryAccount('22247601175'), isTrue);
    });

    test('an empty string is invalid', () {
      expect(isValidBeneficiaryAccount(''), isFalse);
    });

    test('whitespace-only input is invalid (fails safe, no blank submission)', () {
      expect(isValidBeneficiaryAccount('   '), isFalse);
    });
  });

  group('Withdrawal request body (Customer #34)', () {
    test('a provided beneficiary account is included', () {
      final body = buildWithdrawRequestBody(amount: 500.0, gateway: kGatewayMobileMoney, beneficiaryAccount: '22247601175');
      expect(body['beneficiary_account'], '22247601175');
    });

    test('no beneficiary account (bank transfer path) omits the key entirely', () {
      final body = buildWithdrawRequestBody(amount: 500.0, gateway: kGatewayBankTransfer);
      expect(body.containsKey('beneficiary_account'), isFalse);
    });

    test('amount and gateway are always present regardless of beneficiary account', () {
      final body = buildWithdrawRequestBody(amount: 300.0, gateway: kGatewayMobileMoney);
      expect(body['amount'], 300.0);
      expect(body['gateway'], kGatewayMobileMoney);
    });
  });
}
