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
}
