import 'package:flutter/material.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/pages/services_page.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/pages/create_ad_start_page.dart';
import 'package:mezadpay/pages/my_auctions_shell_page.dart';
import 'package:mezadpay/pages/account_shell_page.dart';
import 'package:mezadpay/pages/notifications_page.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';

class ServicesShellPage extends StatelessWidget {
  const ServicesShellPage({super.key});

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
      body: const ServicesPage(),
      floatingActionButton: _buildFab(context),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
      bottomNavigationBar: _ShellBottomNav(activeIndex: 1),
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

class _ShellBottomNav extends StatelessWidget {
  final int activeIndex;
  const _ShellBottomNav({required this.activeIndex});

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
            _buildItem(context, Icons.home_outlined, Icons.home, AppLocalizations.of(context)!.text_1, 0),
            _buildItem(context, Icons.local_shipping_outlined, Icons.local_shipping, AppLocalizations.of(context)!.text_32, 1),
            const SizedBox(width: 48),
            _buildItem(context, Icons.gavel_outlined, Icons.gavel, AppLocalizations.of(context)!.text_27, 2),
            _buildItem(context, Icons.person_outline, Icons.person, AppLocalizations.of(context)!.text_19, 3),
          ],
        ),
      ),
    );
  }

  Widget _buildItem(BuildContext context, IconData icon, IconData activeIcon, String label, int index) {
    bool isSelected = activeIndex == index;
    const Color primaryBlue = Color(0xFF0084FF);

    return InkWell(
      onTap: () {
        if (index == 2) {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => MyAuctionsShellPage()));
        } else if (index == activeIndex) {
          return;
        } else {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) {
            if (index == 0) return const HomePage();
            if (index == 1) return ServicesShellPage();
            if (index == 3) return AccountShellPage();
            return const HomePage();
          }));
        }
      },
      child: Column(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(isSelected ? activeIcon : icon, color: isSelected ? primaryBlue : Colors.grey[600]),
          const SizedBox(height: 4),
          Text(label, style: TextStyle(fontSize: 10, fontWeight: isSelected ? FontWeight.bold : FontWeight.normal, color: isSelected ? primaryBlue : Colors.grey[600])),
        ],
      ),
    );
  }
}
