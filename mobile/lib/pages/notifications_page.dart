import 'dart:convert';

import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import '../services/fcm_service.dart';
import '../services/notifications_api.dart';
import 'auction_details_page.dart';


class NotificationsPage extends ConsumerStatefulWidget {
  const NotificationsPage({super.key});

  @override
  ConsumerState<NotificationsPage> createState() => _NotificationsPageState();
}

class _NotificationsPageState extends ConsumerState<NotificationsPage> {
  final NotificationsApi _notificationsApi = NotificationsApi();
  final FCMService _fcmService = FCMService();
  List<Map<String, dynamic>> _notifications = [];
  bool _isLoading = true;
  String _selectedFilter = 'all';

  @override
  void initState() {
    super.initState();
    _loadNotifications();
    
    // Écouter les nouvelles notifications FCM
    _fcmService.notificationStream.listen((data) {
      _loadNotifications(); // Recharger quand une nouvelle notification arrive
    });
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

  List<Map<String, dynamic>> get _filteredNotifications {
    if (_selectedFilter == 'all') return _notifications;
    return _notifications.where((n) {
      final type = n['type']?.toString() ?? 'system';
      switch (_selectedFilter) {
        case 'auctions':
          return ['auction_pending', 'auction_approved', 'auction_rejected', 'auction_ended', 'auction_won', 'auction_ending_soon'].contains(type);
        case 'messages':
          return type == 'new_message';
        case 'payments':
          return ['payment_received', 'withdrawal_processed'].contains(type);
        default:
          return true;
      }
    }).toList();
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).languageCode;

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: true,
        title: Text(
          l10n.text_48,
          style: GoogleFonts.plusJakartaSans(
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
              style: GoogleFonts.plusJakartaSans(color: const Color(0xFF0081FF), fontSize: 12),
            ),
          ),
        ],
      ),
      body: Column(
        children: [
          // Filtres
          _buildFilterChips(isDarkMode, locale),
          
          // Liste des notifications
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : _filteredNotifications.isEmpty
                    ? _buildEmptyState(isDarkMode, locale)
                    : ListView.separated(
                        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                        itemCount: _filteredNotifications.length,
                        separatorBuilder: (context, index) => const SizedBox(height: 12),
                        itemBuilder: (context, index) {
                          return _buildNotificationItem(context, _filteredNotifications[index], isDarkMode, locale);
                        },
                      ),
          ),
        ],
      ),
    );
  }

  Widget _buildFilterChips(bool isDarkMode, String locale) {
    final filters = [
      {'key': 'all', 'label': locale == 'ar' ? 'الكل' : (locale == 'fr' ? 'Tous' : 'All')},
      {'key': 'auctions', 'label': locale == 'ar' ? 'المزادات' : (locale == 'fr' ? 'Enchères' : 'Auctions')},
      {'key': 'messages', 'label': locale == 'ar' ? 'الرسائل' : (locale == 'fr' ? 'Messages' : 'Messages')},
      {'key': 'payments', 'label': locale == 'ar' ? 'المدفوعات' : (locale == 'fr' ? 'Paiements' : 'Payments')},
    ];

    return Container(
      height: 50,
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        itemCount: filters.length,
        separatorBuilder: (_, __) => const SizedBox(width: 8),
        itemBuilder: (context, index) {
          final filter = filters[index];
          final isSelected = _selectedFilter == filter['key'];
          
          return ChoiceChip(
            label: Text(filter['label']!),
            selected: isSelected,
            onSelected: (_) => setState(() => _selectedFilter = filter['key']!),
            backgroundColor: isDarkMode ? const Color(0xFF1D1D1D) : Colors.grey.shade100,
            selectedColor: const Color(0xFF0081FF).withOpacity(0.1),
            labelStyle: GoogleFonts.plusJakartaSans(
              color: isSelected ? const Color(0xFF0081FF) : (isDarkMode ? Colors.white70 : Colors.black54),
              fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
            ),
            side: BorderSide(
              color: isSelected ? const Color(0xFF0081FF) : Colors.transparent,
            ),
          );
        },
      ),
    );
  }

  Widget _buildEmptyState(bool isDarkMode, String locale) {
    final message = locale == 'ar' 
        ? 'لا توجد إشعارات'
        : (locale == 'fr' ? 'Aucune notification' : 'No notifications');
    
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.notifications_off_outlined,
            size: 64,
            color: isDarkMode ? Colors.grey.shade600 : Colors.grey.shade300,
          ),
          const SizedBox(height: 16),
          Text(
            message,
            style: GoogleFonts.plusJakartaSans(
              color: isDarkMode ? Colors.grey.shade400 : Colors.grey,
              fontSize: 16,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildNotificationItem(BuildContext context, Map<String, dynamic> notification, bool isDarkMode, String locale) {
    final type = notification['type']?.toString() ?? 'system';
    final isRead = notification['read'] == true;
    
    IconData icon;
    Color color;
    String title = notification['title']?.toString() ?? '';
    String description = notification['message']?.toString() ?? notification['description']?.toString() ?? '';
    String time = _formatTime(notification['created_at']?.toString() ?? '');

    // Icônes et couleurs selon le type FCM
    switch (type) {
      case 'auction_pending':
      case 'auction_approved':
      case 'auction_rejected':
      case 'auction_ended':
      case 'auction_won':
      case 'auction_ending_soon':
        icon = Icons.gavel_outlined;
        color = const Color(0xFF0081FF);
        break;
      case 'payment_received':
      case 'withdrawal_processed':
        icon = Icons.payment_outlined;
        color = const Color(0xFF00C58D);
        break;
      case 'new_message':
        icon = Icons.message_outlined;
        color = Colors.orange;
        break;
      case 'auction_reported':
      case 'auction_suspended':
        icon = Icons.report_problem_outlined;
        color = const Color(0xFFFF3B30);
        break;
      case 'promotion':
        icon = Icons.local_offer_outlined;
        color = Colors.purple;
        break;
      default:
        icon = Icons.notifications_outlined;
        color = Colors.blueGrey;
    }

    return GestureDetector(
      onTap: () => _onNotificationTap(notification),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: isRead 
              ? (isDarkMode ? const Color(0xFF1D1D1D) : Colors.white)
              : (isDarkMode ? const Color(0xFF252542) : const Color(0xFFF0F7FF)),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isRead 
                ? (isDarkMode ? Colors.grey.shade800 : Colors.grey.shade200)
                : const Color(0xFF0081FF).withOpacity(0.3),
            width: isRead ? 1 : 1.5,
          ),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Indicateur non lu
            if (!isRead)
              Container(
                width: 8,
                height: 8,
                margin: const EdgeInsets.only(right: 8, top: 4),
                decoration: const BoxDecoration(
                  color: Color(0xFF0081FF),
                  shape: BoxShape.circle,
                ),
              ),
            
            // Icône
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: color.withOpacity(0.1), 
                shape: BoxShape.circle
              ),
              child: Icon(icon, color: color, size: 20),
            ),
            const SizedBox(width: 12),
            
            // Contenu
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          title,
                          style: GoogleFonts.plusJakartaSans(
                            fontWeight: isRead ? FontWeight.w500 : FontWeight.bold,
                            fontSize: 14,
                            color: isDarkMode ? Colors.white : Colors.black87,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        time,
                        style: GoogleFonts.plusJakartaSans(
                          color: Colors.grey,
                          fontSize: 11,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    description,
                    style: GoogleFonts.plusJakartaSans(
                      color: isDarkMode ? Colors.grey.shade400 : Colors.grey.shade600,
                      fontSize: 12,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _formatTime(String createdAt) {
    if (createdAt.isEmpty) return '';
    try {
      final date = DateTime.parse(createdAt);
      final now = DateTime.now();
      final diff = now.difference(date);
      
      if (diff.inMinutes < 1) return 'now';
      if (diff.inHours < 1) return '${diff.inMinutes}m';
      if (diff.inDays < 1) return '${diff.inHours}h';
      if (diff.inDays < 7) return '${diff.inDays}d';
      return DateFormat('dd/MM').format(date);
    } catch (e) {
      return createdAt;
    }
  }

  void _onNotificationTap(Map<String, dynamic> notification) {
    final String? type = notification['type'];
    final String? auctionId = notification['auction_id']?.toString() ?? notification['auctionId']?.toString();
    final String? conversationId = notification['conversation_id']?.toString() ?? notification['conversationId']?.toString();
    
    // Marquer comme lu
    _markAsRead(notification['id']?.toString());

    switch (type) {
      case 'auction_pending':
      case 'auction_approved':
      case 'auction_rejected':
      case 'auction_ended':
      case 'auction_won':
      case 'auction_ending_soon':
        if (auctionId != null) {
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => AuctionDetailsPage(auctionId: auctionId),
            ),
          );
        }
        break;
      // TODO: Add other navigation cases
      default:
        break;
    }
  }

  Future<void> _markAsRead(String? notificationId) async {
    if (notificationId == null) return;
    try {
      // TODO: Call API to mark as read
      setState(() {
        final index = _notifications.indexWhere((n) => n['id']?.toString() == notificationId);
        if (index != -1) {
          _notifications[index]['read'] = true;
        }
      });
    } catch (e) {
      debugPrint('Error marking notification as read: $e');
    }
  }
}
