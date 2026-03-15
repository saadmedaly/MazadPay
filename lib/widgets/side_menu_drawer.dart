import 'package:flutter/material.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/pages/account_page.dart';
import 'package:mezadpay/pages/services_page.dart';
import 'package:mezadpay/pages/language_page.dart';
import 'package:mezadpay/l10n/app_localizations.dart';

class SideMenuDrawer extends StatelessWidget {
  const SideMenuDrawer({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final l10n = AppLocalizations.of(context)!;

    return Drawer(
      backgroundColor: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
      child: SafeArea(
        child: Column(
          children: [
            // Header Profile Section
            Padding(
              padding: const EdgeInsets.all(20.0),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  IconButton(
                    icon: const Icon(Icons.close),
                    onPressed: () => Navigator.of(context).pop(),
                  ),
                  Row(
                    children: [
                      const Text(
                        'بدال سيديا',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(width: 12),
                      const CircleAvatar(
                        radius: 24,
                        backgroundImage: AssetImage('user.png'),
                      ),
                    ],
                  ),
                ],
              ),
            ),

            // Deposit Banner
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20.0),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [Color(0xFF0084FF), Color(0xFF0055FF)],
                    begin: Alignment.centerLeft,
                    end: Alignment.centerRight,
                  ),
                  borderRadius: BorderRadius.circular(16),
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
                                color: Colors.white70,
                                fontSize: 12,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(width: 12),
                        const Icon(
                          Icons.account_balance_wallet,
                          color: Colors.white,
                          size: 32,
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 20),

            // Menu Items
            Expanded(
              child: ListView(
                padding: EdgeInsets.zero,
                children: [
                  _buildMenuItem(
                    context,
                    title: l10n.home,
                    icon: Icons.home,
                  ),
                  _buildMenuItem(
                    context,
                    title: l10n.delivery,
                    icon: Icons.local_shipping_outlined,
                  ),
                  _buildMenuItem(
                    context,
                    title: l10n.myAccount,
                    icon: Icons.person_outline,
                  ),
                  const Divider(height: 32, thickness: 1, indent: 20, endIndent: 20),
                  _buildMenuItem(
                    context,
                    title: l10n.personalInfo,
                    icon: Icons.person_outline,
                  ),
                  _buildMenuItem(
                    context,
                    title: l10n.changeLanguage,
                    icon: Icons.language,
                  ),
                  _buildMenuItem(
                    context,
                    title: l10n.contactUsSide,
                    icon: Icons.phone_outlined,
                  ),
                  const SizedBox(height: 20),

                  // Enhanced Action Buttons
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 20.0),
                    child: Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: isDarkMode ? const Color(0xFF2D2D2D) : const Color(0xFFFFF8E1),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          ElevatedButton.icon(
                            onPressed: () {},
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.red,
                              foregroundColor: Colors.white,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(20),
                              ),
                              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                              minimumSize: Size.zero,
                            ),
                            icon: const Icon(Icons.star_border, size: 18),
                            label: Text(l10n.rateApp, style: const TextStyle(fontWeight: FontWeight.bold)),
                          ),
                          Text(
                            l10n.shareOpinion,
                            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                          ),
                        ],
                      ),
                    ),
                  ),

                  const SizedBox(height: 16),
                  _buildMenuItem(
                    context,
                    title: l10n.termsConditions,
                    icon: Icons.article_outlined,
                    isCompact: true,
                  ),
                  _buildMenuItem(
                    context,
                    title: l10n.helpFaq,
                    icon: Icons.help_outline,
                    isCompact: true,
                  ),
                  _buildMenuItem(
                    context,
                    title: l10n.aboutMazad,
                    icon: Icons.info_outline,
                    isCompact: true,
                  ),

                  // Share App Section
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 20.0, vertical: 16.0),
                    child: Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: isDarkMode ? const Color(0xFF2D2D2D) : const Color(0xFFFFF8E1),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          ElevatedButton.icon(
                            onPressed: () {},
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.black,
                              foregroundColor: Colors.white,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(20),
                              ),
                              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                              minimumSize: Size.zero,
                            ),
                            icon: const Icon(Icons.share, size: 16),
                            label: Text(l10n.shareApp, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
                          ),
                          Expanded(
                            child: Text(
                              l10n.shareAppDesc,
                              textAlign: TextAlign.end,
                              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
            
            // Footer Social Icons
            Padding(
              padding: const EdgeInsets.only(bottom: 20.0),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  _buildSocialIcon(FontAwesomeIcons.facebookF, const Color(0xFF1877F2)),
                  const SizedBox(width: 20),
                  _buildSocialIcon(FontAwesomeIcons.instagram, const Color(0xFFE4405F)),
                  const SizedBox(width: 20),
                  _buildSocialIcon(FontAwesomeIcons.tiktok, const Color(0xFF000000)),
                  const SizedBox(width: 20),
                  _buildSocialIcon(FontAwesomeIcons.snapchat, const Color(0xFFFFFC00), isYellow: true),
                ],
              ),
            ),
            const Text(
              'Version 3.2.0',
              style: TextStyle(color: Colors.grey, fontSize: 12, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }

  Widget _buildMenuItem(
    BuildContext context, {
    required String title,
    required IconData icon,
    bool isSelected = false,
    bool isCompact = false,
  }) {
    Color primaryBlue = const Color(0xFF0084FF);
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final l10n = AppLocalizations.of(context)!;
    
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 20.0, vertical: isCompact ? 4.0 : 8.0),
      child: InkWell(
        onTap: () {
          Navigator.pop(context); // Close drawer first
          if (title == l10n.home) {
            Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const HomePage()));
          } else if (title == l10n.delivery) {
            Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const ServicesPage()));
          } else if (title == l10n.myAccount || title == l10n.personalInfo) {
            Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const AccountPage()));
          } else if (title == l10n.changeLanguage) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const LanguagePage()));
          }
        },
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 8.0),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Icon(
                    icon,
                    color: isSelected ? primaryBlue : Colors.grey[600],
                    size: isCompact ? 20 : 24,
                  ),
                  const SizedBox(width: 16),
                  Text(
                    title,
                    style: TextStyle(
                      fontSize: isCompact ? 14 : 16,
                      fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                      color: isSelected ? primaryBlue : (isDarkMode ? Colors.white : Colors.black87),
                    ),
                  ),
                ],
              ),
              const Icon(Icons.arrow_forward_ios, size: 14, color: Colors.grey),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSocialIcon(IconData icon, Color color, {bool isYellow = false}) {
    return Container(
      width: 44,
      height: 44,
      decoration: BoxDecoration(
        color: isYellow ? color : color.withValues(alpha: 0.1),
        shape: BoxShape.circle,
        border: Border.all(color: color.withValues(alpha: 0.2), width: 1),
      ),
      child: Center(
        child: FaIcon(
          icon, 
          color: isYellow ? Colors.black : color, 
          size: 20
        ),
      ),
    );
  }
}
