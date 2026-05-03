import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:mezadpay/services/notification_api.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Provider pour le FCM Service
final fcmServiceProvider = Provider<FCMService>((ref) => FCMService());

/// Service Firebase Cloud Messaging
/// Gère les notifications push, les tokens FCM et le deep linking
class FCMService {
  static final FCMService _instance = FCMService._internal();
  factory FCMService() => _instance;
  FCMService._internal();

  // Lazy initialization - ces variables seront initialisées dans initialize()
  FirebaseMessaging? _messaging;
  final NotificationApi _notificationApi = NotificationApi();
  final FlutterLocalNotificationsPlugin _localNotifications = FlutterLocalNotificationsPlugin();

  // Stream controller pour les notifications reçues
  final StreamController<Map<String, dynamic>> _notificationController = StreamController<Map<String, dynamic>>.broadcast();
  Stream<Map<String, dynamic>> get notificationStream => _notificationController.stream;

  // Callback pour la navigation (deep linking)
  Function(Map<String, dynamic>)? onNotificationTap;

  bool _initialized = false;
  String? _fcmToken;
  
  // Getter pour accéder à FirebaseMessaging de manière sécurisée
  FirebaseMessaging get _messagingInstance {
    if (_messaging == null) {
      try {
        _messaging = FirebaseMessaging.instance;
      } catch (e) {
        debugPrint('⚠️ FirebaseMessaging not initialized yet: $e');
        rethrow;
      }
    }
    return _messaging!;
  }

  /// Canal de notification pour les enchères (Android)
  static const String _auctionChannelId = 'auction_alerts';
  static const String _auctionChannelName = 'Alertes Enchères';
  static const String _auctionChannelDesc = 'Notifications pour les enchères en temps réel';

  /// Canal de notification pour la modération
  static const String _moderationChannelId = 'moderation';
  static const String _moderationChannelName = 'Modération';
  static const String _moderationChannelDesc = 'Alertes de modération et approbations';

  /// Canal pour les transactions
  static const String _transactionChannelId = 'transactions';
  static const String _transactionChannelName = 'Transactions';
  static const String _transactionChannelDesc = 'Notifications de paiements et retraits';

  /// Canal pour les messages
  static const String _messageChannelId = 'messages';
  static const String _messageChannelName = 'Messages';
  static const String _messageChannelDesc = 'Notifications de messages reçus';

  /// Canal pour les promotions
  static const String _promoChannelId = 'promotions';
  static const String _promoChannelName = 'Promotions';
  static const String _promoChannelDesc = 'Offres promotionnelles et tendances';

  /// Initialize the FCM Service
  Future<void> initialize() async {
    if (_initialized) {
      debugPrint('FCM Service already initialized');
      return;
    }
    
    // FCM n'est pas supporté sur web sans configuration Firebase spécifique
    if (kIsWeb) {
      debugPrint('ℹ️ FCM not available on web platform');
      return;
    }
    
    // Vérifier que Firebase est initialisé
    try {
      Firebase.app();
    } catch (e) {
      debugPrint('⚠️ Firebase not initialized yet, skipping FCM initialization');
      return;
    }
    
    try {
      // Demander les permissions
      await _requestPermissions();

      // Configurer les canaux de notification Android
      await _setupNotificationChannels();

      // Configurer le plugin de notifications locales
      await _setupLocalNotifications();

      // Configurer les handlers FCM
      await _setupFCMHandlers();

      // Récupérer et sauvegarder le token FCM
      await _saveFCMToken();

      // Écouter les changements de token
      _messagingInstance.onTokenRefresh.listen(_onTokenRefresh);

      _initialized = true;
      debugPrint('✅ FCM Service initialized successfully');
    } catch (e) {
      debugPrint('❌ FCM Service initialization error: $e');
    }
  }

  /// Demander les permissions de notification
  Future<void> _requestPermissions() async {
    if (kIsWeb) return;

    final settings = await _messagingInstance.requestPermission(
      alert: true,
      badge: true,
      sound: true,
      provisional: false,
      criticalAlert: false,
    );

    debugPrint('FCM Permission status: ${settings.authorizationStatus}');
  }

