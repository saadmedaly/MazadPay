import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:mezadpay/pages/delivery_details_page.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'home_page.dart';
import 'account_page.dart';
import 'notifications_page.dart';
import 'create_ad_start_page.dart';

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
    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        endDrawer: const SideMenuDrawer(),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          toolbarHeight: 70,
          automaticallyImplyLeading: false,
          title: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              IconButton(
                icon: Icon(Icons.arrow_forward_ios, color: isDarkMode ? Colors.white : Colors.black, size: 28),
                onPressed: () => Navigator.pushReplacement(
                  context,
                  MaterialPageRoute(builder: (context) => const HomePage()),
                ),
              ),
              IconButton(
                icon: Icon(Icons.notifications_outlined, color: isDarkMode ? Colors.white : Colors.black, size: 28),
                onPressed: () {
                   Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => const NotificationsPage()),
                  );
                },
              ),
            ],
          ),
        ),
        body: _currentIndex == 2 
          ? Center(
              child: Text(
                AppLocalizations.of(context)!.text_34,
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.grey),
              ),
            )
          : SingleChildScrollView(
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
                      begin: AlignmentDirectional.centerEnd,
                      end: AlignmentDirectional.centerStart,
                    ),
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Stack(
                    children: [
                      Padding(
                        padding: const EdgeInsets.all(20.0),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Text(
                              AppLocalizations.of(context)!.text_296,
                              style: TextStyle(color: Colors.white, fontSize: 24, fontWeight: FontWeight.bold),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),

              Padding(
                padding: EdgeInsets.symmetric(horizontal: 20.0),
                child: Text(
                  AppLocalizations.of(context)!.text_297,
                  style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
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
                    _buildServiceCard(
                      AppLocalizations.of(context)!.text_32, 
                      Colors.blue[50]!, 
                      Icons.local_shipping_outlined, 
                      const Color(0xFF00C58D),
                      onTap: () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(builder: (context) => const DeliveryDetailsPage()),
                        );
                      },
                    ),
                    _buildServiceCard(AppLocalizations.of(context)!.text_298, Colors.yellow[50]!, Icons.local_taxi, Colors.orange),
                    _buildServiceCard(AppLocalizations.of(context)!.text_299, Colors.green[50]!, Icons.directions_bus, Colors.green),
                    _buildServiceCard(AppLocalizations.of(context)!.text_300, Colors.orange[50]!, Icons.local_shipping, Colors.orangeAccent),
                    _buildServiceCard(AppLocalizations.of(context)!.text_301, Colors.purple[50]!, Icons.more_horiz, Colors.purple),
                  ],
                ),
              ),
              const SizedBox(height: 40),
            ],
          ),
        ),
        floatingActionButton: Padding(
          padding: const EdgeInsetsDirectional.only(top: 20.0),
          child: SizedBox(
            height: 70,
            width: 70,
            child: FloatingActionButton(
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (context) => const CreateAdStartPage()),
                );
              },
              backgroundColor: Colors.transparent,
              elevation: 0,
              highlightElevation: 0,
              child: Image.asset(
                'assets/botum_bar.png',
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
            height: 70, // Increased height
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildNavItem(Icons.home_outlined, Icons.home, AppLocalizations.of(context)!.text_1, 0),
                _buildNavItem(Icons.local_shipping_outlined, Icons.local_shipping, AppLocalizations.of(context)!.text_32, 1),
                const SizedBox(width: 48), // Space for FAB
                _buildNavItem(Icons.storefront_outlined, Icons.storefront, AppLocalizations.of(context)!.text_33, 2),
                _buildNavItem(Icons.person_outline, Icons.person, AppLocalizations.of(context)!.text_19, 3),
              ],
            ),
          ),
        ),
      );
  }

  Widget _buildServiceCard(String title, Color bgColor, IconData icon, Color iconColor, {VoidCallback? onTap}) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
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
        } else if (index == 2) {
          setState(() => _currentIndex = index);
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