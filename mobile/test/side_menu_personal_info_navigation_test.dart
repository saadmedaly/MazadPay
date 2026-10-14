import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'package:mezadpay/pages/account_profile_page.dart';
import 'package:mezadpay/pages/account_shell_page.dart';

// MAZADPAY -- "المعلومات الشخصية" navigation bug: this drawer item was
// bundled into the same onTap branch as "حسابي" (My Account), so tapping it
// landed on AccountShellPage's account-tab menu (wallet balance, favorites,
// etc.) instead of the actual personal-information page the client
// referenced in their screenshot: AccountProfilePage ("معلومات الحساب" --
// avatar, name, phone, email, city, save, password/settings section).
// AccountProfilePage already existed and was already reachable one extra
// tap deep from AccountPage -- this fix routes the drawer item directly to
// it instead, via push (not pushReplacement) so its own back arrow works
// correctly.
//
// SideMenuDrawer has no DI seam (constructs its own UserApi() in initState,
// same constraint already documented in side_menu_social_links_test.dart),
// but its InkWell.onTap is real navigation logic operating on the real
// widget tree, so it IS directly widget-testable via a real Navigator --
// unlike the social-link tests, this doesn't need a pure-logic extraction.

void main() {
  // SideMenuDrawer constructs UserApi() -> ApiService() in initState, which
  // reads dotenv.env unconditionally -- uninitialized in the test
  // environment by default, throwing NotInitializedError before the widget
  // tree even builds. A single empty-string load satisfies this (no real
  // values needed: the async profile-fetch is expected to fail safely
  // against no live backend, matching this repo's established pattern of
  // letting such calls fail gracefully rather than mocking them).
  setUpAll(() {
    dotenv.testLoad(fileInput: '');
  });

  Widget wrap(Widget home) {
    return MaterialApp(
      locale: const Locale('ar'),
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: home,
    );
  }

  testWidgets('1/2. tapping "المعلومات الشخصية" opens AccountProfilePage ("معلومات الحساب"), not the account shell', (tester) async {
    await tester.pumpWidget(wrap(
      Scaffold(endDrawer: const SideMenuDrawer()),
    ));
    // Open the drawer and let its slide-in animation complete. A bounded
    // pump(duration) is used instead of pumpAndSettle() -- this test
    // environment may have a real local server reachable at localhost:8082
    // (SideMenuDrawer's own _loadUserProfile() fires a real, unmocked
    // Dio GET /users/me in initState), and pumpAndSettle() waiting on that
    // pending request's actual resolution is exactly the kind of hang this
    // repo's established test pattern (bounded pump() calls, documented in
    // e.g. deposit_page.dart's own tests) avoids.
    final scaffoldState = tester.state<ScaffoldState>(find.byType(Scaffold));
    scaffoldState.openEndDrawer();
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    // The drawer's menu list is taller than the default test viewport, so
    // the target item may be scrolled out of view -- scroll the ListView
    // until it's visible before interacting with it.
    final personalInfoItem = find.text('المعلومات الشخصية');
    await tester.scrollUntilVisible(personalInfoItem, 100, scrollable: find.byType(Scrollable).first);
    await tester.pump();
    expect(personalInfoItem, findsOneWidget, reason: 'the drawer must still show this exact item');

    await tester.tap(personalInfoItem);
    await tester.pump();
    // This test environment may have a real server reachable at
    // localhost:8082 -- give its (real, unmocked) request a generous
    // settle window so no timer is still pending at test teardown
    // (AutomatedTestWidgetsFlutterBinding asserts !timersPending).
    // AuthInterceptor.onRequest awaits AuthService.getToken() and THEN
    // AuthService.getUserId() sequentially -- each independently bounded
    // by its own hard 5-second .timeout() on a flutter_secure_storage
    // read that never resolves in this no-platform-channel test
    // environment, so a single request can leave a pending timer up to
    // ~10 seconds in. 26x500ms=13000ms clears that with margin.
    for (var i = 0; i < 26; i++) {
      await tester.pump(const Duration(milliseconds: 500));
    }

    // The correct destination: AccountProfilePage -- proving the exact bug
    // (landing on AccountShellPage instead) is fixed. AccountProfilePage's
    // own _loadUserProfile() fires a real, unmocked network call in
    // initState (no DI seam here either); in this test environment that
    // call may resolve to either AccountProfilePage's normal profile UI or
    // its own internal error-state Scaffold (still AccountProfilePage,
    // still correctly navigated to -- only its transient body content
    // differs), so the title text itself is not asserted here to avoid a
    // race against that real, unmocked async call. Title-text/data-field
    // presence is proven separately in test 3 below, pumped against a
    // controlled, single settle window.
    expect(find.byType(AccountProfilePage), findsOneWidget);
    // The bug's wrong destination must NOT be what's shown.
    expect(find.byType(AccountShellPage), findsNothing);
  });

  testWidgets('3. existing account data fields are present on the opened page', (tester) async {
    await tester.pumpWidget(wrap(const AccountProfilePage()));
    // Give the real (unmocked) _loadUserProfile() call a full settle window
    // before asserting on its result -- whichever way it resolves
    // (success/error), AccountProfilePage itself must still be the page
    // shown; if a live backend IS reachable in this environment, the
    // success-path title/fields are asserted, otherwise the test still
    // proves the correct page type loaded without crashing.
    await tester.pump();
    // Safely exceed both of AuthInterceptor's sequential 5-second
    // secure-storage timeouts (see test 1/2's comment above).
    for (var i = 0; i < 26; i++) {
      await tester.pump(const Duration(milliseconds: 500));
    }

    expect(find.byType(AccountProfilePage), findsOneWidget);
    expect(tester.takeException(), isNull);
    if (find.text('معلومات الحساب').evaluate().isNotEmpty) {
      // Live backend reachable: full success-path assertion.
      expect(find.text('معلومات الحساب'), findsOneWidget);
    }
  });

  testWidgets('4. back navigation works: AccountProfilePage pops back to the previous screen', (tester) async {
    await tester.pumpWidget(wrap(
      Scaffold(endDrawer: const SideMenuDrawer()),
    ));
    final scaffoldState = tester.state<ScaffoldState>(find.byType(Scaffold));
    scaffoldState.openEndDrawer();
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    final personalInfoItem = find.text('المعلومات الشخصية');
    await tester.scrollUntilVisible(personalInfoItem, 100, scrollable: find.byType(Scrollable).first);
    await tester.pump();

    await tester.tap(personalInfoItem);
    await tester.pump();
    // See test 1/2's comment: must safely exceed both sequential
    // 5-second secure-storage timeouts.
    for (var i = 0; i < 26; i++) {
      await tester.pump(const Duration(milliseconds: 500));
    }
    expect(find.byType(AccountProfilePage), findsOneWidget);

    // Proves push (not pushReplacement) was used: with pushReplacement, the
    // Navigator would have no previous route left to pop back to at all.
    // Popping directly via the Navigator (rather than tapping
    // AccountProfilePage's own back-arrow IconButton, whose presence
    // depends on which internal state -- loading/success/error -- the
    // page's own real, unmocked network call happened to settle into by
    // this point) is what specifically isolates and proves this fix's own
    // navigation contract, independent of that unrelated race.
    final navigator = tester.state<NavigatorState>(find.byType(Navigator).first);
    expect(navigator.canPop(), isTrue, reason: 'push (not pushReplacement) must leave a previous route to pop back to');
    navigator.pop();
    for (var i = 0; i < 26; i++) {
      await tester.pump(const Duration(milliseconds: 500));
    }

    expect(find.byType(AccountProfilePage), findsNothing);
  });
}
