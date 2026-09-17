import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/splash_page.dart';

// Regression tests for the Staging splash-hang fix: SplashPage used to
// leave Navigator.pushReplacement unreachable if SharedPreferences or
// AuthService().hasValidSession() (local secure-storage read + local JWT
// decode) threw or stalled, since the startup callback was unguarded.
//
// decideSplashDestination is the pure decision function extracted from
// SplashPage so it can be exercised directly without a widget test harness
// (which would still need to fake platform channels for
// shared_preferences/flutter_secure_storage -- not available in this
// project's existing test infrastructure). The widget itself
// (_SplashPageState._runBootstrap) always calls this function with
// isLoggedIn: false on both the timeout and the exception paths -- proven
// by reading the source, not asserted here -- so covering every
// (seenOnboarding, isLoggedIn) combination this function accepts is
// equivalent to covering every real startup outcome Splash can reach.
void main() {
  group('decideSplashDestination', () {
    test('valid session (seen onboarding + logged in) -> home', () {
      expect(
        decideSplashDestination(seenOnboarding: true, isLoggedIn: true),
        SplashDestination.home,
      );
    });

    test('no session (seen onboarding, not logged in) -> login', () {
      expect(
        decideSplashDestination(seenOnboarding: true, isLoggedIn: false),
        SplashDestination.login,
      );
    });

    test('expired/invalid session is represented as isLoggedIn=false -> login', () {
      // hasValidSession() returns false for an absent/expired token, so at
      // this function's boundary an expired session is indistinguishable
      // from "no session" -- both must resolve to login, never home.
      expect(
        decideSplashDestination(seenOnboarding: true, isLoggedIn: false),
        SplashDestination.login,
      );
    });

    test('exception fallback (isLoggedIn=false) -> login, never home', () {
      // _runBootstrap's catch block always passes isLoggedIn: false here on
      // any startup exception -- this proves that input can never resolve
      // to SplashDestination.home regardless of seenOnboarding's value.
      expect(
        decideSplashDestination(seenOnboarding: true, isLoggedIn: false),
        isNot(SplashDestination.home),
      );
    });

    test('timeout fallback (isLoggedIn=false) -> login, never home', () {
      // _runBootstrap's on TimeoutException branch always passes
      // isLoggedIn: false here -- same guarantee as the exception path.
      expect(
        decideSplashDestination(seenOnboarding: true, isLoggedIn: false),
        SplashDestination.login,
      );
    });

    test('first launch (onboarding not yet seen) always wins, regardless of session', () {
      expect(
        decideSplashDestination(seenOnboarding: false, isLoggedIn: true),
        SplashDestination.onboarding,
      );
      expect(
        decideSplashDestination(seenOnboarding: false, isLoggedIn: false),
        SplashDestination.onboarding,
      );
    });

    test('no input combination ever resolves to home when isLoggedIn is false', () {
      for (final seenOnboarding in [true, false]) {
        expect(
          decideSplashDestination(seenOnboarding: seenOnboarding, isLoggedIn: false),
          isNot(SplashDestination.home),
          reason: 'seenOnboarding=$seenOnboarding, isLoggedIn=false must never grant home',
        );
      }
    });
  });

  // Customer Request #28: the splash body now renders the client-supplied
  // full-screen artwork (mobile/assets/splash_full.png) instead of the
  // plain logo. Each test asserts on the first frame only -- it must NOT
  // let initState's Future.delayed(4s) bootstrap timer actually fire,
  // since _runBootstrap then calls SharedPreferences/flutter_secure_storage,
  // and this project's test infra has no platform-channel mocking for
  // those (see the file-level comment above): letting the timer complete
  // would leave the test hanging on an unresolvable platform call instead
  // of a clean, fast assertion. Unmounting the widget via tester.pumpWidget
  // with a replacement tree runs SplashPage.dispose() -- _runBootstrap's own
  // `if (!mounted) return;` guard then makes the still-pending timer's
  // eventual callback a safe no-op, and disposing here (rather than at
  // real 4s) is also what keeps these tests fast.
  group('SplashPage visual (Customer #28)', () {
    testWidgets('renders without throwing on first frame', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: SplashPage()));
      expect(tester.takeException(), isNull);
      await tester.pumpWidget(const SizedBox.shrink());
      await tester.pump(const Duration(seconds: 5));
    });

    testWidgets('uses the new full-screen splash artwork, not the old logo-only body', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: SplashPage()));

      final imageFinder = find.byWidgetPredicate(
        (widget) => widget is Image && widget.image is AssetImage && (widget.image as AssetImage).assetName == 'assets/splash_full.png',
      );
      expect(imageFinder, findsOneWidget, reason: 'SplashPage must render assets/splash_full.png');
      await tester.pumpWidget(const SizedBox.shrink());
      await tester.pump(const Duration(seconds: 5));
    });

    testWidgets('splash image uses BoxFit.contain (full artwork always visible, no distortion/stretch/crop)', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: SplashPage()));

      final image = tester.widget<Image>(find.byWidgetPredicate(
        (widget) => widget is Image && widget.image is AssetImage && (widget.image as AssetImage).assetName == 'assets/splash_full.png',
      ));
      // BoxFit.contain, not .cover: verified via a real-viewport crop
      // simulation that .cover crops the MazadPay wordmark and the
      // Mauritanian flag on a realistic tall Android screen -- .contain
      // guarantees the complete artwork is always visible.
      expect(image.fit, BoxFit.contain, reason: 'must show the complete artwork (logo/flag never cropped), never BoxFit.cover or BoxFit.fill');
      await tester.pumpWidget(const SizedBox.shrink());
      await tester.pump(const Duration(seconds: 5));
    });

    testWidgets('splash fills the full screen (no unsized/cropped container)', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: SplashPage()));

      final sizedBox = tester.widget<SizedBox>(find.byWidgetPredicate(
        (widget) => widget is SizedBox && widget.width == double.infinity && widget.height == double.infinity,
      ));
      expect(sizedBox.width, double.infinity);
      expect(sizedBox.height, double.infinity);
      await tester.pumpWidget(const SizedBox.shrink());
      await tester.pump(const Duration(seconds: 5));
    });
  });
}