  /// Configurer les canaux de notification Android
  Future<void> _setupNotificationChannels() async {
    if (!Platform.isAndroid) return;

    final androidPlugin = _localNotifications.resolvePlatformSpecificImplementation<AndroidFlutterLocalNotificationsPlugin>();
    if (androidPlugin == null) return;

    // Canal des enchères (haute priorité)
    await androidPlugin.createNotificationChannel(
      AndroidNotificationChannel(
        _auctionChannelId,
        _auctionChannelName,
        description: _auctionChannelDesc,
        importance: Importance.high,
        playSound: true,
        sound: const RawResourceAndroidNotificationSound('notification'),
        enableVibration: true,
        showBadge: true,
      ),
    );

    // Canal de modération
    await androidPlugin.createNotificationChannel(
      AndroidNotificationChannel(
        _moderationChannelId,
        _moderationChannelName,
        description: _moderationChannelDesc,
        importance: Importance.high,
        playSound: true,
        enableVibration: true,
      ),
    );

    // Canal des transactions
    await androidPlugin.createNotificationChannel(
      AndroidNotificationChannel(
        _transactionChannelId,
        _transactionChannelName,
        description: _transactionChannelDesc,
        importance: Importance.defaultImportance,
        playSound: true,
      ),
    );

    // Canal des messages
    await androidPlugin.createNotificationChannel(
      AndroidNotificationChannel(
        _messageChannelId,
        _messageChannelName,
        description: _messageChannelDesc,
        importance: Importance.defaultImportance,
        playSound: true,
      ),
    );

    // Canal des promotions (basse priorité)
    await androidPlugin.createNotificationChannel(
      AndroidNotificationChannel(
        _promoChannelId,
        _promoChannelName,
        description: _promoChannelDesc,
        importance: Importance.low,
        playSound: false,
      ),
    );
  }

  /// Configurer le plugin de notifications locales
  Future<void> _setupLocalNotifications() async {
    const androidSettings = AndroidInitializationSettings('@mipmap/ic_launcher');
    const iosSettings = DarwinInitializationSettings(
      requestAlertPermission: false,
      requestBadgePermission: false,
      requestSoundPermission: false,
    );

    const initSettings = InitializationSettings(
      android: androidSettings,
      iOS: iosSettings,
    );

    await _localNotifications.initialize(
      initSettings,
      onDidReceiveNotificationResponse: _onNotificationTap,
    );
  }

  /// Handler pour le tap sur une notification locale
  void _onNotificationTap(NotificationResponse response) {
    if (response.payload == null) return;

    try {
      final data = jsonDecode(response.payload!) as Map<String, dynamic>;
      _notificationController.add(data);
      onNotificationTap?.call(data);
    } catch (e) {
      debugPrint('Error parsing notification payload: $e');
    }
  }

  /// Configurer les handlers FCM
  Future<void> _setupFCMHandlers() async {
    // Handler pour les messages en foreground
    FirebaseMessaging.onMessage.listen(_handleForegroundMessage);

    // Handler pour les messages en background (quand l'app est ouverte mais en arrière-plan)
    FirebaseMessaging.onMessageOpenedApp.listen(_handleBackgroundMessage);

    // Vérifier si l'app a été ouverte par une notification
    final initialMessage = await _messagingInstance.getInitialMessage();
    if (initialMessage != null) {
      _handleTerminatedMessage(initialMessage);
    }
  }

  /// Handler pour les messages en foreground (app visible)
  Future<void> _handleForegroundMessage(RemoteMessage message) async {
    debugPrint('📨 Foreground message received: ${message.messageId}');
    debugPrint('Title: ${message.notification?.title}');
    debugPrint('Body: ${message.notification?.body}');
    debugPrint('Data: ${message.data}');

    // Afficher une notification locale
    await _showLocalNotification(message);

    // Émettre l'événement
    _notificationController.add(message.data);
  }

  /// Handler pour les messages en background (app ouverte mais pas au premier plan)
  Future<void> _handleBackgroundMessage(RemoteMessage message) async {
    debugPrint('📨 Background message opened app: ${message.messageId}');
    _notificationController.add(message.data);
    onNotificationTap?.call(message.data);
  }

