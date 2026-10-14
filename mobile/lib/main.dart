import 'dart:async';
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
import 'package:mezadpay/services/realtime_sync_service.dart';
import 'package:mezadpay/widgets/notification_handler.dart' show NotificationHandler, navigatorKey;


@pragma('vm:entry-point')
Future<void> _firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp();
  developer.log('📨 Background message received: ${message.messageId}');
  developer.log('Title: ${message.notification?.title}');
  developer.log('Body: ${message.notification?.body}');
  developer.log('Data: ${message.data}');
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

class _MazadAppState extends ConsumerState<MazadApp> with WidgetsBindingObserver {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    Future.microtask(() async {
      try {
        final fcmService = FCMService();
        await fcmService.initialize();
        developer.log('✅ FCM Service initialized successfully');
      } catch (e) {
        developer.log('⚠️ FCM Service initialization skipped: $e');
      }
      ref.read(locationProvider.notifier).detectLocation();
      // Customer #20: opens the app-wide global WebSocket (list/content
      // invalidation events) once at startup. A no-op if the user isn't
      // logged in yet (GlobalWebsocketService.connect() checks for a token
      // and simply returns if there is none) -- the login flow doesn't
      // currently call this again, but didChangeAppLifecycleState's resume
      // handler below will pick it up the next time the app is foregrounded
      // after login, and connect() is always safe to call again.
      unawaited(RealtimeSyncService().start());
    });
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    super.didChangeAppLifecycleState(state);
    // Customer #20, Phase 6: automatic catch-up on resume, no manual
    // refresh required. RealtimeSyncService.onAppResumed() both fires the
    // catch-up signal immediately and reconnects the global WS if it had
    // dropped while backgrounded.
    if (state == AppLifecycleState.resumed) {
      unawaited(RealtimeSyncService().onAppResumed());
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
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
