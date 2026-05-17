import 'package:flutter/material.dart';
import 'package:mezadpay/pages/language_page.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/pages/login_page.dart';
import '../services/auth_service.dart';
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
    _checkRouting();
  }

  Future<void> _checkRouting() async {
    // Wait for splash screen minimum delay
    await Future.delayed(const Duration(seconds: 4));
    
    if (!mounted) return;
    
    try {
      final authService = AuthService();
      final loggedIn = await authService.isLoggedIn();
      
      if (loggedIn) {
        Navigator.of(context).pushReplacement(
          MaterialPageRoute(builder: (_) => const HomePage()),
        );
      } else {
        final registered = await authService.hasRegistered();
        if (registered) {
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(builder: (_) => const LoginPage()),
          );
        } else {
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(builder: (_) => const LanguagePage()),
          );
        }
      }
    } catch (e) {
      // Clean fallback in case of storage issues
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(builder: (_) => const LanguagePage()),
      );
    }
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
