import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:mezadpay/models/models.dart';
import 'package:mezadpay/services/notification_api.dart';

part 'notification_provider.g.dart';

/// État des notifications
class NotificationState {
  final List<Notification> notifications;
  final bool isLoading;
  final String? error;
  final int unreadCount;
  final int currentPage;
  final bool hasMore;
  final bool isMarkingAsRead;

  NotificationState({
    this.notifications = const [],
    this.isLoading = false,
    this.error,
    this.unreadCount = 0,
    this.currentPage = 1,
    this.hasMore = true,
    this.isMarkingAsRead = false,
  });

  NotificationState copyWith({
    List<Notification>? notifications,
    bool? isLoading,
    String? error,
    int? unreadCount,
    int? currentPage,
    bool? hasMore,
    bool? isMarkingAsRead,
  }) {
    return NotificationState(
      notifications: notifications ?? this.notifications,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      unreadCount: unreadCount ?? this.unreadCount,
      currentPage: currentPage ?? this.currentPage,
      hasMore: hasMore ?? this.hasMore,
      isMarkingAsRead: isMarkingAsRead ?? this.isMarkingAsRead,
    );
  }

  /// Notifications non lues
  List<Notification> get unreadNotifications =>
      notifications.where((n) => !n.isRead).toList();

  /// Notifications lues
  List<Notification> get readNotifications =>
      notifications.where((n) => n.isRead).toList();
}

/// Provider pour la gestion des notifications
@riverpod
class NotificationNotifier extends _$NotificationNotifier {
  final NotificationApi _notificationApi = NotificationApi();

  @override
  NotificationState build() {
    return NotificationState();
  }

  /// Charger les notifications
  Future<void> loadNotifications({
    int page = 1,
    int limit = 20,
    bool refresh = false,
  }) async {
    if (refresh) {
      state = state.copyWith(
        isLoading: true,
        error: null,
        currentPage: 1,
        notifications: [],
      );
    } else {
      state = state.copyWith(isLoading: true, error: null);
    }

    try {
      final response = await _notificationApi.getNotifications(
        page: refresh ? 1 : page,
        limit: limit,
      );

      if (response.success && response.data != null) {
        final data = response.data!;
        final notificationsList = data['notifications'] as List<dynamic>? ?? [];
        final notifications = notificationsList
            .map((e) => Notification.fromJson(e as Map<String, dynamic>))
            .toList();

        final unreadCount = data['unread_count'] ?? 0;
        final currentPage = data['page'] ?? page;

        state = state.copyWith(
          isLoading: false,
          notifications: refresh
              ? notifications
              : [...state.notifications, ...notifications],
          unreadCount: unreadCount,
          currentPage: currentPage,
          hasMore: notifications.length >= limit,
          error: null,
        );
      } else {
        state = state.copyWith(
          isLoading: false,
          error: response.error?.message ?? 'Erreur de chargement des notifications',
        );
      }
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  /// Charger plus de notifications (pagination)
  Future<void> loadMoreNotifications() async {
    if (state.isLoading || !state.hasMore) return;

    await loadNotifications(page: state.currentPage + 1);
  }

  /// Marquer une notification comme lue
  Future<bool> markAsRead(String notificationId) async {
    state = state.copyWith(isMarkingAsRead: true);

    try {
      final response = await _notificationApi.markNotificationAsRead(notificationId);

      if (response.success) {
        // Mettre à jour localement
        final updatedNotifications = state.notifications.map((n) {
          if (n.id == notificationId) {
            return n.markAsRead();
          }
          return n;
        }).toList();

        state = state.copyWith(
          notifications: updatedNotifications,
          unreadCount: state.unreadCount > 0 ? state.unreadCount - 1 : 0,
          isMarkingAsRead: false,
        );
        return true;
      } else {
        state = state.copyWith(
          isMarkingAsRead: false,
          error: response.error?.message,
        );
        return false;
      }
    } catch (e) {
      state = state.copyWith(
        isMarkingAsRead: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Marquer toutes les notifications comme lues
  Future<bool> markAllAsRead() async {
    state = state.copyWith(isMarkingAsRead: true);

    try {
      final response = await _notificationApi.markAllNotificationsAsRead();

      if (response.success) {
        // Mettre à jour localement
        final updatedNotifications = state.notifications.map((n) {
          return n.markAsRead();
        }).toList();

        state = state.copyWith(
          notifications: updatedNotifications,
          unreadCount: 0,
          isMarkingAsRead: false,
        );
        return true;
      } else {
        state = state.copyWith(
          isMarkingAsRead: false,
          error: response.error?.message,
        );
        return false;
      }
    } catch (e) {
      state = state.copyWith(
        isMarkingAsRead: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Sauvegarder le token FCM pour les push notifications
  Future<bool> saveFCMToken({
    required String fcmToken,
    required String deviceId,
    required String platform,
  }) async {
    try {
      final response = await _notificationApi.saveToken(
        fcmToken: fcmToken,
        deviceId: deviceId,
        platform: platform,
      );

      return response.success;
    } catch (e) {
      return false;
    }
  }

  /// Rafraîchir les notifications
  Future<void> refresh() async {
    await loadNotifications(refresh: true);
  }

  /// Effacer les erreurs
  void clearError() {
    state = state.copyWith(error: null);
  }

  /// Supprimer une notification localement
  void removeNotification(String notificationId) {
    final wasUnread = state.notifications
        .firstWhere((n) => n.id == notificationId, orElse: () => null as Notification)
        .isRead == false;

    final updatedNotifications = state.notifications
        .where((n) => n.id != notificationId)
        .toList();

    state = state.copyWith(
      notifications: updatedNotifications,
      unreadCount: wasUnread && state.unreadCount > 0
          ? state.unreadCount - 1
          : state.unreadCount,
    );
  }
}
