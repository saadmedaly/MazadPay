import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class DeliveryDetailsPage extends StatelessWidget {
  const DeliveryDetailsPage({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          title: Text(
            'تفاصيل التوصيل',
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
        ),
        body: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildTrackingCard(isDarkMode),
              const SizedBox(height: 32),
              Text(
                'حالة التوصيل',
                style: GoogleFonts.plusJakartaSans(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 24),
              _buildTimeline(isDarkMode),
              const SizedBox(height: 32),
              _buildAddressCard(isDarkMode),
              const SizedBox(height: 32),
              _buildDeliveryGuyCard(isDarkMode),
            ],
          ),
        ),
      );
  }

  Widget _buildTrackingCard(bool isDarkMode) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFF0081FF), Color(0xFF0055FF)],
          begin: AlignmentDirectional.topStart,
          end: AlignmentDirectional.bottomEnd,
        ),
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(color: const Color(0xFF0081FF).withOpacity(0.3), blurRadius: 15, offset: const Offset(0, 8)),
        ],
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('رقم التتبع', style: GoogleFonts.plusJakartaSans(color: Colors.white70, fontSize: 12)),
              Text('MP-88294-2024', style: GoogleFonts.plusJakartaSans(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14)),
            ],
          ),
          const SizedBox(height: 20),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('الوصول المتوقع', style: GoogleFonts.plusJakartaSans(color: Colors.white70, fontSize: 12)),
                  Text('اليوم، 04:30 م', style: GoogleFonts.plusJakartaSans(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 18)),
                ],
              ),
              const Icon(Icons.local_shipping, color: Colors.white, size: 40),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildTimeline(bool isDarkMode) {
    return Column(
      children: [
        _buildTimelineStep('تم استلام الطلب', 'تم تأكيد طلبك بنجاح.', '10:00 ص', true, true, isDarkMode),
        _buildTimelineStep('في مرحلة التجهيز', 'يتم فحص المنتج وتجهيزه للشحن.', '11:20 ص', true, true, isDarkMode),
        _buildTimelineStep('تم الشحن', 'الطلب في طريقه إلى مدينتك.', '01:45 م', true, false, isDarkMode),
        _buildTimelineStep('قيد التوصيل', 'المندوب في طريقه إلى عنوانك.', 'قريباً', false, false, isDarkMode),
      ],
    );
  }

  Widget _buildTimelineStep(String title, String subtitle, String time, bool isCompleted, bool isLast, bool isDarkMode) {
    return IntrinsicHeight(
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Column(
            children: [
              Container(
                width: 24,
                height: 24,
                decoration: BoxDecoration(
                  color: isCompleted ? const Color(0xFF00C58D) : Colors.grey[300],
                  shape: BoxShape.circle,
                  border: Border.all(color: Colors.white, width: 4),
                ),
                child: isCompleted ? const Icon(Icons.check, size: 12, color: Colors.white) : null,
              ),
              if (isLast)
                Expanded(
                  child: Container(width: 2, color: isCompleted ? const Color(0xFF00C58D) : Colors.grey[300]),
                ),
            ],
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                 Row(
                   mainAxisAlignment: MainAxisAlignment.spaceBetween,
                   children: [
                     Text(title, style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.bold, fontSize: 14, color: isCompleted ? (isDarkMode ? Colors.white : Colors.black) : Colors.grey)),
                     Text(time, style: GoogleFonts.plusJakartaSans(fontSize: 12, color: Colors.grey)),
                   ],
                 ),
                 const SizedBox(height: 4),
                 Text(subtitle, style: GoogleFonts.plusJakartaSans(fontSize: 12, color: Colors.grey[600])),
                 const SizedBox(height: 24),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildAddressCard(bool isDarkMode) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.withOpacity(0.1)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(color: const Color(0xFF0081FF).withOpacity(0.1), borderRadius: BorderRadius.circular(12)),
            child: const Icon(Icons.location_on, color: Color(0xFF0081FF)),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('عنوان التوصيل', style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 12)),
                const SizedBox(height: 4),
                Text('نواكشوط، تفرغ زينة، شارع المختار ولد داداه', style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.bold, fontSize: 14)),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDeliveryGuyCard(bool isDarkMode) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.withOpacity(0.1)),
      ),
      child: Row(
        children: [
          const CircleAvatar(
            radius: 24,
            backgroundImage: AssetImage('assets/user.png'),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('محمد الأمين', style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.bold, fontSize: 14)),
                Text('مندوب التوصيل', style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 12)),
              ],
            ),
          ),
          IconButton(
            icon: const Icon(Icons.phone, color: Color(0xFF00C58D)),
            onPressed: () {},
          ),
          IconButton(
            icon: const Icon(Icons.chat_bubble, color: Color(0xFF0081FF)),
            onPressed: () {},
          ),
        ],
      ),
    );
  }
}
