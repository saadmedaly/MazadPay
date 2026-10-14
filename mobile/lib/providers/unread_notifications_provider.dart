import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../services/notifications_api.dart';

/// Customer Request #30: Home's notification bell badge.
///
/// Deliberately reuses NotificationsApi (the service notifications_page.dart
/// already uses -- GET /notifications returning a correctly-typed bare
/// array, see notifications_api.dart), not the separate/unused
/// NotificationNotifier in notification_provider.dart (that class already
/// expects a backend `unread_count` field the API has never actually sent,
/// so its own unreadCount has always silently defaulted to 0 -- introducing
/// a second, real notification-loading path there would be exactly the
/// "second notification system" this ticket says not to build). The
/// backend has no unread-count endpoint/field at all, so this counts
/// `is_read == false` client-side across the same notification list the
/// Notifications page already fetches and displays -- no backend change.
class UnreadNotificationsCount extends StateNotifier<int> {
  UnreadNotificationsCount() : super(0);

  final NotificationsApi _api = NotificationsApi();

  Future<void> refresh() async {
    try {
      final response = await _api.getNotifications();
      if (response.success && response.data != null) {
        state = countUnread(response.data!);
      }
    } catch (_) {
      // Never let a failed refresh crash or block the bell icon -- the
      // badge simply keeps its last known value, mirroring how the rest of
      // this app's notification fetches fail safely (snackbar/no-op, never
      // a thrown exception reaching the UI).
    }
  }

  /// Pure counting logic, extracted for direct unit testing without a live
  /// API call. Mirrors the exact `is_read` field/default already used by
  /// Notification.fromJson (models/notification.dart) and
  /// notifications_page.dart's own local read-state handling.
  static int countUnread(List<dynamic> notifications) {
    return notifications.where((item) {
      if (item is Map) {
        return item['is_read'] != true;
      }
      return false;
    }).length;
  }

}

final unreadNotificationsCountProvider = StateNotifierProvider<UnreadNotificationsCount, int>(
  (ref) => UnreadNotificationsCount(),
);