  /// Handler pour les messages quand l'app était terminée
  Future<void> _handleTerminatedMessage(RemoteMessage message) async {
    debugPrint('📨 Terminated message opened app: ${message.messageId}');
    // Delay pour laisser l'app s'initialiser
    await Future.delayed(const Duration(seconds: 2));
    _notificationController.add(message.data);
    onNotificationTap?.call(message.data);
  }

  /// Afficher une notification locale
  Future<void> _showLocalNotification(RemoteMessage message) async {
    final notification = message.notification;
    if (notification == null) return;

    final String channelId = _getChannelId(message.data['type']);
    final int notificationId = _generateNotificationId(message);

    final androidDetails = AndroidNotificationDetails(
      channelId,
      _getChannelName(channelId),
      channelDescription: _getChannelDescription(channelId),
      importance: _getImportance(channelId),
      priority: Priority.high,
      showWhen: true,
      enableVibration: true,
      playSound: channelId != _promoChannelId,
      icon: '@mipmap/ic_launcher',
      largeIcon: message.data['imageUrl'] != null
          ? FilePathAndroidBitmap(message.data['imageUrl'])
          : null,
      styleInformation: message.data['bigText'] == 'true'
          ? const BigTextStyleInformation('')
          : null,
    );

    final iosDetails = DarwinNotificationDetails(
      presentAlert: true,
      presentBadge: true,
      presentSound: channelId != _promoChannelId,
      interruptionLevel: channelId == _auctionChannelId
          ? InterruptionLevel.timeSensitive
          : InterruptionLevel.active,
    );

    final details = NotificationDetails(
      android: androidDetails,
      iOS: iosDetails,
    );

    await _localNotifications.show(
      notificationId,
      notification.title,
      notification.body,
      details,
      payload: jsonEncode(message.data),
    );
  }

  /// Récupérer le canal approprié selon le type
  String _getChannelId(String? type) {
    switch (type) {
      case 'auction_pending':
      case 'auction_approved':
      case 'auction_rejected':
      case 'auction_ended':
      case 'auction_won':
      case 'auction_ending_soon':
        return _auctionChannelId;
      case 'auction_reported':
      case 'auction_suspended':
        return _moderationChannelId;
      case 'payment_received':
      case 'withdrawal_processed':
        return _transactionChannelId;
      case 'new_message':
        return _messageChannelId;
      case 'promotion':
        return _promoChannelId;
      default:
        return _auctionChannelId;
    }
  }

  String _getChannelName(String channelId) {
    switch (channelId) {
      case _auctionChannelId:
        return _auctionChannelName;
      case _moderationChannelId:
        return _moderationChannelName;
      case _transactionChannelId:
        return _transactionChannelName;
      case _messageChannelId:
        return _messageChannelName;
      case _promoChannelId:
        return _promoChannelName;
      default:
        return _auctionChannelName;
    }
  }

  String _getChannelDescription(String channelId) {
    switch (channelId) {
      case _auctionChannelId:
        return _auctionChannelDesc;
      case _moderationChannelId:
        return _moderationChannelDesc;
      case _transactionChannelId:
        return _transactionChannelDesc;
      case _messageChannelId:
        return _messageChannelDesc;
      case _promoChannelId:
        return _promoChannelDesc;
      default:
        return _auctionChannelDesc;
    }
  }

  Importance _getImportance(String channelId) {
    switch (channelId) {
      case _auctionChannelId:
      case _moderationChannelId:
        return Importance.high;
      case _transactionChannelId:
      case _messageChannelId:
        return Importance.defaultImportance;
      case _promoChannelId:
        return Importance.low;
      default:
        return Importance.defaultImportance;
    }
  }

  /// Générer un ID unique pour la notification
  int _generateNotificationId(RemoteMessage message) {
    final data = message.data;
    final String idString = '${data['type']}_${data['auctionId'] ?? data['messageId'] ?? DateTime.now().millisecondsSinceEpoch}';
    return idString.hashCode.abs();
  }

  /// Sauvegarder le token FCM
  Future<void> _saveFCMToken() async {
    try {
      _fcmToken = await _messagingInstance.getToken();
      if (_fcmToken != null) {
        debugPrint('📱 FCM Token: $_fcmToken');
        await _sendTokenToServer(_fcmToken!);
      }
    } catch (e) {
      debugPrint('Error getting FCM token: $e');
    }
  }

