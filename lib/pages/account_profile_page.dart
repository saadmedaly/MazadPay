import 'package:flutter/material.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'account_page.dart';

class AccountProfilePage extends StatelessWidget {
  const AccountProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    const Color primaryBlue = Color(0xFF0084FF);

    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFF5F7FA),
        endDrawer: const SideMenuDrawer(),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          toolbarHeight: 70,
          automaticallyImplyLeading: false,
          title: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  IconButton(
                    icon: Icon(Icons.arrow_forward_ios, color: isDarkMode ? Colors.white : Colors.black, size: 20),
                    onPressed: () => Navigator.pop(context),
                  ),
                  Text(
                    'معلومات الحساب',
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18, color: isDarkMode ? Colors.white : Colors.black),
                  ),
                ],
              ),
              Builder(
                builder: (context) => IconButton(
                  icon: Icon(Icons.menu, color: isDarkMode ? Colors.white : Colors.black, size: 28),
                  onPressed: () => Scaffold.of(context).openEndDrawer(),
                ),
              ),
            ],
          ),
        ),
        body: SingleChildScrollView(
          child: Column(
            children: [
              // Profile Header
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(24),
                child: Column(
                  children: [
                    Stack(
                      children: [
                        Container(
                          width: 90,
                          height: 90,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            boxShadow: [BoxShadow(color: primaryBlue.withOpacity(0.3), blurRadius: 12, offset: const Offset(0, 4))],
                          ),
                          child: ClipOval(
                            child: Image.asset(
                              'user.png',
                              fit: BoxFit.cover,
                              errorBuilder: (c, e, s) => Container(
                                decoration: const BoxDecoration(
                                  shape: BoxShape.circle,
                                  gradient: LinearGradient(
                                    colors: [Color(0xFF0055FF), Color(0xFF0084FF)],
                                    begin: Alignment.topLeft,
                                    end: Alignment.bottomRight,
                                  ),
                                ),
                                child: const Center(child: Text('ب', style: TextStyle(color: Colors.white, fontSize: 36, fontWeight: FontWeight.bold))),
                              ),
                            ),
                          ),
                        ),
                        Positioned(
                          bottom: 0,
                          left: 0,
                          child: Container(
                            width: 28,
                            height: 28,
                            decoration: const BoxDecoration(color: primaryBlue, shape: BoxShape.circle),
                            child: const Icon(Icons.camera_alt, color: Colors.white, size: 14),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    const Text('بدال سيديا', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 20)),
                    const SizedBox(height: 4),
                    Text('+222 20 00 00 00', style: TextStyle(color: Colors.grey[500], fontSize: 14)),
                  ],
                ),
              ),

              // Profile Info Fields
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('معلوماتي', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                    const SizedBox(height: 16),
                    _buildInfoField(context, 'الاسم الكامل', 'بدال سيديا', Icons.person_outline, isDarkMode),
                    _buildInfoField(context, 'رقم الهاتف', '+222 20 00 00 00', Icons.phone_outlined, isDarkMode),
                    _buildInfoField(context, 'البريد الإلكتروني', 'badal@example.com', Icons.email_outlined, isDarkMode),
                    _buildInfoField(context, 'المدينة', 'نواكشوط', Icons.location_city_outlined, isDarkMode),

                    const SizedBox(height: 32),
                    const Text('الإعدادات', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                    const SizedBox(height: 16),

                    _buildSettingTile(context, 'تغيير كلمة المرور', Icons.lock_outline, isDarkMode),
                    _buildSettingTile(context, 'اللغة', Icons.language, isDarkMode, trailing: 'العربية'),
                    _buildSettingTile(context, 'الإشعارات', Icons.notifications_outlined, isDarkMode, hasSwitch: true),

                    const SizedBox(height: 32),

                    // Logout Button
                    SizedBox(
                      width: double.infinity,
                      height: 54,
                      child: OutlinedButton(
                        onPressed: () {
                          Navigator.pushAndRemoveUntil(
                            context,
                            MaterialPageRoute(builder: (context) => const AccountPage()),
                            (route) => false,
                          );
                        },
                        style: OutlinedButton.styleFrom(
                          side: const BorderSide(color: Colors.red, width: 1.5),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                        ),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: const [
                            Icon(Icons.logout, color: Colors.red),
                            SizedBox(width: 8),
                            Text('تسجيل الخروج', style: TextStyle(color: Colors.red, fontWeight: FontWeight.bold, fontSize: 16)),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 24),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildInfoField(BuildContext context, String label, String value, IconData icon, bool isDarkMode) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.withOpacity(0.15)),
      ),
      child: Row(
        children: [
          Icon(icon, color: const Color(0xFF0084FF), size: 20),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(label, style: TextStyle(color: Colors.grey[500], fontSize: 11)),
                const SizedBox(height: 2),
                Text(value, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
              ],
            ),
          ),
          Icon(Icons.edit_outlined, color: Colors.grey[400], size: 16),
        ],
      ),
    );
  }

  Widget _buildSettingTile(BuildContext context, String title, IconData icon, bool isDarkMode, {String? trailing, bool hasSwitch = false}) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.withOpacity(0.15)),
      ),
      child: ListTile(
        leading: Container(
          width: 36,
          height: 36,
          decoration: BoxDecoration(
            color: const Color(0xFF0084FF).withOpacity(0.1),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, color: const Color(0xFF0084FF), size: 18),
        ),
        title: Text(title, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
        trailing: hasSwitch
            ? Switch(
                value: true,
                onChanged: (_) {},
                activeThumbColor: const Color(0xFF0084FF),
              )
            : trailing != null
                ? Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(trailing, style: TextStyle(color: Colors.grey[500], fontSize: 12)),
                      const SizedBox(width: 4),
                      Icon(Icons.arrow_back_ios, size: 14, color: Colors.grey[400]),
                    ],
                  )
                : Icon(Icons.arrow_back_ios, size: 14, color: Colors.grey[400]),
        onTap: () {},
      ),
    );
  }
}
