import 'package:flutter/material.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/pages/account_page.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/pages/create_ad_start_page.dart';
import 'package:mezadpay/pages/services_shell_page.dart';
import 'package:mezadpay/pages/my_auctions_shell_page.dart';
import 'package:mezadpay/pages/notifications_page.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';

class AccountShellPage extends StatelessWidget {
  const AccountShellPage({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
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
              textDirection: TextDirection.ltr,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text('M', style: TextStyle(fontSize: 26, fontWeight: FontWeight.w900, fontStyle: FontStyle.italic, color: isDarkMode ? Colors.white : const Color(0xFF135BEC))),
                Text('azad', style: TextStyle(fontSize: 22, fontWeight: FontWeight.w900, color: isDarkMode ? Colors.white : Colors.black)),
                Text('Pay', style: TextStyle(fontSize: 22, fontWeight: FontWeight.w900, color: isDarkMode ? Colors.white : const Color(0xFF135BEC))),
              ],
            ),
            IconButton(
              icon: Icon(Icons.notifications_outlined, color: isDarkMode ? Colors.white : Colors.black, size: 28),
              onPressed: () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const NotificationsPage())),
            ),
          ],
        ),
      ),
      body: const AccountPage(),
      floatingActionButton: _buildFab(context),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
      bottomNavigationBar: _AccountShellBottomNav(),
    );
  }

  Widget _buildFab(BuildContext context) {
    return GestureDetector(
      onTap: () => Navigator.of(context).push(MaterialPageRoute(builder: (context) => const CreateAdStartPage())),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          SizedBox(
            height: 70, width: 70,
            child: FloatingActionButton(
              onPressed: () => Navigator.of(context).push(MaterialPageRoute(builder: (context) => const CreateAdStartPage())),
              backgroundColor: Colors.transparent, elevation: 0, highlightElevation: 0,
              child: Image.asset('assets/botum_bar.png', fit: BoxFit.contain),
            ),
          ),
          Text('إعلان جديد', style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Colors.grey[600])),
        ],
      ),
    );
  }
}

class _AccountShellBottomNav extends StatelessWidget {
  const _AccountShellBottomNav();

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return BottomAppBar(
      color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
      shape: const CircularNotchedRectangle(),
      notchMargin: 8,
      child: SizedBox(
        height: 70,
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: [
            _buildItem(context, Icons.home_outlined, AppLocalizations.of(context)!.text_1, 0),
            _buildItem(context, Icons.local_shipping_outlined, AppLocalizations.of(context)!.text_32, 1),
            const SizedBox(width: 48),
            _buildItem(context, Icons.gavel_outlined, AppLocalizations.of(context)!.text_27, 2),
            _buildItem(context, Icons.person, AppLocalizations.of(context)!.text_19, 3, active: true),
          ],
        ),
      ),
    );
  }

  Widget _buildItem(BuildContext context, IconData icon, String label, int index, {bool active = false}) {
    const Color primaryBlue = Color(0xFF0084FF);

    return InkWell(
      onTap: () {
        if (index == 2) {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => MyAuctionsShellPage()));
        } else if (index == 3) {
          return; // Already on account page
        } else {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) {
            if (index == 0) return const HomePage();
            if (index == 1) return ServicesShellPage();
            return const HomePage();
          }));
        }
      },
      child: Column(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, color: active ? primaryBlue : Colors.grey[600]),
          const SizedBox(height: 4),
          Text(label, style: TextStyle(fontSize: 10, fontWeight: active ? FontWeight.bold : FontWeight.normal, color: active ? primaryBlue : Colors.grey[600])),
        ],
      ),
    );
  }
}
