import 'dart:async';
import 'dart:developer' as developer;

import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/pages/auction_details_page.dart';
import 'package:mezadpay/pages/auction_winner_page.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/pages/deposit_page.dart';
import 'package:mezadpay/pages/notification_detail_page.dart';
import 'package:mezadpay/services/auction_api.dart';
import 'package:mezadpay/services/fcm_service.dart';
import 'package:mezadpay/services/notifications_api.dart';
import 'package:mezadpay/services/realtime_sync_service.dart';

/// Global key pour accéder au Navigator depuis n'importe où
final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();

/// Widget qui gère la navigation depuis les notifications FCM
/// Placez ce widget au sommet de l'arborescence de navigation
class NotificationHandler extends ConsumerStatefulWidget {
  final Widget child;

  const NotificationHandler({
    super.key,
    required this.child,
  });

  @override
  ConsumerState<NotificationHandler> createState() => _NotificationHandlerState();
}

class _NotificationHandlerState extends ConsumerState<NotificationHandler> {
  FCMService? _fcmService;
  StreamSubscription? _notificationSubscription;
  StreamSubscription<RealtimeEvent>? _winnerRealtimeSubscription;

  // Customer #23: session-level "already surfaced" guard -- the smallest
  // safe mechanism to satisfy "foreground realtime event may trigger
  // congratulations once" without any new persistence. Cleared on app
  // restart (in-memory only) by design: a genuinely new session re-showing
  // a still-recent win is an acceptable, conservative trade-off versus the
  // alternative of needing new persisted state; the explicit anti-goal is
  // "do NOT automatically replay every old historical win as a new popup"
  // on a normal resume/reconnect within the SAME session, which this set
  // fully prevents since it is never cleared by resume/reconnect.
  final Set<String> _surfacedWinnerAuctionIds = {};
  bool _winnerCheckInFlight = false;

  @override
  void initState() {
    super.initState();
    _initializeFCM();
    _winnerRealtimeSubscription = RealtimeSyncService().events.listen((event) {
      if (event.type == RealtimeEventType.auctionStatusChanged) {
        _maybeShowForegroundWinner(event.entityId);
      }
    });
  }

  /// Customer #23 foreground winner coordinator: on ANY auction.status_changed
  /// event (an auction closing is the only thing that ever fires it -- see
  /// AuctionService.FinalizeExpiredAuction/emitAuctionEvent), checks whether
  /// the CURRENT signed-in user actually won THAT auction, using the real
  /// authoritative GET /users/me/winnings endpoint -- never guessing winner
  /// identity client-side from the event itself, which carries no winner
  /// information. If the changed auction is present in the response and
  /// hasn't already been surfaced this session, opens AuctionWinnerPage once.
  /// This runs regardless of whether MyWinningsPage is currently open --
  /// exactly the "global foreground experience" requirement -- since
  /// NotificationHandler lives at the app root, not inside any one page.
  Future<void> _maybeShowForegroundWinner(String auctionId) async {
    if (auctionId.isEmpty) return;
    if (_surfacedWinnerAuctionIds.contains(auctionId)) return;
    // Coalesce overlapping calls (e.g. several status_changed events in a
    // short burst) into a single in-flight winnings check at a time, rather
    // than firing one request per event.
    if (_winnerCheckInFlight) return;
    _winnerCheckInFlight = true;
    try {
      final response = await AuctionApi().getMyWinnings();
      if (!mounted) return;
      if (!response.success || response.data == null) return;

      final dynamic responseData = response.data!;
      List<dynamic> winnings = [];
      if (responseData is List) {
        winnings = responseData;
      } else if (responseData is Map<String, dynamic>) {
        winnings = (responseData['auctions'] ?? responseData['data'] ?? []) as List<dynamic>;
      }

      final match = winnings.whereType<Map<String, dynamic>>().where(
            (w) => w['id']?.toString() == auctionId,
          );
      if (match.isEmpty) {
        // Either this user isn't the winner of the changed auction, or the
        // change wasn't a win at all (e.g. a no-bid/cancelled close) --
        // both are simply "nothing to show", not an error.
        return;
      }

      if (_surfacedWinnerAuctionIds.contains(auctionId)) return; // re-check after await
      _surfacedWinnerAuctionIds.add(auctionId);

      final navigator = navigatorKey.currentState;
      if (navigator == null) return;
      navigator.push(
        MaterialPageRoute(builder: (context) => AuctionWinnerPage(auctionId: auctionId)),
      );
    } catch (e) {
      developer.log('Foreground winner check error: $e');
    } finally {
      _winnerCheckInFlight = false;
    }
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
    _winnerRealtimeSubscription?.cancel();
    _fcmService?.onNotificationTap = null;
    super.dispose();
  }

