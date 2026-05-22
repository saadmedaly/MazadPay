import 'package:flutter/material.dart';
import 'package:mezadpay/pages/language_page.dart';
import 'package:mezadpay/pages/start_bidding_page.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/services/auth_service.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../widgets/mazad_pay_logo.dart';

class SplashPage extends StatefulWidget {
  const SplashPage({super.key});

  @override
  State<SplashPage> createState() => _SplashPageState();
}

class _SplashPageState extends State<SplashPage> {
  @override
  void initState() {
    super.initState();
    Future.delayed(const Duration(seconds: 4), () async {
      if (!mounted) return;
      final prefs = await SharedPreferences.getInstance();
      final seen = prefs.getBool('onboarding_seen') ?? false;
      final isLoggedIn = await AuthService().isLoggedIn();
      if (!mounted) return;

      Widget destination;
      if (!seen) {
        // First time: show onboarding
        destination = const LanguagePage();
      } else if (isLoggedIn) {
        // Seen onboarding + has valid token: go directly to home
        destination = const HomePage();
      } else {
        // Seen onboarding but not logged in: show start/login screen
        destination = StartBiddingPage();
      }

      Navigator.of(context).pushReplacement(
        MaterialPageRoute(builder: (_) => destination),
      );
    });
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
