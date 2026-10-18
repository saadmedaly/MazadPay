import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/models/bankily_payment.dart';
import 'package:mezadpay/pages/bankily_bpay_page.dart';
import 'package:mezadpay/pages/bankily_payment_status_page.dart';
import 'package:mezadpay/pages/deposit_page.dart';
import 'package:mezadpay/services/api_service.dart';
import 'package:mezadpay/services/bankily_api.dart';

const _cfg = BankilyConfig(enabled: true, merchantCode: '027922', minAmount: 100, maxAmount: 100000, allowDecimals: false);

BankilyPayment _payment(String status, {String id = 'p1'}) => BankilyPayment.fromJson({
      'id': id,
      'operation_id': 'MZP1234567890123',
      'amount': '500',
      'currency_code': 'MRU',
      'client_phone': '48****11',
      'status': status,
      'bankily_transaction_id': status == 'SUCCESS' ? 'BK-1' : null,
    });

class FakeBankilyApi extends BankilyApi {
  final List<Map<String, String>> payCalls = [];
  final List<Object> payErrors;
  BankilyPayment payResult;
  final List<BankilyPayment> getResults;
  int getCalls = 0;
  int checkCalls = 0;
  BankilyPayment? checkResult;

  FakeBankilyApi({List<Object>? payErrors, BankilyPayment? payResult, List<BankilyPayment>? getResults})
      : payErrors = payErrors ?? [],
        payResult = payResult ?? _payment('SUCCESS'),
        getResults = getResults ?? [],
        super(api: null);

  @override
  Future<BankilyPayment> pay({required String amount, required String clientPhone, required String passcode, required String idempotencyKey}) async {
    payCalls.add({'amount': amount, 'phone': clientPhone, 'passcode': passcode, 'key': idempotencyKey});
    if (payErrors.isNotEmpty) throw payErrors.removeAt(0);
    return payResult;
  }

  @override
  Future<BankilyPayment> getPayment(String id) async {
    getCalls++;
    return getResults.isNotEmpty ? getResults.removeAt(0) : _payment('PENDING_VERIFICATION');
  }

  @override
  Future<BankilyPayment> check(String id) async {
    checkCalls++;
    return checkResult ?? _payment('PENDING_VERIFICATION');
  }

  @override
  Future<List<BankilyPayment>> list() async => const [];
}

Widget _wrap(Widget child, {Locale locale = const Locale('fr')}) => MaterialApp(
      locale: locale,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: AppLocalizations.supportedLocales,
      home: child,
    );

Future<void> _fill(WidgetTester tester, {String amount = '500', String phone = '48110011', String passcode = '2580'}) async {
  await tester.enterText(find.byKey(const Key('bpay-amount')), amount);
  await tester.enterText(find.byKey(const Key('bpay-phone')), phone);
  await tester.enterText(find.byKey(const Key('bpay-passcode')), passcode);
}

Future<void> _tapPay(WidgetTester tester) async {
  await tester.ensureVisible(find.byKey(const Key('bpay-pay')));
  await tester.tap(find.byKey(const Key('bpay-pay')));
}

