import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class NotificationsPage extends StatelessWidget {
  const NotificationsPage({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          title: Text(
            'الإشعارات',
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
              onPressed: () {},
              child: Text(
                'تحديد ككل كمقروء',
                style: GoogleFonts.plusJakartaSans(color: const Color(0xFF0081FF), fontSize: 12),
              ),
            ),
          ],
        ),
        body: ListView.separated(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
          itemCount: 5,
          separatorBuilder: (context, index) => const SizedBox(height: 12),
          itemBuilder: (context, index) {
            return _buildNotificationItem(index, isDarkMode);
          },
        ),
      ),
    );
  }

  Widget _buildNotificationItem(int index, bool isDarkMode) {
    final types = ['bid', 'win', 'system', 'payment', 'ad'];
    final type = types[index % types.length];
    
    IconData icon;
    Color color;
    String title;
    String description;
    String time = 'منذ 5 دقائق';

    switch (type) {
      case 'bid':
        icon = Icons.gavel_outlined;
        color = Colors.orange;
        title = 'تنبيه مزايدة';
        description = 'لقد تم تجاوز عرضك في مزاد Toyota Corolla.';
        break;
      case 'win':
        icon = Icons.emoji_events_outlined;
        color = const Color(0xFF00C58D);
        title = 'مبروك الفوز!';
        description = 'لقد فزت بمزاد iPhone 15 Pro Max.';
        break;
      case 'payment':
        icon = Icons.payment_outlined;
        color = const Color(0xFF0081FF);
        title = 'تأكيد الدفع';
        description = 'تم استلام دفعة مبلغ التأمين بنجاح.';
        break;
      case 'system':
        icon = Icons.notifications_active_outlined;
        color = Colors.blueGrey;
        title = 'تحديث النظام';
        description = 'هناك ميزات جديدة متاحة في التطبيق الآن.';
        break;
      default:
        icon = Icons.info_outline;
        color = Colors.grey;
        title = 'تحديث الإعلان';
        description = 'تمت الموافقة على إعلانك ونشره في التطبيق.';
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
                    Text(title, style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.bold, fontSize: 14)),
                    Text(time, style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 10)),
                  ],
                ),
                const SizedBox(height: 4),
                Text(
                  description,
                  style: GoogleFonts.plusJakartaSans(color: isDarkMode ? Colors.grey[400] : Colors.grey[600], fontSize: 12),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
