import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/pages/account_page.dart';
import 'package:mezadpay/pages/services_page.dart';
import 'package:mezadpay/pages/about_mazad_pay_page.dart';
import 'package:mezadpay/pages/how_to_bid_page.dart';
import 'package:mezadpay/pages/my_auctions_page.dart';
import 'package:mezadpay/pages/favorites_page.dart';
import 'package:mezadpay/pages/support_page.dart';
import 'package:mezadpay/pages/all_auctions_page.dart';
import 'package:mezadpay/pages/privacy_policy_page.dart';
import 'package:mezadpay/widgets/app_modals.dart';

class SideMenuDrawer extends StatelessWidget {
  const SideMenuDrawer({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
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
                      Text(
                        AppLocalizations.of(context)!.text_37,
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(width: 12),
                      const CircleAvatar(
                        radius: 24,
                        backgroundImage: AssetImage('assets/user.png'),
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
                    begin: AlignmentDirectional.centerStart,
                    end: AlignmentDirectional.centerEnd,
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
                    Flexible(
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Flexible(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.end,
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Text(
                                  AppLocalizations.of(context)!.text_379,
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 18,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                                const SizedBox(height: 4),
                                Text(
                                  AppLocalizations.of(context)!.text_24,
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                  style: const TextStyle(
                                    color: Colors.white70,
                                    fontSize: 12,
                                  ),
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(width: 12),
                          const Icon(
                            Icons.account_balance_wallet,
                            color: Colors.white,
                            size: 32,
                          ),
                        ],
                      ),
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
                    title: AppLocalizations.of(context)!.text_1,
                    icon: Icons.home,
                  ),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_32,
                    icon: Icons.local_shipping_outlined,
                  ),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_19,
                    icon: Icons.person_outline,
                  ),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_27,
                    icon: Icons.gavel,
                  ),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_2,
                    icon: Icons.list_alt,
                  ),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_380,
                    icon: Icons.chat_bubble_outline,
                  ),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_284,
                    icon: Icons.privacy_tip_outlined,
                  ),
                  const Divider(height: 32, thickness: 1, indent: 20, endIndent: 20),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_381,
                    icon: Icons.person_outline,
                  ),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_382,
                    icon: Icons.language,
                  ),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_28,
                    icon: Icons.favorite_border,
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
                           Flexible(
                             child: ElevatedButton.icon(
                               onPressed: () => AppModals.showRatingModal(context),
                               style: ElevatedButton.styleFrom(
                                 backgroundColor: Colors.red,
                                 foregroundColor: Colors.white,
                                 shape: RoundedRectangleBorder(
                                   borderRadius: BorderRadius.circular(20),
                                 ),
                                 padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                                 minimumSize: Size.zero,
                               ),
                               icon: const Icon(Icons.star_border, size: 18),
                               label: Text(
                                 AppLocalizations.of(context)!.text_383,
                                 style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12),
                                 overflow: TextOverflow.ellipsis,
                               ),
                             ),
                           ),
                           const SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              AppLocalizations.of(context)!.text_384,
                              textAlign: TextAlign.end,
                              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 11),
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),

                  const SizedBox(height: 16),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_385,
                    icon: Icons.article_outlined,
                    isCompact: true,
                  ),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_386,
                    icon: Icons.help_outline,
                    isCompact: true,
                  ),
                  _buildMenuItem(
                    context,
                    title: AppLocalizations.of(context)!.text_387,
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
                           Flexible(
                             child: ElevatedButton.icon(
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
                               label: Text(
                                 AppLocalizations.of(context)!.text_388,
                                 style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold),
                                 overflow: TextOverflow.ellipsis,
                               ),
                             ),
                           ),
                           const SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              AppLocalizations.of(context)!.text_389,
                              textAlign: TextAlign.right,
                              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12),
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
              padding: const EdgeInsetsDirectional.only(bottom: 20.0),
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
    
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 20.0, vertical: isCompact ? 4.0 : 8.0),
      child: InkWell(
        onTap: () {
          Navigator.pop(context); // Close drawer first
          if (title == AppLocalizations.of(context)!.text_1) {
            Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const HomePage()));
          } else if (title == AppLocalizations.of(context)!.text_32) {
            Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const ServicesPage()));
          } else if (title == AppLocalizations.of(context)!.text_19 || title == AppLocalizations.of(context)!.text_381) {
            Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const AccountPage()));
          } else if (title == AppLocalizations.of(context)!.text_27) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const MyAuctionsPage()));
          } else if (title == AppLocalizations.of(context)!.text_2) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const AllAuctionsPage()));
          } else if (title == AppLocalizations.of(context)!.text_380) {
            AppModals.showContactModal(context);
          } else if (title == AppLocalizations.of(context)!.text_284) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const PrivacyPolicyPage()));
          } else if (title == AppLocalizations.of(context)!.text_28) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const FavoritesPage()));
          } else if (title == AppLocalizations.of(context)!.text_386) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const HowToBidPage()));
          } else if (title == AppLocalizations.of(context)!.text_387 || title == AppLocalizations.of(context)!.text_390) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const AboutMazadPayPage()));
          } else if (title == AppLocalizations.of(context)!.text_382) {
            AppModals.showLanguageModal(context);
          }
        },
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 8.0),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              Text(
                title,
                style: TextStyle(
                  fontSize: isCompact ? 14 : 16,
                  fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                  color: isSelected ? primaryBlue : (isDarkMode ? Colors.white : Colors.black87),
                ),
              ),
              const SizedBox(width: 16),
              Icon(
                icon,
                color: isSelected ? primaryBlue : Colors.grey[600],
                size: isCompact ? 20 : 24,
              ),
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
        color: isYellow ? color : color.withOpacity(0.1),
        shape: BoxShape.circle,
        border: Border.all(color: color.withOpacity(0.2), width: 1),
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