  /// Envoyer le token au serveur
  Future<void> _sendTokenToServer(String token) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final deviceId = prefs.getString('device_id') ?? DateTime.now().millisecondsSinceEpoch.toString();
      
      await _notificationApi.saveToken(
        fcmToken: token,
        deviceId: deviceId,
        platform: Platform.isAndroid ? 'android' : (Platform.isIOS ? 'ios' : 'web'),
      );
      
      await prefs.setString('fcm_token', token);
      debugPrint('✅ FCM Token saved to server');
    } catch (e) {
      debugPrint('❌ Error saving FCM token to server: $e');
    }
  }

  /// Handler pour le refresh du token
  Future<void> _onTokenRefresh(String token) async {
    debugPrint('🔄 FCM Token refreshed: $token');
    _fcmToken = token;
    await _sendTokenToServer(token);
  }

  /// Récupérer le token FCM actuel
  Future<String?> getFCMToken() async {
    if (_fcmToken != null) return _fcmToken;
    _fcmToken = await _messagingInstance.getToken();
    return _fcmToken;
  }

  /// S'abonner à un topic
  Future<void> subscribeToTopic(String topic) async {
    try {
      await _messagingInstance.subscribeToTopic(topic);
      debugPrint('✅ Subscribed to topic: $topic');
    } catch (e) {
      debugPrint('❌ Error subscribing to topic $topic: $e');
    }
  }

  /// Se désabonner d'un topic
  Future<void> unsubscribeFromTopic(String topic) async {
    try {
      await _messagingInstance.unsubscribeFromTopic(topic);
      debugPrint('✅ Unsubscribed from topic: $topic');
    } catch (e) {
      debugPrint('❌ Error unsubscribing from topic $topic: $e');
    }
  }

  /// Supprimer le token FCM
  Future<void> deleteToken() async {
    try {
      await _messagingInstance.deleteToken();
      _fcmToken = null;
      debugPrint('✅ FCM Token deleted');
    } catch (e) {
      debugPrint('❌ Error deleting FCM token: $e');
    }
  }

  /// Gérer les notifications selon la langue
  static Map<String, String> getLocalizedNotification(
    String type,
    String languageCode, {
    Map<String, String>? params,
  }) {
    final localizations = _notificationLocalizations[type];
    if (localizations == null) {
      return {'title': '', 'body': ''};
    }

    final localeData = localizations[languageCode] ?? localizations['en']!;
    
    String title = localeData['title']!;
    String body = localeData['body']!;

    // Remplacer les paramètres
    if (params != null) {
      params.forEach((key, value) {
        title = title.replaceAll('{$key}', value);
        body = body.replaceAll('{$key}', value);
      });
    }

    return {'title': title, 'body': body};
  }

  /// Localisations des notifications
  static final Map<String, Map<String, Map<String, String>>> _notificationLocalizations = {
    'auction_pending': {
      'ar': {
        'title': 'مزاد جديد في الانتظار',
        'body': '{userName} أنشأ مزاد: {auctionTitle}',
      },
      'fr': {
        'title': 'Nouvelle enchère en attente',
        'body': '{userName} a créé une enchère: {auctionTitle}',
      },
      'en': {
        'title': 'New auction pending',
        'body': '{userName} created an auction: {auctionTitle}',
      },
    },
    'auction_approved': {
      'ar': {
        'title': 'تمت الموافقة على المزاد!',
        'body': 'مزادك "{auctionTitle}" أصبح متاحًا الآن',
      },
      'fr': {
        'title': 'Enchère approuvée !',
        'body': 'Votre enchère "{auctionTitle}" est maintenant en ligne',
      },
      'en': {
        'title': 'Auction approved!',
        'body': 'Your auction "{auctionTitle}" is now live',
      },
    },
    'auction_rejected': {
      'ar': {
        'title': 'تم رفض المزاد',
        'body': 'السبب: {reason}',
      },
      'fr': {
        'title': 'Enchère refusée',
        'body': 'Raison: {reason}',
      },
      'en': {
        'title': 'Auction rejected',
        'body': 'Reason: {reason}',
      },
    },
    'auction_ending_soon': {
      'ar': {
        'title': '⚡ الفرصة الأخيرة!',
        'body': '"{auctionTitle}" ينتهي في 5 دقائق',
      },
      'fr': {
        'title': '⚡ Dernière chance !',
        'body': '"{auctionTitle}" se termine dans 5 minutes',
      },
      'en': {
        'title': '⚡ Last chance!',
        'body': '"{auctionTitle}" ends in 5 minutes',
      },
    },
    'auction_won': {
      'ar': {
        'title': 'تهانينا! لقد فزت',
        'body': 'مزاد "{auctionTitle}" - {finalPrice} MRU',
      },
      'fr': {
        'title': 'Félicitations ! Vous avez gagné',
        'body': 'Enchère "{auctionTitle}" - {finalPrice} MRU',
      },
      'en': {
        'title': 'Congratulations! You won',
        'body': 'Auction "{auctionTitle}" - {finalPrice} MRU',
      },
    },
    'auction_ended': {
      'ar': {
        'title': 'انتهى المزاد',
        'body': 'تم بيع "{auctionTitle}" بـ {finalPrice} MRU',
      },
      'fr': {
        'title': 'Enchère terminée',
        'body': '"{auctionTitle}" vendu pour {finalPrice} MRU',
      },
      'en': {
        'title': 'Auction ended',
        'body': '"{auctionTitle}" sold for {finalPrice} MRU',
      },
    },
    'payment_received': {
      'ar': {
        'title': '💰 تم استلام الدفع',
        'body': '{amount} MRU لمزاد "{auctionTitle}"',
      },
      'fr': {
        'title': '💰 Paiement reçu',
        'body': '{amount} MRU pour "{auctionTitle}"',
      },
      'en': {
        'title': '💰 Payment received',
        'body': '{amount} MRU for "{auctionTitle}"',
      },
    },
    'new_message': {
      'ar': {
        'title': 'رسالة جديدة من {senderName}',
        'body': '{messagePreview}...',
      },
      'fr': {
        'title': 'Nouveau message de {senderName}',
        'body': '{messagePreview}...',
      },
      'en': {
        'title': 'New message from {senderName}',
        'body': '{messagePreview}...',
      },
    },
    'auction_reported': {
      'ar': {
        'title': '🚨 مزاد مُبلّغ عنه',
        'body': 'إبلاغ من {reporter} على "{auctionTitle}"',
      },
      'fr': {
        'title': '🚨 Enchère signalée',
        'body': 'Signalement de {reporter} sur "{auctionTitle}"',
      },
      'en': {
        'title': '🚨 Auction reported',
        'body': 'Report from {reporter} on "{auctionTitle}"',
      },
    },
    'promotion': {
      'ar': {
        'title': '🎉 عرض خاص',
        'body': '{promoText}',
      },
      'fr': {
        'title': '🎉 Promotion',
        'body': '{promoText}',
      },
      'en': {
        'title': '🎉 Special offer',
        'body': '{promoText}',
      },
    },
  };

  /// Dispose
  void dispose() {
    _notificationController.close();
  }
}

/// Extension pour créer une notification de deep linking
extension DeepLinkExtension on FCMService {
  /// Extraire la route de navigation depuis les données de notification
  String? extractRoute(Map<String, dynamic> data) {
    final String? type = data['type'];
    final String? auctionId = data['auctionId'];
    final String? conversationId = data['conversationId'];

    switch (type) {
      case 'auction_pending':
      case 'auction_approved':
      case 'auction_rejected':
      case 'auction_ended':
      case 'auction_won':
      case 'auction_ending_soon':
        if (auctionId != null) {
          return '/auction/$auctionId';
        }
        break;
      case 'new_message':
        if (conversationId != null) {
          return '/messages/$conversationId';
        }
        return '/messages';
      case 'payment_received':
      case 'withdrawal_processed':
        return '/wallet';
      case 'auction_reported':
      case 'auction_suspended':
        if (auctionId != null) {
          return '/admin/auction/$auctionId';
        }
        return '/admin';
      default:
        return null;
    }
    return null;
  }
}
