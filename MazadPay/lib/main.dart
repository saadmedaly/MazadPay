import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/core/theme.dart';
import 'package:mezadpay/pages/splash_page.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/providers/locale_provider.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  
  // Disable runtime font fetching for offline use.
  GoogleFonts.config.allowRuntimeFetching = false;

  runApp(
    const ProviderScope(
      child: MazadApp(),
    ),
  );
}

class MazadApp extends ConsumerWidget {
  const MazadApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final currentLocale = ref.watch(localeNotifierProvider);

    return MaterialApp(
      title: 'MazadPay',
      locale: currentLocale,
      localizationsDelegates: [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [
        Locale('ar'),
        Locale('fr'),
        Locale('en'),
      ],
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      themeMode: ThemeMode.light,
      home: const SplashPage(),
    );
  }
}