void main() {
  group('BankilyPayment model and input rules', () {
    test('status parsing folds internal states into pending', () {
      expect(BankilyPayment.parseStatus('SUCCESS'), BankilyPaymentStatus.success);
      expect(BankilyPayment.parseStatus('FAILED'), BankilyPaymentStatus.failed);
      expect(BankilyPayment.parseStatus('PENDING_VERIFICATION'), BankilyPaymentStatus.pending);
      expect(BankilyPayment.parseStatus('SUBMITTED'), BankilyPaymentStatus.pending);
      expect(BankilyPayment.parseStatus(null), BankilyPaymentStatus.pending);
    });

    test('phone rules accept Mauritanian mobiles only', () {
      expect(BankilyInputRules.isValidPhone('48110011'), isTrue);
      expect(BankilyInputRules.isValidPhone('+222 48 11 00 11'), isTrue);
      expect(BankilyInputRules.isValidPhone('12345678'), isFalse);
      expect(BankilyInputRules.isValidPhone('4811001'), isFalse);
    });

    test('amount rules follow backend config', () {
      expect(BankilyInputRules.isValidAmount(100, _cfg), isTrue);
      expect(BankilyInputRules.isValidAmount(99, _cfg), isFalse);
      expect(BankilyInputRules.isValidAmount(100.5, _cfg), isFalse);
      expect(BankilyInputRules.isValidAmount(100001, _cfg), isFalse);
    });

    test('timeouts and 5xx are uncertain; validation errors are not', () {
      expect(BankilyApi.isUncertain(const ApiException('t', code: 'receiveTimeout')), isTrue);
      expect(BankilyApi.isUncertain(const ApiException('s', code: 'server_error', httpStatus: 502)), isTrue);
      expect(BankilyApi.isUncertain(const ApiException('v', code: 'bankily_invalid_phone', httpStatus: 422)), isFalse);
    });
  });

  group('BankilyBpayPage', () {
    testWidgets('shows merchant code and masks the passcode', (tester) async {
      await tester.pumpWidget(_wrap(BankilyBpayPage(config: _cfg, api: FakeBankilyApi())));
      await tester.pump();
      expect(find.byKey(const Key('bpay-merchant-code')), findsOneWidget);
      expect(find.text('027922'), findsOneWidget);
      final passcode = tester.widget<TextField>(find.byKey(const Key('bpay-passcode')));
      expect(passcode.obscureText, isTrue);
      expect(passcode.enableSuggestions, isFalse);
      expect(passcode.autocorrect, isFalse);
    });

    testWidgets('invalid input never reaches the backend', (tester) async {
      final api = FakeBankilyApi();
      await tester.pumpWidget(_wrap(BankilyBpayPage(config: _cfg, api: api)));
      await _fill(tester, phone: '12345678');
      await _tapPay(tester);
      await tester.pump();
      expect(api.payCalls, isEmpty);
      expect(find.text('Numéro Bankily invalide'), findsOneWidget);
    });

    testWidgets('success opens the success state and clears the passcode', (tester) async {
      final api = FakeBankilyApi(payResult: _payment('SUCCESS'));
      await tester.pumpWidget(_wrap(BankilyBpayPage(config: _cfg, api: api)));
      await _fill(tester);
      await _tapPay(tester);
      await tester.pumpAndSettle();
      expect(api.payCalls, hasLength(1));
      expect(api.payCalls.single['phone'], '48110011');
      expect(find.byKey(const ValueKey('bpay-status-success')), findsOneWidget);
      expect(find.text('Portefeuille rechargé'), findsOneWidget);
    });

    testWidgets('network timeout retries with the SAME idempotency key', (tester) async {
      final api = FakeBankilyApi(
        payErrors: [const ApiException('timeout', code: 'receiveTimeout')],
        payResult: _payment('PENDING_VERIFICATION'),
      );
      await tester.pumpWidget(_wrap(BankilyBpayPage(config: _cfg, api: api)));
      await _fill(tester);
      await _tapPay(tester);
      await tester.pump(const Duration(seconds: 3));
      await tester.pump();
      await tester.pump();
      expect(api.payCalls, hasLength(2));
      expect(api.payCalls[0]['key'], api.payCalls[1]['key']);
      expect(find.byKey(const ValueKey('bpay-status-pending')), findsOneWidget);
      // Stop the status page's poll timer.
      await tester.pumpWidget(const SizedBox());
    });

    testWidgets('double tap submits once', (tester) async {
      final api = FakeBankilyApi(payResult: _payment('SUCCESS'));
      await tester.pumpWidget(_wrap(BankilyBpayPage(config: _cfg, api: api)));
      await _fill(tester);
      await _tapPay(tester);
      await tester.tap(find.byKey(const Key('bpay-pay')), warnIfMissed: false);
      await tester.pumpAndSettle();
      expect(api.payCalls, hasLength(1));
    });

    testWidgets('a definite rejection is shown and not retried', (tester) async {
      final api = FakeBankilyApi(payErrors: [const ApiException('x', code: 'bankily_invalid_amount', httpStatus: 422)]);
      await tester.pumpWidget(_wrap(BankilyBpayPage(config: _cfg, api: api)));
      await _fill(tester);
      await _tapPay(tester);
      await tester.pumpAndSettle();
      expect(api.payCalls, hasLength(1));
      expect(find.text('Montant invalide'), findsOneWidget);
    });
  });

  group('BankilyPaymentStatusPage', () {
    testWidgets('failed shows retry and manual fallback', (tester) async {
      await tester.pumpWidget(_wrap(BankilyPaymentStatusPage(payment: _payment('FAILED'), api: FakeBankilyApi()), locale: const Locale('ar')));
      await tester.pump();
      expect(find.byKey(const ValueKey('bpay-status-failed')), findsOneWidget);
      expect(find.text('حاول مجدداً'), findsOneWidget);
      expect(find.text('استخدام التحويل اليدوي'), findsOneWidget);
    });

    testWidgets('pending polls until the backend reports success', (tester) async {
      final api = FakeBankilyApi(getResults: [_payment('PENDING_VERIFICATION'), _payment('SUCCESS')]);
      await tester.pumpWidget(_wrap(
        BankilyPaymentStatusPage(payment: _payment('PENDING_VERIFICATION'), api: api, pollInterval: const Duration(milliseconds: 10)),
        locale: const Locale('en'),
      ));
      expect(find.byKey(const ValueKey('bpay-status-pending')), findsOneWidget);
      expect(find.text('Check again'), findsOneWidget);
      await tester.pump(const Duration(milliseconds: 20));
      await tester.pump(const Duration(milliseconds: 20));
      await tester.pump();
      expect(api.getCalls, 2);
      expect(find.byKey(const ValueKey('bpay-status-success')), findsOneWidget);
    });

    testWidgets('check again calls the backend check endpoint only', (tester) async {
      final api = FakeBankilyApi()..checkResult = _payment('SUCCESS');
      await tester.pumpWidget(_wrap(
        BankilyPaymentStatusPage(payment: _payment('PENDING_VERIFICATION'), api: api, pollInterval: const Duration(hours: 1)),
        locale: const Locale('en'),
      ));
      await tester.tap(find.text('Check again'));
      await tester.pump();
      await tester.pump();
      expect(api.checkCalls, 1);
      expect(api.payCalls, isEmpty);
      expect(find.byKey(const ValueKey('bpay-status-success')), findsOneWidget);
      await tester.pumpWidget(const SizedBox());
    });
  });

  testWidgets('DepositPage keeps the manual flow and hides B-PAY when unavailable', (tester) async {
    await tester.pumpWidget(_wrap(const DepositPage(), locale: const Locale('ar')));
    await tester.pump();
    expect(find.byKey(const Key('bpay-entry')), findsNothing);
    expect(find.text('تعليمات الدفع'), findsOneWidget);
  });
}
