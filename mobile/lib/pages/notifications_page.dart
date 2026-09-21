import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import '../services/fcm_service.dart';
import '../services/notification_api.dart';
import '../services/notifications_api.dart';
import 'auction_details_page.dart';
import 'auction_winner_page.dart';
import 'deposit_page.dart';
import 'notification_detail_page.dart';
import 'withdrawal_detail_page.dart';


/// Customer #21 hardening round: pure tab-classification decision, extracted
/// from _filteredNotifications for direct unit testing. The 'auctions' and
/// 'payments' tabs are closed whitelists (an unrecognized type matches
/// neither -- it still renders correctly under 'all', just isn't picked up
/// by a specific tab); 'all' (or any other filter key) matches everything.
///
/// 'payments' additionally covers deposit_submitted/withdrawal_submitted
/// (Customer #21's new submission-acknowledgment types) alongside the
/// existing deposit_confirmed/deposit_rejected/withdrawal_processed outcome
/// types -- both are wallet/payment notifications and belong on the same
/// tab, even though only the outcome types were originally listed here.
bool matchesNotificationTabFilter(String type, String filterKey) {
  switch (filterKey) {
    case 'auctions':
      return const ['auction_pending', 'auction_approved', 'auction_rejected', 'auction_ended', 'auction_won'].contains(type);
    case 'payments':
      return const [
        'payment_received',
        'deposit_submitted',
        'deposit_confirmed',
        'deposit_rejected',
        'withdrawal_submitted',
        'withdrawal_processed',
      ].contains(type);
    default:
      return true;
  }
}

class NotificationsPage extends ConsumerStatefulWidget {
  const NotificationsPage({super.key});

  @override
  ConsumerState<NotificationsPage> createState() => _NotificationsPageState();
}

class _NotificationsPageState extends ConsumerState<NotificationsPage> {
  final NotificationApi _notificationApi = NotificationApi();
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
        // NotificationsApi.getNotifications() now returns ApiResponse<List<dynamic>>
        // directly (see notifications_api.dart) -- the backend's "data" field is
        // always a bare array, never a nested Map, so no runtime type branching
        // is needed here anymore.
        if (response.success && response.data != null) {
          _notifications = response.data!.map((item) => item as Map<String, dynamic>).toList();
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
          notification['is_read'] = true;
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
      return matchesNotificationTabFilter(type, _selectedFilter);
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
    final isRead = notification['is_read'] == true;
    
    IconData icon;
    Color color;
    String title = notification['title']?.toString() ?? '';
    String description = notification['body']?.toString() ?? '';
    String time = _formatTime(notification['created_at']?.toString() ?? '');

    // Icônes et couleurs selon le type FCM
    switch (type) {
      case 'auction_pending':
      case 'auction_approved':
      case 'auction_rejected':
      case 'auction_ended':
      case 'auction_won':
      case 'bid_outbid':
        icon = Icons.gavel_outlined;
        color = const Color(0xFF0081FF);
        break;
      case 'payment_received':
      case 'deposit_confirmed':
      case 'deposit_rejected':
      case 'withdrawal_processed':
        icon = Icons.payment_outlined;
        color = const Color(0xFF00C58D);
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
    final String? auctionId = notification['auction_id']?.toString()
        ?? notification['auctionId']?.toString()
        ?? (notification['data'] is Map ? (notification['data'] as Map)['auctionId']?.toString() : null)
        ?? (notification['data'] is Map ? (notification['data'] as Map)['auction_id']?.toString() : null);
    final String? transactionId = notification['transaction_id']?.toString()
        ?? (notification['data'] is Map ? (notification['data'] as Map)['transaction_id']?.toString() : null);

    
    // Marquer comme lu
    _markAsRead(notification['id']?.toString());

    switch (type) {
      // Customer feedback #11: same authoritative-win reasoning as
      // notification_handler.dart's _navigateFromNotification -- auction_won
      // routes to the congratulations page here too, so tapping it from the
      // in-app notification list behaves the same as tapping the push itself.
      case 'auction_won':
        if (auctionId != null) {
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => AuctionWinnerPage(auctionId: auctionId),
            ),
          );
        } else {
          // Customer #22: this exact type already has a meaningful,
          // specialized destination -- only fall back to the generic detail
          // screen when the specific reference (auctionId) this case needs
          // is actually missing, never degrade a normally-working
          // auction_won notification just because it happens to share a
          // switch arm with the new fallback logic below.
          _openGenericDetail(notification);
        }
        break;
      case 'auction_pending':
      case 'auction_approved':
      case 'auction_rejected':
      case 'auction_ended':
      case 'bid_outbid':
        if (auctionId != null) {
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => AuctionDetailsPage(auctionId: auctionId),
            ),
          );
        } else {
          _openGenericDetail(notification);
        }
        break;
      case 'payment_received':
      case 'deposit_confirmed':
      case 'deposit_rejected':
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (context) => const DepositPage(),
          ),
        );
        break;
      // MAZADPAY (withdrawal notification routing bug): same fix as
      // notification_handler.dart's push-tap routing -- withdrawal_processed
      // must never open DepositPage. Routes to the withdrawal's own detail
      // screen using the real transaction_id; falls back to the wallet
      // screen only if that id is unexpectedly missing.
      case 'withdrawal_processed':
        if (transactionId != null && transactionId.isNotEmpty) {
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => WithdrawalDetailPage(transactionId: transactionId),
            ),
          );
        } else {
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => const DepositPage(),
            ),
          );
        }
        break;
      // Customer #22: admin broadcast notifications (general/new_auction/
      // transaction) have no real entity/target today -- see the audit --
      // so they open the generic Notification Detail page instead of
      // falling through to a silent no-op. If a future admin-sent
      // new_auction ever DOES carry a real auctionId, prefer that
      // specialized navigation over the generic page (routing precedence:
      // specialized > reference > generic fallback).
      case 'general':
      case 'transaction':
      case 'new_auction':
        if (auctionId != null) {
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => AuctionDetailsPage(auctionId: auctionId),
            ),
          );
        } else {
          _openGenericDetail(notification);
        }
        break;
      default:
        // Customer #22: any other/unrecognized type with real content
        // (title+body) still deserves to be openable rather than silently
        // doing nothing -- this is strictly additive: every type this repo
        // already knew how to navigate for is handled by a case above and
        // never reaches this branch.
        _openGenericDetail(notification);
        break;
    }
  }

  void _openGenericDetail(Map<String, dynamic> notification) {
    final createdAtRaw = notification['created_at']?.toString();
    final createdAt = createdAtRaw != null
        ? (DateTime.tryParse(createdAtRaw) ?? DateTime.now())
        : DateTime.now();
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => NotificationDetailPage(
          title: notification['title']?.toString() ?? '',
          body: notification['body']?.toString(),
          imageUrl: notification['image_url']?.toString(),
          createdAt: createdAt,
        ),
      ),
    );
  }

  Future<void> _markAsRead(String? notificationId) async {
    if (notificationId == null) return;
    try {
      await _notificationApi.markNotificationAsRead(notificationId);
      setState(() {
        final index = _notifications.indexWhere((n) => n['id']?.toString() == notificationId);
        if (index != -1) {
          _notifications[index]['is_read'] = true;
        }
      });
    } catch (e) {
      debugPrint('Error marking notification as read: $e');
    }
  }
}
