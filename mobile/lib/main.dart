import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/constants/api_constants.dart';
import 'presentation/pages/splash_page.dart';
import 'services/notification_service.dart';

// Global key for navigation from background notifications
final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  
  // Initialize Firebase
  await Firebase.initializeApp();
  
  // Set up background message handler
  FirebaseMessaging.onBackgroundMessage(firebaseMessagingBackgroundHandler);
  
  runApp(
    const ProviderScope(
      child: MazadPayApp(),
    ),
  );
}

class MazadPayApp extends ConsumerStatefulWidget {
  const MazadPayApp({super.key});

  @override
  ConsumerState<MazadPayApp> createState() => _MazadPayAppState();
}

class _MazadPayAppState extends ConsumerState<MazadPayApp> {
  @override
  void initState() {
    super.initState();
    _initNotifications();
  }

  Future<void> _initNotifications() async {
    final notificationService = ref.read(notificationServiceProvider);
    await notificationService.initialize();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'MazadPay',
      debugShowCheckedModeBanner: false,
      navigatorKey: navigatorKey,
      
      // Theme - Adaptable au design existant
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF2E7D32), // Green - couleur MazadPay
          brightness: Brightness.light,
        ),
        fontFamily: 'Cairo', // Police arabe
        appBarTheme: const AppBarTheme(
          centerTitle: true,
          elevation: 0,
        ),
        cardTheme: CardTheme(
          elevation: 2,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ),
      
      // Internationalization
      localizationsDelegates: const [
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [
        Locale('ar'), // Arabe (par défaut)
        Locale('fr'), // Français
        Locale('en'), // Anglais
      ],
      locale: const Locale('ar'),
      
      // RTL pour l'arabe
      builder: (context, child) {
        return Directionality(
          textDirection: TextDirection.rtl,
          child: child!,
        );
      },
      
      home: const SplashPage(),
    );
  }
}
