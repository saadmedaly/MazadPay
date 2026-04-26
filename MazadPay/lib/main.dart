import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:mezadpay/core/theme.dart';
import 'package:mezadpay/pages/splash_page.dart';

import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/providers/locale_provider.dart';
import 'package:mezadpay/providers/location_provider.dart';
import 'package:mezadpay/services/cache_service.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  // Charger les variables d'environnement
  await dotenv.load(fileName: ".env");
  
  // Initialiser le cache local (Hive)
  await CacheService.instance.init();
  
  runApp(
    const ProviderScope(
      child: MazadApp(),
    ),
  );
}

class MazadApp extends ConsumerStatefulWidget {
  const MazadApp({super.key});

  @override
  ConsumerState<MazadApp> createState() => _MazadAppState();
}

class _MazadAppState extends ConsumerState<MazadApp> {
  @override
  void initState() {
    super.initState();
    // Démarrer la détection de localisation en arrière-plan
    Future.microtask(() {
      ref.read(locationProvider.notifier).detectLocation();
    });
  }

  @override
  Widget build(BuildContext context) {
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
