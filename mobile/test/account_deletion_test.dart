import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/pages/account_profile_page.dart';
import 'package:mezadpay/services/api_service.dart';
import 'package:mezadpay/widgets/account_deletion_flow.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Fakes the backend for AccountProfilePage + DELETE /users/me and records
/// every request, so no test ever reaches a real server.
class _FakeBackend extends Interceptor {
  final List<String> calls = [];
  int deleteStatus = 200;
  String deleteCode = '';

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    calls.add('${options.method} ${options.path}');
    if (options.method == 'GET' && options.path.endsWith('/users/me')) {
      return handler.resolve(Response(
        requestOptions: options,
        statusCode: 200,
        data: jsonEncode({
          'success': true,
          'data': {'id': 'u1', 'full_name': 'Review Demo', 'phone': '+22200000000', 'notifications_enabled': true},
        }),
      ));
    }
    if (options.method == 'DELETE' && options.path.endsWith('/users/me')) {
      if (deleteStatus == 200) {
        return handler.resolve(Response(
          requestOptions: options,
          statusCode: 200,
          data: {'success': true, 'data': {'message': 'Account deletion completed'}}, // DELETE: ApiService returns Dio's decoded map
        ));
      }
      return handler.reject(DioException(
        requestOptions: options,
        type: DioExceptionType.badResponse,
        response: Response(
          requestOptions: options,
          statusCode: deleteStatus,
          data: {'success': false, 'error': {'code': deleteCode, 'message': 'blocked'}},
        ),
      ));
    }
    // Anything else the page fires (notification prefs, etc.): harmless empty success.
    handler.resolve(Response(requestOptions: options, statusCode: 200, data: jsonEncode({'success': true, 'data': {}})));
  }

  int get deleteCalls => calls.where((c) => c == 'DELETE /users/me').length;
}

Widget _app(Widget home) => ProviderScope(
      child: MaterialApp(
        locale: const Locale('ar'),
        localizationsDelegates: const [
          AppLocalizations.delegate,
          GlobalMaterialLocalizations.delegate,
          GlobalWidgetsLocalizations.delegate,
          GlobalCupertinoLocalizations.delegate,
        ],
        supportedLocales: AppLocalizations.supportedLocales,
        home: home,
      ),
    );

Future<void> settle(WidgetTester tester) async {
  for (var i = 0; i < 10; i++) {
    await tester.pump(const Duration(milliseconds: 100));
    await tester.runAsync(() => Future<void>.delayed(const Duration(milliseconds: 20)));
  }
  await tester.pump();
}

