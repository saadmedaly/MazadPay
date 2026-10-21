import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/models/bankily_payment.dart';
import 'package:mezadpay/pages/bankily_bpay_page.dart';
import 'package:mezadpay/pages/bankily_payment_status_page.dart';
import 'package:mezadpay/services/bankily_api.dart';

const _cfg = BankilyConfig(enabled: true, merchantCode: '027922', minAmount: 100, maxAmount: 100000, allowDecimals: false);

final _pending = BankilyPayment.fromJson({
  'id': 'pending-1',
  'operation_id': 'MZP0000000000001',
  'amount': '100',
  'currency_code': 'MRU',
  'client_phone': '34****33',
  'status': 'PENDING_VERIFICATION',
});

/// Read-only fake: lists one pending payment; any attempt to pay fails the test.
class _PendingListApi extends BankilyApi {
  int payCalls = 0;
  int checkCalls = 0;

  _PendingListApi() : super(api: null);

  @override
  Future<List<BankilyPayment>> list() async => [_pending];

  @override
  Future<BankilyPayment> getPayment(String id) async => _pending;

  @override
  Future<BankilyPayment> check(String id) async {
    checkCalls++;
    return _pending;
  }

  @override
  Future<BankilyPayment> pay({required String amount, required String clientPhone, required String passcode, required String idempotencyKey}) {
    payCalls++;
    throw StateError('UI-only test must never create a payment');
  }
}

Widget _wrap(Widget child) => MaterialApp(
      locale: const Locale('ar'),
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: AppLocalizations.supportedLocales,
      home: child,
    );

void main() {
  testWidgets('pending-payment row renders inside its own Material, is tappable, and opens the pending status flow', (tester) async {
    final api = _PendingListApi();
    await tester.pumpWidget(_wrap(BankilyBpayPage(config: _cfg, api: api)));
    await tester.pump(); // let list() resolve
    await tester.pump();

    final row = find.byKey(const ValueKey('bpay-pending-pending-1'));
    expect(row, findsOneWidget);
    // 1. Renders without the ListTile/ink assertion (debug mode is on in widget tests).
    expect(tester.takeException(), isNull);
    // The tile's nearest Material ancestor is the pending banner itself.
    final material = tester.widget<Material>(find.ancestor(of: row, matching: find.byType(Material)).first);
    expect(material.color, const Color(0xFFFFF7E6));

    // 2 + 3. Tappable and opens the pending status page for that payment.
    await tester.tap(row);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 400));
    expect(find.byType(BankilyPaymentStatusPage), findsOneWidget);
    expect(find.byKey(const ValueKey('bpay-status-pending')), findsOneWidget);
    expect(
      find.descendant(of: find.byType(BankilyPaymentStatusPage), matching: find.text('MZP0000000000001')),
      findsOneWidget,
    );
    expect(tester.takeException(), isNull);

    // 4. UI-only: no payment created, nothing credited, no check triggered by the tap.
    expect(api.payCalls, 0);
    expect(api.checkCalls, 0);

    await tester.pumpWidget(const SizedBox()); // dispose the status page's poll timer
  });
}
