import 'dart:developer' as developer;

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
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
import 'package:mezadpay/services/fcm_service.dart';
import 'package:mezadpay/widgets/notification_handler.dart' show NotificationHandler, navigatorKey;


@pragma('vm:entry-point')
Future<void> _firebaseMessagingBackgroundHandler(RemoteMessage message) async {
   await Firebase.initializeApp();
  
  developer.log('📨 Background message received: ${message.messageId}');
  developer.log('Title: ${message.notification?.title}');
  developer.log('Body: ${message.notification?.body}');
  developer.log('Data: ${message.data}');
  
  // TODO: Enregistrer la notification localement pour affichage ultérieur
}

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
   await dotenv.load(fileName: ".env");
  
   try {
    await Firebase.initializeApp();
    developer.log('✅ Firebase initialized successfully');
  } catch (e) {
    developer.log('⚠️ Firebase initialization skipped: $e');
  }
  
   await CacheService.instance.init();
  
   try {
    final fcmService = FCMService();
    await fcmService.initialize();
    developer.log('✅ FCM Service initialized successfully');
  } catch (e) {
    developer.log('⚠️ FCM Service initialization skipped: $e');
  }
  
   FirebaseMessaging.onBackgroundMessage(_firebaseMessagingBackgroundHandler);
  
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

    return NotificationHandler(
      child: MaterialApp(
        navigatorKey: navigatorKey,
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
      ),
    );
  }
}
