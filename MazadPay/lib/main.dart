import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/core/theme.dart';
import 'package:mezadpay/pages/splash_page.dart';
import 'package:google_fonts/google_fonts.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  
  // Optional: Disable runtime font fetching if there are persistive issues.
  GoogleFonts.config.allowRuntimeFetching = true;

  runApp(
    const ProviderScope(
      child: MazadApp(),
    ),
  );
}

class MazadApp extends StatelessWidget {
  const MazadApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'MazadPay',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme.copyWith(
        textTheme: GoogleFonts.poppinsTextTheme(
          ThemeData.light().textTheme,
        ),
      ),
      darkTheme: AppTheme.darkTheme.copyWith(
        textTheme: GoogleFonts.poppinsTextTheme(
          ThemeData.dark().textTheme,
        ),
      ),
      themeMode: ThemeMode.light,
      home: const SplashPage(),
    );
  }
}
