import 'package:flutter/material.dart';
import 'package:mezadpay/widgets/side_menu_drawer.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/pages/services_page.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/widgets/custom_app_bar.dart';
import 'deposit_page.dart';
import 'account_profile_page.dart';

class AccountPage extends StatefulWidget {
  const AccountPage({super.key});

  @override
  State<AccountPage> createState() => _AccountPageState();
}

class _AccountPageState extends State<AccountPage> {
  int _currentIndex = 3;
  bool _showBalance = false;

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final l10n = AppLocalizations.of(context)!;
    const Color primaryBlue = Color(0xFF0084FF);

    return Scaffold(
      backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
      drawer: const SideMenuDrawer(),
      appBar: CustomAppBar(
        title: l10n.myAccount,
        showBack: true,
        onBackPress: () {
          Navigator.pushReplacement(
            context,
            MaterialPageRoute(builder: (context) => const HomePage()),
          );
        },
      ),
      body: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20.0),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 16),
              
              // Insurance Card
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(24),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [Color(0xFF0084FF), Color(0xFF0055FF)],
                    begin: Alignment.centerRight,
                    end: Alignment.centerLeft,
                  ),
                  borderRadius: BorderRadius.circular(16),
                  boxShadow: [
                    BoxShadow(
                      color: primaryBlue.withValues(alpha: 0.3),
                      blurRadius: 10,
                      offset: const Offset(0, 5),
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          l10n.insuranceAmount,
                          style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold),
                        ),
                        const Icon(Icons.more_horiz, color: Colors.white),
                      ],
                    ),
                    const SizedBox(height: 24),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              l10n.availableBalance,
                              style: const TextStyle(color: Colors.white70, fontSize: 12),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              _showBalance ?  'MRU 500.00' : '••••••' ,
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 24,
                                fontWeight: FontWeight.bold,
                              ),
                              textDirection: TextDirection.ltr,
                            ),
                          ],
                        ),
                        IconButton(
                          icon: Icon(
                              _showBalance ?    Icons.visibility_off_outlined : Icons.visibility_outlined,
                            color: Colors.white,
                          ),
                          onPressed: () {
                            setState(() {
                              _showBalance = !_showBalance;
                            });
                          },
                        ),
                      ],
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),
              Text(
                l10n.chooseMethod,
                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
              ),
              const SizedBox(height: 16),

              // Deposit Now Button (Red)
              InkWell(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => const DepositPage()),
                  );
                },
                borderRadius: BorderRadius.circular(12),
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [Color(0xFFE50000), Color(0xFFFF4040)],
                      begin: Alignment.centerRight,
                      end: Alignment.centerLeft,
                    ),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Icon(
                        Icons.arrow_back_ios,
                        color: Colors.white,
                        size: 20,
                      ),
                      Row(
                        children: [
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Text(
                                l10n.depositNow,
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 18,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                l10n.startBiddingJourney,
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 12,
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(width: 16),
                          const Icon(
                            Icons.account_balance_wallet,
                            color: Colors.white,
                            size: 40,
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 16),

              // Retrieve Insurance Button (Green)
              InkWell(
                onTap: () {},
                borderRadius: BorderRadius.circular(12),
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [Color(0xFF00AA00), Color(0xFF33CC33)],
                      begin: Alignment.centerRight,
                      end: Alignment.centerLeft,
                    ),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Icon(
                        Icons.arrow_back_ios,
                        color: Colors.white,
                        size: 20,
                      ),
                      Row(
                        children: [
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              Text(
                                l10n.retrieveInsurance,
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 18,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                l10n.startBiddingJourney,
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 12,
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(width: 16),
                          const Icon(
                            Icons.account_balance_wallet,
                            color: Colors.white,
                            size: 40,
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),

              const SizedBox(height: 32),
              Text(
                l10n.yourActivities,
                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
              ),
              const SizedBox(height: 16),

              // Activity Grid
              Row(
                children: [
                  Expanded(
                    child: _buildActivityItem(
                      icon: Icons.gavel,
                      title: l10n.myAuctions,
                      isDarkMode: isDarkMode,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _buildActivityItem(
                      icon: Icons.favorite_border,
                      title: l10n.favorites,
                      iconColor: Colors.red,
                      isDarkMode: isDarkMode,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _buildActivityItem(
                      icon: Icons.emoji_events_outlined,
                      title: l10n.wonItems,
                      isDarkMode: isDarkMode,
                    ),
                  ),
                ],
              ),
              
              const SizedBox(height: 24),
              Center(
                child: Column(
                  children: [
                    Text(
                      l10n.communicationCenter,
                      style: TextStyle(color: Colors.grey[600], fontSize: 12),
                    ),
                    const SizedBox(height: 4),
                    InkWell(
                      onTap: () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(builder: (context) => const AccountProfilePage()),
                        );
                      },
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(Icons.person_outline, size: 16),
                          const SizedBox(width: 4),
                          Text(
                            l10n.personalAccountInfo,
                            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12, color: Color(0xFF0084FF)),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 40), // Space for bottom nav
            ],
          ),
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

  Widget _buildActivityItem({required IconData icon, required String title, Color? iconColor, required bool isDarkMode}) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 8),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.withValues(alpha: 0.2)),
        boxShadow: [
          if (!isDarkMode)
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.02),
              blurRadius: 5,
              offset: const Offset(0, 2),
            ),
        ],
      ),
      child: Column(
        children: [
          Icon(icon, color: iconColor ?? (isDarkMode ? Colors.white : Colors.black87), size: 28),
          const SizedBox(height: 12),
          Text(
            title,
            textAlign: TextAlign.center,
            style: const TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.bold,
            ),
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
          Navigator.pushReplacement(
            context,
            MaterialPageRoute(builder: (context) => const HomePage()),
          );
        } else if (index == 1) {
          Navigator.pushReplacement(
            context,
            MaterialPageRoute(builder: (context) => const ServicesPage()),
          );
        } else if (index == 3) {
          setState(() => _currentIndex = index);
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
