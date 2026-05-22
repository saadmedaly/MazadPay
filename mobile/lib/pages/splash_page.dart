import 'package:flutter/material.dart';
import 'package:mezadpay/pages/language_page.dart';
import 'package:mezadpay/pages/start_bidding_page.dart';
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
      if (!mounted) return;
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(
          builder: (_) => seen ? StartBiddingPage() : const LanguagePage(),
        ),
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