  /// Handler pour le tap sur une notification
  void _handleNotificationTap(Map<String, dynamic> data) {
    _navigateFromNotification(data);
  }

  /// Handler pour les notifications reçues
  void _handleNotification(Map<String, dynamic> data) {
    final String? type = data['type'];
    if (type == 'auction_reported') {
      _navigateFromNotification(data);
    }
  }

  /// Navigation vers la bonne page selon la notification
  void _navigateFromNotification(Map<String, dynamic> data) {
    // S'assurer que le widget est monté avant de naviguer
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;

      final String? type = data['type'];
      final String? auctionId = data['auctionId'] ?? data['auction_id'];

      switch (type) {
        // Customer feedback #11: auction_won is only ever sent by the backend
        // scheduler to the confirmed winner (topBid.UserID, persisted via
        // SetWinner before the push fires -- see auction_scheduler.go), so a
        // tap on it is an authoritative win signal, not a client-side guess.
        // Routes to the congratulations page (AuctionWinnerPage) instead of
        // plain auction details; every other auction event keeps going there.
        case 'auction_won':
          if (auctionId != null) {
            _navigateToWinner(auctionId);
          } else {
            _openGenericDetailById(data);
          }
          break;
        case 'auction_pending':
        case 'auction_approved':
        case 'auction_rejected':
        case 'auction_ended':
        case 'bid_outbid':
          if (auctionId != null) {
            _navigateToAuction(auctionId);
          } else {
            _openGenericDetailById(data);
          }
          break;
        case 'new_message':
          _navigateToHome();
          break;
        case 'payment_received':
        case 'deposit_confirmed':
        case 'deposit_rejected':
        case 'withdrawal_processed':
          _navigateToWallet();
          break;
        // Customer #22: admin broadcast notifications (general/new_auction/
        // transaction) previously fell into `default: _navigateToHome()` --
        // a misleading redirect that wiped the nav stack and gave no
        // indication the tap was notification-related. Preferred
        // specialized navigation first (if this specific push DOES carry a
        // real auctionId), then the generic detail screen looked up by
        // notification_id, matching the exact routing precedence used by
        // the in-app list (notifications_page.dart's _onNotificationTap).
        case 'general':
        case 'transaction':
        case 'new_auction':
          if (auctionId != null) {
            _navigateToAuction(auctionId);
          } else {
            _openGenericDetailById(data);
          }
          break;
        default:
          // Customer #22: an unrecognized type still deserves an attempt at
          // opening its real content (via notification_id) rather than an
          // unconditional silent redirect to Home -- _openGenericDetailById
          // itself falls back to Home only if no id is present or the
          // lookup fails, matching the "fail safely" requirement.
          _openGenericDetailById(data);
      }
    });
  }

  /// Customer #22: looks up the tapped push's own DB row via the
  /// already-user-scoped GET /notifications (never a new GET
  /// /notifications/:id endpoint), then opens NotificationDetailPage for it.
  /// Used both for a genuinely generic/broadcast type and as the "no real
  /// target present" fallback for types that normally have a specialized
  /// destination. Fails safely to Home if notification_id is absent, the
  /// fetch fails (e.g. auth not ready yet), or no matching row is found --
  /// never throws, never crashes.
  Future<void> _openGenericDetailById(Map<String, dynamic> data) async {
    final notificationId = data['notification_id']?.toString();
    if (notificationId == null || notificationId.isEmpty) {
      _navigateToHome();
      return;
    }

    try {
      final response = await NotificationsApi().getNotifications();
      if (!mounted) return;
      if (!response.success || response.data == null) {
        _navigateToHome();
        return;
      }
      final match = response.data!.whereType<Map<String, dynamic>>().where(
            (n) => n['id']?.toString() == notificationId,
          );
      if (match.isEmpty) {
        _navigateToHome();
        return;
      }
      final notification = match.first;
      final createdAtRaw = notification['created_at']?.toString();
      final createdAt = createdAtRaw != null
          ? (DateTime.tryParse(createdAtRaw) ?? DateTime.now())
          : DateTime.now();

      final navigator = navigatorKey.currentState;
      if (navigator == null) return;
      navigator.push(
        MaterialPageRoute(
          builder: (context) => NotificationDetailPage(
            title: notification['title']?.toString() ?? '',
            body: notification['body']?.toString(),
            imageUrl: notification['image_url']?.toString(),
            createdAt: createdAt,
          ),
        ),
      );
    } catch (e) {
      developer.log('Notification detail lookup error: $e');
      if (mounted) _navigateToHome();
    }
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

  void _navigateToWinner(String auctionId) {
    try {
      final navigator = navigatorKey.currentState;
      if (navigator != null) {
        navigator.push(
          MaterialPageRoute(
            builder: (context) => AuctionWinnerPage(auctionId: auctionId),
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
