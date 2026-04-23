// lib/services/notification_service.dart
// Service pour Firebase Cloud Messaging et notifications locales

import 'dart:convert';

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter/material.dart';

import '../data/models/notification_model.dart';
import '../domain/repositories/auth_repository.dart';
import '../data/datasources/remote/api_client.dart';
import '../core/constants/api_constants.dart';

// Provider pour le service de notification
final notificationServiceProvider = Provider<NotificationService>((ref) {
  final dio = ref.watch(apiClientProvider);
  final authRepo = ref.watch(authRepositoryProvider);
  return NotificationService(dio: dio, authRepository: authRepo);
});

// Stream pour les notifications reçues en foreground
final foregroundNotificationProvider = StreamProvider<RemoteMessage>((ref) {
  return FirebaseMessaging.onMessage;
});

class NotificationService {
  final Dio dio;
  final AuthRepository authRepository;
  final FirebaseMessaging _messaging = FirebaseMessaging.instance;
  final FlutterLocalNotificationsPlugin _localNotifications = 
      FlutterLocalNotificationsPlugin();

  NotificationService({required this.dio, required this.authRepository});

  /// Initialise Firebase Messaging et les notifications locales
  Future<void> initialize() async {
    // Demande la permission
    final settings = await _messaging.requestPermission(
      alert: true,
      badge: true,
      sound: true,
      provisional: false,
    );

    print('Notification permission: ${settings.authorizationStatus}');

    // Configure les notifications locales
    const androidSettings = AndroidInitializationSettings('@mipmap/ic_launcher');
    const iosSettings = DarwinInitializationSettings(
      requestAlertPermission: true,
      requestBadgePermission: true,
      requestSoundPermission: true,
    );
    
    await _localNotifications.initialize(
      const InitializationSettings(
        android: androidSettings,
        iOS: iosSettings,
      ),
      onDidReceiveNotificationResponse: _onNotificationTap,
    );

    // Configure le canal Android pour les notifications
    const androidChannel = AndroidNotificationChannel(
      'high_importance_channel',
      'MazadPay Notifications',
      importance: Importance.high,
      playSound: true,
    );

    await _localNotifications
        .resolvePlatformSpecificImplementation<AndroidFlutterLocalNotificationsPlugin>()
        ?.createNotificationChannel(androidChannel);

    // Gère les messages en foreground
    FirebaseMessaging.onMessage.listen(_handleForegroundMessage);

    // Gère le token FCM
    _messaging.onTokenRefresh.listen(_onTokenRefresh);

    // Récupère et envoie le token initial
    await _sendFCMToken();
  }

  /// Envoie le token FCM au backend
  Future<void> _sendFCMToken() async {
    try {
      final token = await _messaging.getToken();
      if (token != null) {
        await _onTokenRefresh(token);
      }
    } catch (e) {
      print('Error getting FCM token: $e');
    }
  }

  /// Envoie le token au backend
  Future<void> _onTokenRefresh(String token) async {
    try {
      final request = PushTokenRequest(
        fcmToken: token,
        platform: _getPlatform(),
      );

      await dio.post(
        ApiConstants.pushTokens,
        data: request.toJson(),
      );
    } catch (e) {
      print('Error sending FCM token: $e');
    }
  }

  /// Gère les messages en foreground
  void _handleForegroundMessage(RemoteMessage message) {
    print('Foreground message: ${message.notification?.title}');
    
    // Affiche une notification locale
    _showLocalNotification(message);
    
    // TODO: Émettre un événement pour mettre à jour l'UI si nécessaire
  }

  /// Affiche une notification locale
  Future<void> _showLocalNotification(RemoteMessage message) async {
    final notification = message.notification;
    final android = message.notification?.android;

    if (notification == null) return;

    await _localNotifications.show(
      message.hashCode,
      notification.title,
      notification.body,
      NotificationDetails(
        android: AndroidNotificationDetails(
          'high_importance_channel',
          'MazadPay Notifications',
          channelDescription: 'Notifications for MazadPay app',
          importance: Importance.high,
          priority: Priority.high,
          icon: android?.smallIcon,
        ),
        iOS: const DarwinNotificationDetails(
          presentAlert: true,
          presentBadge: true,
          presentSound: true,
        ),
      ),
      payload: jsonEncode(message.data),
    );
  }

  /// Gère le tap sur une notification
  void _onNotificationTap(NotificationResponse response) {
    if (response.payload != null) {
      final data = jsonDecode(response.payload!);
      _handleNotificationTap(data);
    }
  }

  /// Navigation basée sur le type de notification
  void _handleNotificationTap(Map<String, dynamic> data) {
    final type = data['type'];
    final auctionId = data['auction_id'];

    switch (type) {
      case 'bid_placed':
      case 'auction_ended':
      case 'auction_won':
      case 'auction_ending':
        // TODO: Naviguer vers la page de détail de l'enchère
        // navigatorKey.currentState?.pushNamed('/auction/$auctionId');
        break;
      default:
        break;
    }
  }

  /// Récupère les notifications non lues
  Future<List<NotificationModel>> getUnreadNotifications() async {
    try {
      final response = await dio.get(
        ApiConstants.notifications,
        queryParameters: {'status': 'unread'},
      );

      if (response.statusCode == 200) {
        final List<dynamic> data = response.data['data'];
        return data.map((e) => NotificationModel.fromJson(e)).toList();
      }
      return [];
    } catch (e) {
      print('Error fetching notifications: $e');
      return [];
    }
  }

  /// Marque une notification comme lue
  Future<void> markAsRead(String notificationId) async {
    try {
      await dio.put('${ApiConstants.notifications}/$notificationId/read');
    } catch (e) {
      print('Error marking notification as read: $e');
    }
  }

  /// Marque toutes les notifications comme lues
  Future<void> markAllAsRead() async {
    try {
      await dio.put('${ApiConstants.notifications}/read-all');
    } catch (e) {
      print('Error marking all notifications as read: $e');
    }
  }

  /// Supprime le token FCM (logout)
  Future<void> deleteToken() async {
    try {
      await _messaging.deleteToken();
    } catch (e) {
      print('Error deleting FCM token: $e');
    }
  }

  /// Récupère le nombre de notifications non lues
  Future<int> getUnreadCount() async {
    try {
      final response = await dio.get('${ApiConstants.notifications}/unread-count');
      return response.data['count'] ?? 0;
    } catch (e) {
      return 0;
    }
  }

  String _getPlatform() {
    if (Platform.isIOS) return 'ios';
    if (Platform.isAndroid) return 'android';
    return 'unknown';
  }
}

/// Handler pour les messages en background
@pragma('vm:entry-point')
Future<void> firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp();
  print('Background message: ${message.notification?.title}');
  // TODO: Traiter le message en background si nécessaire
}
