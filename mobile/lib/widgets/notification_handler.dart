import 'dart:async';
import 'dart:developer' as developer;

import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/pages/auction_details_page.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/pages/deposit_page.dart';
import 'package:mezadpay/services/fcm_service.dart';

/// Global key pour accéder au Navigator depuis n'importe où
final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();

/// Widget qui gère la navigation depuis les notifications FCM
/// Placez ce widget au sommet de l'arborescence de navigation
class NotificationHandler extends ConsumerStatefulWidget {
  final Widget child;

  const NotificationHandler({
    Key? key,
    required this.child,
  }) : super(key: key);

  @override
  ConsumerState<NotificationHandler> createState() => _NotificationHandlerState();
}

class _NotificationHandlerState extends ConsumerState<NotificationHandler> {
  FCMService? _fcmService;
  StreamSubscription? _notificationSubscription;

  @override
  void initState() {
    super.initState();
    _initializeFCM();
  }
  
  Future<void> _initializeFCM() async {
    // Skip FCM sur web (non supporté sans configuration spécifique)
    if (kIsWeb) {
      developer.log('ℹ️ FCM not available on web, skipping initialization');
      return;
    }
    
    try {
      // Vérifier que Firebase est initialisé avant d'accéder au FCMService
      Firebase.app();
      
      _fcmService = FCMService();
      
      // Configurer le callback de navigation
      _fcmService!.onNotificationTap = _handleNotificationTap;
      
      // Écouter les notifications en temps réel
      _notificationSubscription = _fcmService!.notificationStream.listen(_handleNotification);
      
      developer.log('✅ NotificationHandler initialized');
    } catch (e) {
      developer.log('⚠️ Firebase not initialized yet, will retry...');
      // Réessayer après un délai
      Future.delayed(const Duration(seconds: 2), () {
        if (mounted) _initializeFCM();
      });
    }
  }

  @override
  void dispose() {
    _notificationSubscription?.cancel();
    _fcmService?.onNotificationTap = null;
    super.dispose();
  }

  /// Handler pour le tap sur une notification
  void _handleNotificationTap(Map<String, dynamic> data) {
    _navigateFromNotification(data);
  }

  /// Handler pour les notifications reçues
  void _handleNotification(Map<String, dynamic> data) {
    // Optionnel: Afficher un badge ou une alerte
    final String? type = data['type'];
    if (type == 'auction_ending_soon' || type == 'auction_reported') {
      // Ces notifications sont urgentes
      _navigateFromNotification(data);
    }
  }

  /// Navigation vers la bonne page selon la notification
  void _navigateFromNotification(Map<String, dynamic> data) {
    // S'assurer que le widget est monté avant de naviguer
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      
      final String? type = data['type'];
      final String? auctionId = data['auctionId'];

      switch (type) {
        case 'auction_pending':
        case 'auction_approved':
        case 'auction_rejected':
        case 'auction_ended':
        case 'auction_won':
        case 'auction_ending_soon':
          if (auctionId != null) {
            _navigateToAuction(auctionId);
          }
          break;
        case 'new_message':
          // TODO: Navigate to messages page when available
          _navigateToHome();
          break;
        case 'payment_received':
        case 'withdrawal_processed':
          _navigateToWallet();
          break;
        default:
          // Navigation vers la page d'accueil par défaut
          _navigateToHome();
      }
    });
  }

  void _navigateToAuction(String auctionId) {
    try {
      final navigator = navigatorKey.currentState;
      if (navigator != null) {
        navigator.push(
          MaterialPageRoute(
            builder: (context) => AuctionDetailsPage(auctionId: auctionId),
          ),
        );
      }
    } catch (e) {
      developer.log('Navigation error: $e');
    }
  }

  void _navigateToWallet() {
    try {
      final navigator = navigatorKey.currentState;
      if (navigator != null) {
        navigator.push(
          MaterialPageRoute(
            builder: (context) => const DepositPage(),
          ),
        );
      }
    } catch (e) {
      developer.log('Navigation error: $e');
    }
  }

  void _navigateToHome() {
    try {
      final navigator = navigatorKey.currentState;
      if (navigator != null) {
        navigator.pushAndRemoveUntil(
          MaterialPageRoute(builder: (context) => const HomePage()),
          (route) => false,
        );
      }
    } catch (e) {
      developer.log('Navigation error: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    return widget.child;
  }
}

/// Provider pour suivre les notifications non lues
final unreadNotificationsProvider = StateNotifierProvider<UnreadNotificationsNotifier, int>((ref) {
  return UnreadNotificationsNotifier();
});

class UnreadNotificationsNotifier extends StateNotifier<int> {
  UnreadNotificationsNotifier() : super(0) {
    _init();
  }

  void _init() {
    // Écouter les nouvelles notifications
    FCMService().notificationStream.listen((data) {
      increment();
    });
  }

  void increment() => state++;
  void decrement() => state = state > 0 ? state - 1 : 0;
  void reset() => state = 0;
  void setCount(int count) => state = count;
}