void main() {
  late _FakeBackend backend;
  late int clearCalls;

  AccountDeletionFlow flow() => AccountDeletionFlow(
        clearLocalSession: () async => clearCalls++,
        loginBuilder: (_) => const Scaffold(body: Text('LOGIN_SCREEN')),
      );

  setUpAll(() {
    dotenv.testLoad(fileInput: 'API_BASE_URL=http://fake.test/v1/api\nWS_BASE_URL=ws://fake.test/v1/api');
    backend = _FakeBackend();
    ApiService().dio.interceptors.insert(0, backend);
  });

  setUp(() {
    backend.calls.clear();
    backend.deleteStatus = 200;
    backend.deleteCode = '';
    clearCalls = 0;
    FlutterSecureStorage.setMockInitialValues({'jwt_token': 't', 'refresh_token': 'r', 'user_id': 'u1'});
    SharedPreferences.setMockInitialValues({});
  });

  Future<void> openDeleteDialog(WidgetTester tester) async {
    tester.view.physicalSize = const Size(1179, 2556);
    tester.view.devicePixelRatio = 3;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(_app(AccountProfilePage(deletionFlow: flow())));
    await settle(tester);
    final tile = find.byKey(const Key('delete-account-tile'));
    await tester.ensureVisible(tile);
    await tester.pumpAndSettle();
    await tester.tap(tile);
    await tester.pumpAndSettle();
  }

  testWidgets('delete item is visible in AccountProfilePage, above Logout', (tester) async {
    tester.view.physicalSize = const Size(1179, 2556);
    tester.view.devicePixelRatio = 3;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(_app(AccountProfilePage(deletionFlow: flow())));
    await settle(tester);
    final tile = find.byKey(const Key('delete-account-tile'));
    await tester.ensureVisible(tile);
    await tester.pumpAndSettle();
    expect(tile, findsOneWidget);
    expect(find.descendant(of: tile, matching: find.text('حذف الحساب')), findsOneWidget);
    final logout = find.text('تسجيل الخروج');
    expect(logout, findsOneWidget);
    expect(tester.getTopLeft(tile).dy, lessThan(tester.getTopLeft(logout).dy));
  });

  testWidgets('cancel does nothing', (tester) async {
    await openDeleteDialog(tester);
    expect(find.byKey(const Key('delete-account-confirm')), findsOneWidget);

    await tester.tap(find.byKey(const Key('delete-account-cancel')));
    await settle(tester);

    expect(backend.deleteCalls, 0);
    expect(clearCalls, 0);
    expect(find.text('LOGIN_SCREEN'), findsNothing);
    expect(find.byType(AccountProfilePage), findsOneWidget);
  });

  testWidgets('confirm calls DELETE /users/me, clears the session and lands on Login with no way back', (tester) async {
    await openDeleteDialog(tester);
    await tester.tap(find.byKey(const Key('delete-account-confirm')));
    await settle(tester);
    await tester.pumpAndSettle();

    expect(backend.deleteCalls, 1);
    expect(clearCalls, 1);
    expect(find.text('LOGIN_SCREEN'), findsOneWidget);
    expect(find.text('تم حذف حسابك بنجاح'), findsOneWidget);
    expect(find.byType(AccountProfilePage), findsNothing);
    final navigator = tester.state<NavigatorState>(find.byType(Navigator).first);
    expect(navigator.canPop(), isFalse, reason: 'back must not return to authenticated pages');
  });

  testWidgets('blocked deletion shows the reason and keeps the user signed in', (tester) async {
    backend.deleteStatus = 409;
    backend.deleteCode = 'account_deletion_blocked_balance';
    await openDeleteDialog(tester);
    await tester.tap(find.byKey(const Key('delete-account-confirm')));
    await settle(tester);
    await tester.pumpAndSettle();

    expect(backend.deleteCalls, 1);
    expect(find.text('لا يمكن حذف الحساب الآن'), findsOneWidget);
    expect(find.text('لديك رصيد في المحفظة. يرجى سحب رصيدك أولاً ثم حذف الحساب.'), findsOneWidget);
    expect(clearCalls, 0);
    expect(find.text('LOGIN_SCREEN'), findsNothing);
  });

  test('every backend blocker code has a localized message', () async {
    for (final locale in AppLocalizations.supportedLocales) {
      final l10n = await AppLocalizations.delegate.load(locale);
      for (final code in [
        'account_deletion_blocked_balance',
        'account_deletion_blocked_hold',
        'account_deletion_blocked_pending_deposit',
        'account_deletion_blocked_pending_withdrawal',
        'account_deletion_blocked_leading_bid',
        'account_deletion_blocked_unpaid_win',
        'account_deletion_blocked_active_listing',
      ]) {
        expect(AccountDeletionFlow.blockedReason(l10n, code), isNotEmpty, reason: '$locale $code');
      }
      expect(AccountDeletionFlow.blockedReason(l10n, 'server_error'), isNull);
    }
  });

  test('clearLocalAccountData wipes secure storage and account prefs, keeps device prefs', () async {
    FlutterSecureStorage.setMockInitialValues({'jwt_token': 't', 'refresh_token': 'r', 'user_id': 'u1', 'has_registered': 'true'});
    SharedPreferences.setMockInitialValues({
      'fcm_token': 'tok',
      'device_id': 'dev',
      'cached_location': '{}',
      'local_favorites': <String>['a'],
      'favorites_pending_sync': <String>['a'],
      'onboarding_seen': true,
    });

    await clearLocalAccountData();

    expect(await const FlutterSecureStorage().readAll(), isEmpty);
    final prefs = await SharedPreferences.getInstance();
    for (final key in accountScopedPrefKeys) {
      expect(prefs.containsKey(key), isFalse, reason: key);
    }
    expect(prefs.getBool('onboarding_seen'), isTrue);
  });
}
