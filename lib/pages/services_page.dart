import 'package:flutter/material.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/widgets/custom_app_bar.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/pages/account_page.dart';

class ServicesPage extends StatefulWidget {
  const ServicesPage({super.key});

  @override
  State<ServicesPage> createState() => _ServicesPageState();
}

class _ServicesPageState extends State<ServicesPage> {
  int _currentIndex = 1;

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
      drawer: const SideMenuDrawer(),
      appBar: CustomAppBar(
        showBack: true,
        onBackPress: () => Navigator.pushReplacement(
          context,
          MaterialPageRoute(builder: (context) => const HomePage()),
        ),
      ),
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Promo Banner
            Padding(
              padding: const EdgeInsets.all(20.0),
              child: Container(
                width: double.infinity,
                height: 100,
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [Color(0xFF0084FF), Color(0xFF0055FF)],
                    begin: Alignment.centerRight,
                    end: Alignment.centerLeft,
                  ),
                  borderRadius: BorderRadius.circular(16),
                ),
                child: Padding(
                  padding: const EdgeInsets.all(20.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        l10n.earnTime,
                        style: const TextStyle(color: Colors.white, fontSize: 24, fontWeight: FontWeight.bold),
                      ),
                    ],
                  ),
                ),
              ),
            ),

            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20.0),
              child: Text(
                l10n.services,
                style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
              ),
            ),
            const SizedBox(height: 16),

            // Services Grid
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20.0),
              child: GridView.count(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                crossAxisCount: 2,
                crossAxisSpacing: 16,
                mainAxisSpacing: 16,
                childAspectRatio: 1.2,
                children: [
                  _buildServiceCard(l10n.deliveryService, Colors.blue[50]!, Icons.motorcycle, Colors.blue),
                  _buildServiceCard(l10n.course, Colors.yellow[50]!, Icons.local_taxi, Colors.orange),
                  _buildServiceCard(l10n.intercityCourse, Colors.green[50]!, Icons.directions_bus, Colors.green),
                  _buildServiceCard(l10n.goodsTransport, Colors.orange[50]!, Icons.local_shipping, Colors.orangeAccent),
                  _buildServiceCard(l10n.other, Colors.purple[50]!, Icons.more_horiz, Colors.purple),
                ],
              ),
            ),
            const SizedBox(height: 40),
          ],
        ),
      ),
      floatingActionButton: Padding(
        padding: const EdgeInsets.only(top: 20.0),
        child: SizedBox(
          height: 70,
          width: 70,
          child: FloatingActionButton(
            onPressed: () {},
            backgroundColor: Colors.transparent,
            elevation: 0,
            highlightElevation: 0,
            child: Image.asset(
              'botum_bar.png',
              fit: BoxFit.contain,
            ),
          ),
        ),
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
      bottomNavigationBar: BottomAppBar(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        shape: const CircularNotchedRectangle(),
        notchMargin: 8,
        child: SizedBox(
          height: 60,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _buildNavItem(Icons.home_outlined, Icons.home, l10n.home, 0),
              _buildNavItem(Icons.local_shipping_outlined, Icons.local_shipping, l10n.delivery, 1),
              const SizedBox(width: 48), // Space for FAB
              _buildNavItem(Icons.storefront_outlined, Icons.storefront, l10n.ecommerce, 2),
              _buildNavItem(Icons.person_outline, Icons.person, l10n.myAccount, 3),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildServiceCard(String title, Color bgColor, IconData icon, Color iconColor) {
    return Container(
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, size: 48, color: iconColor),
          const SizedBox(height: 12),
          Text(
            title,
            textAlign: TextAlign.center,
            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
          ),
        ],
      ),
    );
  }

  Widget _buildNavItem(IconData icon, IconData activeIcon, String label, int index) {
    bool isSelected = _currentIndex == index;
    const Color primaryBlue = Color(0xFF0084FF);
    return InkWell(
      onTap: () {
        if (index == 0) {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const HomePage()));
        } else if (index == 1) {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const ServicesPage()));
        } else if (index == 3) {
          Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const AccountPage()));
        } else {
          setState(() => _currentIndex = index);
        }
      },
      child: Column(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(isSelected ? activeIcon : icon, color: isSelected ? primaryBlue : Colors.grey[600]),
          const SizedBox(height: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 10,
              fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
              color: isSelected ? primaryBlue : Colors.grey[600],
            ),
          ),
        ],
      ),
    );
  }
}
