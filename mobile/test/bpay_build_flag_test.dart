import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/core/feature_flags.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/pages/deposit_page.dart';
import 'package:mezadpay/services/api_service.dart';

/// Answers the requests DepositPage makes as if the backend had B-PAY ENABLED,
/// and records which paths were requested.
class _FakeBackend extends Interceptor {
  final List<String> paths = [];

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    paths.add(options.path);
    if (options.path.endsWith('/users/wallet/bankily/config')) {
      return handler.resolve(Response(
        requestOptions: options,
        statusCode: 200,
        data: '{"success":true,"data":{"enabled":true,"merchant_code":"027922","min_amount":100,"max_amount":100000,"allow_decimals":false}}',
      ));
    }
    if (options.path.endsWith('/users/wallet/payment-methods')) {
      return handler.resolve(Response(
        requestOptions: options,
        statusCode: 200,
        data: '{"success":true,"data":[{"id":10,"code":"sada","name_ar":"سداد","name_fr":"Sada","name_en":"Sada","is_active":true}]}',
      ));
    }
    handler.reject(DioException(requestOptions: options, type: DioExceptionType.connectionError));
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

/// Lets the fake HTTP futures complete on the real clock, then rebuilds.
Future<void> settle(WidgetTester tester) async {
  for (var i = 0; i < 10; i++) {
    await tester.pump(const Duration(milliseconds: 100)); // advance the fake clock (Dio's internal timers)
    await tester.runAsync(() => Future<void>.delayed(const Duration(milliseconds: 20)));
  }
  await tester.pump();
}

void main() {
  late _FakeBackend backend;

  setUpAll(() {
    dotenv.testLoad(fileInput: 'API_BASE_URL=http://fake.test/v1/api\nWS_BASE_URL=ws://fake.test/v1/api');
    backend = _FakeBackend();
    // First in the chain, so no request reaches secure storage or the network.
    ApiService().dio.interceptors.insert(0, backend);
  });

  setUp(() => backend.paths.clear());

  test('B-PAY is OFF by default (no BPAY_ENABLED dart-define)', () {
    expect(kBpayEnabled, isFalse);
  });

  testWidgets('flag OFF: no B-PAY card even though the backend reports B-PAY enabled; manual flow intact', (tester) async {
    await tester.pumpWidget(_wrap(const DepositPage(bpayFeatureEnabled: false)));
    await settle(tester);

    expect(find.byKey(const Key('bpay-entry')), findsNothing);
    expect(backend.paths.where((p) => p.contains('bankily')), isEmpty, reason: 'B-PAY config must not even be requested');
    // Manual deposit UI: instructions box and the Production manual method.
    expect(find.text('تعليمات الدفع'), findsOneWidget);
    expect(find.text('سداد'), findsOneWidget);
    expect(find.text('إتمام الدفع'), findsOneWidget);
  });

  testWidgets('flag ON (contrast): the same backend answer shows the B-PAY card', (tester) async {
    await tester.pumpWidget(_wrap(const DepositPage(bpayFeatureEnabled: true)));
    await settle(tester);

    expect(find.byKey(const Key('bpay-entry')), findsOneWidget);
    expect(find.text('تعليمات الدفع'), findsOneWidget);
  });
}
