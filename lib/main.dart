import 'package:flutter/material.dart';
import 'package:mezadpay/core/theme.dart';
import 'package:mezadpay/pages/splash_page.dart';

void main() {
  runApp(const MazadApp());
}

class MazadApp extends StatelessWidget {
  const MazadApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'MazadPay',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      themeMode: ThemeMode.light,
      home: const SplashPage(),
    );
  }
}
