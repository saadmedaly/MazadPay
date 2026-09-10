import 'dart:async';

import 'package:flutter/material.dart';
import 'package:mezadpay/pages/language_page.dart';
import 'package:mezadpay/pages/start_bidding_page.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/services/auth_service.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../widgets/mazad_pay_logo.dart';

/// Bounded wait for the LOCAL-only session/onboarding bootstrap
/// (SharedPreferences + flutter_secure_storage/Android Keystore read + local
/// JWT decode -- no network call is made here). Generous relative to how
/// fast local storage normally resolves, but still finite: if the platform
/// channel stalls, this timeout is what guarantees Splash still leaves
/// within a bounded time instead of hanging forever.
const Duration kSplashBootstrapTimeout = Duration(seconds: 5);

/// Where Splash should navigate. Never includes a "stay on Splash" case --
/// every startup outcome (success, no session, exception, timeout) maps to
/// exactly one of these.
enum SplashDestination { onboarding, home, login }

/// Pure decision logic, isolated from Navigator/BuildContext so it is
/// directly unit-testable without a widget test harness. Given the raw
/// onboarding/session inputs, decides where Splash sends the user. Failing
/// closed is the whole point: `isLoggedIn: false` (the value used for both
/// the timeout and the exception paths in [SplashPage]) always resolves to
/// `login` here, never `home` -- there is no branch that can turn a failure
/// into an authenticated destination.
SplashDestination decideSplashDestination({
  required bool seenOnboarding,
  required bool isLoggedIn,
}) {
  if (!seenOnboarding) return SplashDestination.onboarding;
  return isLoggedIn ? SplashDestination.home : SplashDestination.login;
}

/// Loads the local-only onboarding flag + session validity. No network call
/// -- see [AuthService.hasValidSession].
Future<({bool seenOnboarding, bool isLoggedIn})> loadLocalSplashState() async {
  final prefs = await SharedPreferences.getInstance();
  final seen = prefs.getBool('onboarding_seen') ?? false;
  final isLoggedIn = await AuthService().hasValidSession();
  return (seenOnboarding: seen, isLoggedIn: isLoggedIn);
}

class SplashPage extends StatefulWidget {
  const SplashPage({super.key});

  @override
  State<SplashPage> createState() => _SplashPageState();
}

class _SplashPageState extends State<SplashPage> {
  bool _navigated = false;

  @override
  void initState() {
    super.initState();
    debugPrint('[Splash] startup begin');
    Future.delayed(const Duration(seconds: 4), _runBootstrap);
  }

  Future<void> _runBootstrap() async {
    if (!mounted) return;

    // Splash startup must NEVER depend on the backend being reachable --
    // loadLocalSplashState() only reads local secure storage and decodes
    // the JWT payload locally (no HTTP call), so no separate network guard
    // is needed here; the timeout below exists purely for local-storage /
    // platform-channel stalls (SharedPreferences, secure storage, Android
    // Keystore), not for network latency.
    bool seenOnboarding = false;
    bool isLoggedIn = false;

    try {
      debugPrint('[Splash] local session check begin');
      final result = await loadLocalSplashState().timeout(kSplashBootstrapTimeout);
      seenOnboarding = result.seenOnboarding;
      isLoggedIn = result.isLoggedIn;
      debugPrint(isLoggedIn ? '[Splash] session valid' : '[Splash] session unavailable');
    } on TimeoutException {
      debugPrint('[Splash] startup timeout');
      // Never assume authenticated on timeout -- fail closed to "no session".
      seenOnboarding = true;
      isLoggedIn = false;
    } catch (e) {
      debugPrint('[Splash] startup exception: ${e.runtimeType}');
      // Never assume authenticated on any startup exception -- fail closed.
      seenOnboarding = true;
      isLoggedIn = false;
    }

    _navigate(decideSplashDestination(seenOnboarding: seenOnboarding, isLoggedIn: isLoggedIn));
  }

  void _navigate(SplashDestination destination) {
    // Guards against both post-dispose navigation and a double-navigation
    // race (bootstrap is only ever triggered once per Splash instance, but
    // this keeps the guarantee explicit and cheap).
    if (!mounted || _navigated) return;
    _navigated = true;

    Widget page;
    switch (destination) {
      case SplashDestination.onboarding:
        page = const LanguagePage();
        debugPrint('[Splash] navigating -> onboarding');
        break;
      case SplashDestination.home:
        page = const HomePage();
        debugPrint('[Splash] navigating -> home');
        break;
      case SplashDestination.login:
        page = StartBiddingPage();
        debugPrint('[Splash] navigating -> login');
        break;
    }

    Navigator.of(context).pushReplacement(
      MaterialPageRoute(builder: (_) => page),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFFBFBFB),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [const MazadPayLogo(fontSize: 64, arabicFontSize: 32)],
        ),
      ),
    );
  }
}
