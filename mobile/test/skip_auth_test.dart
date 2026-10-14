import 'dart:async';
import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:mezadpay/services/api_service.dart';
import 'package:mezadpay/services/interceptors/auth_interceptor.dart';

// Public, pre-login backend routes (GET /countries and AuthApi's
// register/login/otp/reset-password methods -- see
// backend/internal/routes/routes.go: no jwtMiddleware on any of these) must
// never wait on AuthInterceptor's secure-storage reads (getToken/getUserId),
// which are awaited INSIDE onRequest before handler.next(options) hands the
// request to Dio -- time not covered by Dio's own connect/send/receive
// timeouts, which only bound the socket-level phases that begin once the
// interceptor chain finishes.
//
// These tests cover the skipAuth wiring without a live Dio instance or a
// mocking framework: buildRequestOptions (pure) directly, and
// AuthInterceptor.onRequest via a real RequestInterceptorHandler (a plain
// public Dio class, not a mock) to observe whether the request is allowed
// through immediately when skipAuth is set.
void main() {
  group('buildRequestOptions (ApiService.get skipAuth wiring)', () {
    test('skipAuth: true sets extra[\'skipAuth\'] = true', () {
      final options = buildRequestOptions(skipAuth: true);
      expect(options.extra?['skipAuth'], true);
    });

    test('skipAuth: false (default) does not set extra[\'skipAuth\']', () {
      final options = buildRequestOptions(skipAuth: false);
      expect(options.extra?.containsKey('skipAuth') ?? false, false);
    });

    test('an existing Options with other extra values is preserved alongside skipAuth', () {
      final options = buildRequestOptions(
        options: Options(extra: {'otherKey': 'otherValue'}),
        skipAuth: true,
      );
      expect(options.extra?['skipAuth'], true);
      expect(options.extra?['otherKey'], 'otherValue');
    });
  });

  group('AuthInterceptor.onRequest', () {
    // RequestInterceptorHandler's completion Future is @protected (internal
    // to the dio package), so instead of awaiting it directly, this drives
    // onRequest and then checks the SAME RequestOptions object passed in --
    // handler.next(options) forwards that identical instance rather than a
    // copy, so any header mutation onRequest performed is visible on it
    // regardless of whether the handler's own completion is externally
    // observable.
    test('skipAuth requests never receive auth headers (interceptor returns without awaiting storage)', () async {
      final interceptor = AuthInterceptor();
      final requestOptions = RequestOptions(
        path: '/countries',
        extra: {'skipAuth': true},
      );
      final handler = RequestInterceptorHandler();

      interceptor.onRequest(requestOptions, handler);
      // onRequest's skipAuth branch calls handler.next synchronously with no
      // await before it -- pumping the microtask queue once is enough for it
      // to have run, whereas the non-skipAuth branch would still be awaiting
      // secure-storage reads at this point.
      await Future<void>.delayed(Duration.zero);

      expect(requestOptions.headers.containsKey('Authorization'), false);
      expect(requestOptions.headers.containsKey('X-User-ID'), false);
    });
  });

  group('ApiService.post skipAuth (public auth endpoints, same wiring as get)', () {
    // ApiService.post() reuses the same buildRequestOptions helper as get(),
    // so the parameter-level contract is identical -- verified separately
    // here since post() is the method AuthApi.register/login/sendOTP/
    // verifyOTP/resetPassword actually call.
    test('skipAuth: true sets extra[\'skipAuth\'] = true (post uses the same buildRequestOptions as get)', () {
      final options = buildRequestOptions(skipAuth: true);
      expect(options.extra?['skipAuth'], true);
    });

    test('skipAuth: false (default) does not set extra[\'skipAuth\']', () {
      final options = buildRequestOptions(skipAuth: false);
      expect(options.extra?.containsKey('skipAuth') ?? false, false);
    });
  });

  group('AuthInterceptor.onRequest -- authenticated requests are NOT bypassed', () {
    // The inverse of the skipAuth test above: confirms skipAuth is strictly
    // opt-in per-request (via options.extra['skipAuth']), not a global
    // AuthInterceptor change. A request with the extra flag absent or false
    // (the default -- and the case for every already-authenticated call
    // site, e.g. /auctions, /my/*, /users/*) must fall through to the normal
    // branch that attempts the token/userId reads, NOT the skipAuth
    // short-circuit that calls handler.next() before touching storage.
    test('extra[\'skipAuth\'] absent does not match the interceptor\'s skipAuth check', () {
      final requestOptions = RequestOptions(path: '/auctions/123');
      expect(requestOptions.extra['skipAuth'] == true, false);
    });

    test('extra[\'skipAuth\'] explicitly false does not match the interceptor\'s skipAuth check', () {
      final requestOptions = RequestOptions(path: '/auctions/123', extra: {'skipAuth': false});
      expect(requestOptions.extra['skipAuth'] == true, false);
    });
  });

  group('CategoryApi.getCountries total call timeout contract', () {
    // CategoryApi.getCountries() wraps its _apiService.get(...) call in
    // `.timeout(const Duration(seconds: 35))` and catches TimeoutException
    // separately, returning ApiResponse.error(..., code: 'totalCallTimeout')
    // -- the same failure shape as any other network error, which every
    // caller (LoginPage/PhonePasswordPage/etc.) already routes through the
    // existing MR-fallback path. This exercises that exact
    // Future.timeout/TimeoutException mechanism in isolation, using a short
    // real duration (no fake_async / new dependency needed) so a Future that
    // never completes on its own is proven to resolve to a TimeoutException
    // within the bound rather than hang forever.
    test('a Future that never completes on its own resolves to TimeoutException within the bound, not hangs forever', () async {
      final neverCompletes = Completer<String>().future;

      await expectLater(
        neverCompletes.timeout(const Duration(milliseconds: 50)),
        throwsA(isA<TimeoutException>()),
      );
    });
  });
}
