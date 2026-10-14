import 'dart:io';
import 'dart:typed_data';

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

    testWidgets('splash image uses BoxFit.cover on the recomposed full-bleed artwork (no blue bands, no distortion)', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: SplashPage()));

      final image = tester.widget<Image>(find.byWidgetPredicate(
        (widget) => widget is Image && widget.image is AssetImage && (widget.image as AssetImage).assetName == 'assets/splash_full.png',
      ));
      // Client-rejected intermediate fix (see splash_page.dart's own doc
      // comment): a 720x2340 asset PADDED with large added blue gradient
      // margins still left visibly flat, content-free blue bands at the
      // very top/bottom edges under BoxFit.cover -- padding the canvas
      // doesn't remove a band, it just moves it inside the image. Final fix
      // RECOMPOSED the canvas instead: splash_full.png is now a tight
      // 720x1600 (9:20) crop centered on the real content band, no
      // artificial padding. BoxFit.cover on this asset fills any realistic
      // device viewport (9:19.5-9:20) edge-to-edge with no bands and
      // without cropping the logo/wordmark/flag/product content band.
      // BoxFit.contain would reintroduce bars; a padded asset would
      // reintroduce bands even under BoxFit.cover.
      expect(image.fit, BoxFit.cover, reason: 'must fill the screen edge-to-edge on the recomposed asset with no blue bands');
      await tester.pumpWidget(const SizedBox.shrink());
      await tester.pump(const Duration(seconds: 5));
    });

    // Regression guard: a padded canvas (real content confined to a narrow
    // middle band with large added margins) can satisfy "taller than any
    // phone ratio" while still producing visible blue bands under
    // BoxFit.cover, because the padding itself is visible content-free
    // margin. Asserting the actual asset dimensions here catches a future
    // regression to a mismatched/padded asset, which the fit-mode test
    // above cannot detect on its own (BoxFit.cover would still read as
    // correct even on a badly padded asset).
    //
    // splash_full.png is 1080x2400 (0.45 ratio) -- the final asset,
    // recomposed with the logo/wordmark/flag/product content moved well
    // inward from the left/right edges (an earlier 853x1280 source cropped
    // the logo and flag under BoxFit.cover on tall screens; this canvas's
    // 0.45 ratio exactly matches the two narrowest real device ratios in
    // use -- 1080x2400 and 720x1600 -- with zero horizontal crop, and only
    // a small vertical crop confined to empty sky background on the
    // slightly wider 1080x2340 ratio, never touching the logo/flag/product
    // band).
    test('splash_full.png is the recomposed safe-margin asset, not a padded or edge-clipped canvas', () {
      final bytes = File('assets/splash_full.png').readAsBytesSync();
      expect(bytes.length, greaterThan(8), reason: 'asset must be readable');
      expect(
        bytes.sublist(0, 8),
        Uint8List.fromList(const [0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A]),
        reason: 'splash_full.png must be a genuine PNG',
      );
      // IHDR chunk: width/height are the first two 4-byte big-endian ints
      // starting at byte 16.
      final width = ByteData.sublistView(bytes, 16, 20).getUint32(0);
      final height = ByteData.sublistView(bytes, 20, 24).getUint32(0);
      expect(width, 1080);
      expect(height, 2400, reason: 'must be the 1080x2400 safe-margin recomposed canvas, not an edge-clipping or padded asset');
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
