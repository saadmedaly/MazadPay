import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/notifications_api.dart';


class NotificationsPage extends ConsumerStatefulWidget {
  const NotificationsPage({super.key});

  @override
  ConsumerState<NotificationsPage> createState() => _NotificationsPageState();
}

class _NotificationsPageState extends ConsumerState<NotificationsPage> {
  final NotificationsApi _notificationsApi = NotificationsApi();
  List<Map<String, dynamic>> _notifications = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadNotifications();
  }

  Future<void> _loadNotifications() async {
    try {
      final response = await _notificationsApi.getNotifications();

      setState(() {
        _isLoading = false;
        if (response.success && response.data != null) {
          final dynamic responseData = response.data!;
          List<dynamic> notificationList = [];
          // La réponse peut être directement une liste ou un objet avec 'data' ou 'notifications'
          if (responseData is List) {
            notificationList = responseData;
          } else if (responseData is Map<String, dynamic>) {
            notificationList = (responseData['notifications'] ?? responseData['data'] ?? []) as List<dynamic>;
          }
          _notifications = notificationList.map((item) => item as Map<String, dynamic>).toList();
        }
      });
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.error_loading_notifications)),
        );
      }
    }
  }

  Future<void> _markAllAsRead() async {
    try {
      await _notificationsApi.markAllAsRead();
      setState(() {
        for (var notification in _notifications) {
          notification['read'] = true;
        }
      });
    } catch (e) {
      // Handle error silently with localized message
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.error_connection)),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          title: Text(
            l10n.text_48,
            style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: isDarkMode ? Colors.white : Colors.black,
            ),
          ),
          leading: IconButton(
            icon: Icon(Icons.arrow_back_ios, color: isDarkMode ? Colors.white : Colors.black, size: 20),
            onPressed: () => Navigator.of(context).pop(),
          ),
          actions: [
            TextButton(
              onPressed: _markAllAsRead,
              child: Text(
                l10n.text_245,
                style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: const Color(0xFF0081FF), fontSize: 12),
              ),
            ),
          ],
        ),
        body: _isLoading
            ? const Center(child: CircularProgressIndicator())
            : _notifications.isEmpty
                ? Center(
                    child: Text(
                      l10n.no_notifications,
                      style: TextStyle(color: Colors.grey, fontSize: 16),
                    ),
                  )
                : ListView.separated(
                    padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                    itemCount: _notifications.length,
                    separatorBuilder: (context, index) => const SizedBox(height: 12),
                    itemBuilder: (context, index) {
                      return _buildNotificationItem(context, _notifications[index], isDarkMode);
                    },
                  ),
      );
  }

  Widget _buildNotificationItem(BuildContext context, Map<String, dynamic> notification, bool isDarkMode) {
    final type = notification['type']?.toString() ?? 'system';
    final l10n = AppLocalizations.of(context)!;
    
    IconData icon;
    Color color;
    String title = notification['title']?.toString() ?? '';
    String description = notification['message']?.toString() ?? notification['description']?.toString() ?? '';
    String time = notification['created_at']?.toString() ?? l10n.text_246;

    switch (type) {
      case 'bid':
        icon = Icons.gavel_outlined;
        color = Colors.orange;
        break;
      case 'win':
        icon = Icons.emoji_events_outlined;
        color = const Color(0xFF00C58D);
        break;
      case 'payment':
        icon = Icons.payment_outlined;
        color = const Color(0xFF0081FF);
        break;
      case 'system':
        icon = Icons.notifications_active_outlined;
        color = Colors.blueGrey;
        break;
      default:
        icon = Icons.info_outline;
        color = Colors.grey;
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.grey.withOpacity(0.1)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(color: color.withOpacity(0.1), shape: BoxShape.circle),
            child: Icon(icon, color: color, size: 20),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(title, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontWeight: FontWeight.bold, fontSize: 14)),
                    Text(time, style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.grey, fontSize: 10)),
                  ],
                ),
                const SizedBox(height: 4),
                Text(
                  description,
                  style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: isDarkMode ? Colors.grey[400] : Colors.grey[600], fontSize: 12),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
