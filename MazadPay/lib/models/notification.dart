/// Modèle Notification basé sur le backend Go
/// Correspond à `backend/internal/models/notification.go`
class Notification {
  final String id;
  final String userId;
  final String type;
  final String title;
  final String? body;
  final bool isRead;
  final String? referenceId;
  final String? referenceType;
  final Map<String, dynamic>? data;
  final DateTime createdAt;

  Notification({
    required this.id,
    required this.userId,
    required this.type,
    required this.title,
    this.body,
    this.isRead = false,
    this.referenceId,
    this.referenceType,
    this.data,
    required this.createdAt,
  });

  factory Notification.fromJson(Map<String, dynamic> json) {
    return Notification(
      id: json['id'] ?? '',
      userId: json['user_id'] ?? '',
      type: json['type'] ?? '',
      title: json['title'] ?? '',
      body: json['body'],
      isRead: json['is_read'] ?? false,
      referenceId: json['reference_id']?.toString(),
      referenceType: json['reference_type'],
      data: json['data'] != null
          ? Map<String, dynamic>.from(json['data'])
          : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'type': type,
      'title': title,
      'body': body,
      'is_read': isRead,
      'reference_id': referenceId,
      'reference_type': referenceType,
      'data': data,
      'created_at': createdAt.toIso8601String(),
    };
  }

  /// Marquer comme lue
  Notification markAsRead() {
    return Notification(
      id: id,
      userId: userId,
      type: type,
      title: title,
      body: body,
      isRead: true,
      referenceId: referenceId,
      referenceType: referenceType,
      data: data,
      createdAt: createdAt,
    );
  }

  /// Formate la date relative (il y a X minutes/heures/jours)
  String getTimeAgo() {
    final now = DateTime.now();
    final difference = now.difference(createdAt);

    if (difference.inDays > 365) {
      return '${difference.inDays ~/ 365}y';
    } else if (difference.inDays > 30) {
      return '${difference.inDays ~/ 30}mo';
    } else if (difference.inDays > 0) {
      return '${difference.inDays}d';
    } else if (difference.inHours > 0) {
      return '${difference.inHours}h';
    } else if (difference.inMinutes > 0) {
      return '${difference.inMinutes}m';
    } else {
      return 'now';
    }
  }
}

/// Types de notifications
class NotificationType {
  static const String bidPlaced = 'bid_placed';
  static const String bidOutbid = 'bid_outbid';
  static const String auctionWon = 'auction_won';
  static const String auctionLost = 'auction_lost';
  static const String auctionEnding = 'auction_ending';
  static const String auctionEnded = 'auction_ended';
  static const String paymentReceived = 'payment_received';
  static const String paymentConfirmed = 'payment_confirmed';
  static const String kycApproved = 'kyc_approved';
  static const String kycRejected = 'kyc_rejected';
  static const String depositConfirmed = 'deposit_confirmed';
  static const String withdrawalProcessed = 'withdrawal_processed';
  static const String newMessage = 'new_message';
  static const String system = 'system';
  static const String promotion = 'promotion';
}

/// Modèle PushToken pour FCM
class PushToken {
  final String id;
  final String userId;
  final String fcmToken;
  final String deviceId;
  final String platform; // android, ios, web
  final bool isActive;
  final DateTime createdAt;
  final DateTime updatedAt;

  PushToken({
    required this.id,
    required this.userId,
    required this.fcmToken,
    required this.deviceId,
    required this.platform,
    this.isActive = true,
    required this.createdAt,
    required this.updatedAt,
  });

  factory PushToken.fromJson(Map<String, dynamic> json) {
    return PushToken(
      id: json['id'] ?? '',
      userId: json['user_id'] ?? '',
      fcmToken: json['fcm_token'] ?? '',
      deviceId: json['device_id'] ?? '',
      platform: json['platform'] ?? 'android',
      isActive: json['is_active'] ?? true,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'])
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'fcm_token': fcmToken,
      'device_id': deviceId,
      'platform': platform,
      'is_active': isActive,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }
}

/// Préférences de notifications utilisateur
class NotificationPreferences {
  final bool emailEnabled;
  final bool pushEnabled;
  final bool smsEnabled;
  final Map<String, bool> typePreferences;

  NotificationPreferences({
    this.emailEnabled = true,
    this.pushEnabled = true,
    this.smsEnabled = false,
    Map<String, bool>? typePreferences,
  }) : typePreferences = typePreferences ?? {};

  factory NotificationPreferences.fromJson(Map<String, dynamic> json) {
    return NotificationPreferences(
      emailEnabled: json['email_enabled'] ?? true,
      pushEnabled: json['push_enabled'] ?? true,
      smsEnabled: json['sms_enabled'] ?? false,
      typePreferences: json['type_preferences'] != null
          ? Map<String, bool>.from(json['type_preferences'])
          : {},
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'email_enabled': emailEnabled,
      'push_enabled': pushEnabled,
      'sms_enabled': smsEnabled,
      'type_preferences': typePreferences,
    };
  }
}
