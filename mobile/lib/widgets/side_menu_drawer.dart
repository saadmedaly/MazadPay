import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:share_plus/share_plus.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:mezadpay/pages/home_page.dart';
import 'package:mezadpay/pages/account_page.dart';
import 'package:mezadpay/pages/services_shell_page.dart';
import 'package:mezadpay/pages/about_mazad_pay_page.dart';
import 'package:mezadpay/pages/support_page.dart';
import 'package:mezadpay/pages/my_auctions_shell_page.dart';
import 'package:mezadpay/pages/favorites_page.dart';
import 'package:mezadpay/pages/all_auctions_page.dart';
import 'package:mezadpay/pages/privacy_policy_page.dart';
import 'package:mezadpay/pages/terms_page.dart';
import 'package:mezadpay/widgets/app_modals.dart';
import 'package:mezadpay/services/user_api.dart';
import 'package:mezadpay/pages/account_shell_page.dart';
import 'package:mezadpay/pages/account_profile_page.dart';
import 'package:mezadpay/pages/requests_page.dart';
import 'package:mezadpay/pages/settings_page.dart';

/// Note #3 (client feedback): the drawer's social icons/share button. Only
/// Facebook and TikTok have official client-supplied URLs; Snapchat and
/// Instagram have no official MazadPay URL anywhere in this repository
/// (mobile/backend/web all checked) -- left null rather than guessed, per
/// explicit instruction. SocialPlatform.snapchat/.instagram's tap handler
/// safely no-ops (nothing to open) until the client supplies real links.
enum SocialPlatform { facebook, tiktok, snapchat, instagram }

/// Official MazadPay social links, exactly as supplied by the client for
/// Facebook/TikTok. Kept as a single source of truth so the URL used by the
/// footer social icon matches whatever else in the app might reference it.
const Map<SocialPlatform, String> socialPlatformUrls = {
  SocialPlatform.facebook:
      'https://web.facebook.com/mazadpay?__cft__[0]=AZhR63A_FCJiSqEMNQwseVLTbJS_yXsNyEHbRMf1zTttKph3kTH8qGbXVkCtzluqN8vhXAJDlN0ai1HUx2lmNiP9OFCD1maDfZR5eDJR9YancWSU7PxM7Pq-NDIdSFZlkrRyNg2dGYBIQddWdK4Y2_XAnhbgo431Y7qeTjeHXK6iwCPdkXIxu6RfH3d82to&__tn__=%2Cd%3C%2CP-R',
  SocialPlatform.tiktok: 'https://www.tiktok.com/@mazad.pay?_r=1&_t=ZS-99tq1r65FCP',
  // SocialPlatform.snapchat / .instagram intentionally absent: no official
  // URL exists yet (SNAPCHAT_URL_MISSING / INSTAGRAM_URL_MISSING).
};

/// Pure lookup: the launchable Uri for a social platform, or null if no
/// official URL is configured yet. Both the app/OS's own App Links (for
/// facebook.com/tiktok.com when the native app is installed) and a plain
/// web fallback are handled by the SAME https URL -- launchUrl with
/// LaunchMode.externalApplication lets the OS route it to the installed
/// native app when one has registered as the link's handler, falling back
/// to the browser otherwise. No separate app-scheme URI is needed.
Uri? socialPlatformUri(SocialPlatform platform) {
  final url = socialPlatformUrls[platform];
  if (url == null || url.isEmpty) return null;
  return Uri.tryParse(url);
}

class SideMenuDrawer extends StatefulWidget {
  const SideMenuDrawer({super.key});

  @override
  State<SideMenuDrawer> createState() => _SideMenuDrawerState();
}

class _SideMenuDrawerState extends State<SideMenuDrawer> {
  final UserApi _userApi = UserApi();
  Map<String, dynamic>? _userData;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadUserProfile();
  }

  Future<void> _loadUserProfile() async {
    try {
      final response = await _userApi.getProfile();
      if (!mounted) return;
      
      if (response.success && response.data != null) {
        setState(() {
          _userData = response.data!['user'] ?? response.data!;
          _isLoading = false;
        });
      } else {
        setState(() => _isLoading = false);
      }
    } catch (e) {
      if (!mounted) return;
      setState(() => _isLoading = false);
    }
  }

  String get _userFullName {
    if (_userData == null) return '';
    final name = _userData!['full_name']?.toString() ??
                 _userData!['fullname']?.toString() ??
                 _userData!['name']?.toString() ??
                 '';
    if (name.isNotEmpty) return name;
    // Fallback to masked phone if available
    final phone = _userData!['phone']?.toString() ?? '';
    return phone.isNotEmpty ? phone : '';
  }

  String? get _userAvatarUrl {
    if (_userData == null) return null;
    final avatar = _userData!['avatar_url']?.toString() ??
                   _userData!['avatar']?.toString() ??
                   _userData!['image_url']?.toString() ??
                   _userData!['image']?.toString();
    return avatar?.isNotEmpty == true ? avatar : null;
  }

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
                      if (_isLoading)
                        const SizedBox(
                          width: 24,
                          height: 24,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      else if (_userFullName.isNotEmpty)
                        Container(
                          constraints: const BoxConstraints(maxWidth: 150),
                          child: Text(
                            _userFullName,
                            style: const TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                            ),
                            overflow: TextOverflow.ellipsis,
                            maxLines: 1,
                          ),
                        )
                      else
                        Text(
                          AppLocalizations.of(context)!.text_37,
                          style: const TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      const SizedBox(width: 12),
                      CircleAvatar(
                        radius: 24,
                        backgroundImage: _userAvatarUrl != null
                            ? NetworkImage(_userAvatarUrl!)
                            : const AssetImage('assets/defualtprofile.png') as ImageProvider,
                        onBackgroundImageError: _userAvatarUrl != null
                            ? (_, __) {}
                            : null,
                        child: _userAvatarUrl == null && _isLoading
                            ? const CircularProgressIndicator(strokeWidth: 2)
                            : null,
                      ),
                    ],
                  ),
                ],
              ),
            ),

            // Deposit Banner
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20.0),
              child: InkWell(
                onTap: () {
                  Navigator.pop(context); // Close drawer
                  Navigator.push(
                    context,
                    MaterialPageRoute(builder: (context) => AccountShellPage()),
                  );
                },
                borderRadius: BorderRadius.circular(16),
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
                    title: AppLocalizations.of(context)!.text_401,
                    icon: Icons.request_quote_outlined,
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
                               onPressed: () => _shareApp(context),
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
                  _buildSocialIcon(
                    FontAwesomeIcons.facebookF,
                    const Color(0xFF1877F2),
                    platform: SocialPlatform.facebook,
                  ),
                  const SizedBox(width: 20),
                  _buildSocialIcon(
                    FontAwesomeIcons.instagram,
                    const Color(0xFFE4405F),
                    platform: SocialPlatform.instagram,
                  ),
                  const SizedBox(width: 20),
                  _buildSocialIcon(
                    FontAwesomeIcons.tiktok,
                    const Color(0xFF000000),
                    platform: SocialPlatform.tiktok,
                  ),
                  const SizedBox(width: 20),
                  _buildSocialIcon(
                    FontAwesomeIcons.snapchat,
                    const Color(0xFFFFFC00),
                    isYellow: true,
                    platform: SocialPlatform.snapchat,
                  ),
                ],
              ),
            ),
            const Text(
              'Version 1.0.0',
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
            Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => ServicesShellPage()));
          } else if (title == AppLocalizations.of(context)!.text_19) {
            Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => AccountShellPage()));
          } else if (title == AppLocalizations.of(context)!.text_381) {
            // MAZADPAY -- "المعلومات الشخصية" navigation bug: this item was
            // bundled into the same branch as "حسابي" (text_19), landing on
            // AccountShellPage's account-tab menu (wallet balance, favorites,
            // etc.) instead of the actual personal-information page the
            // client expects. AccountProfilePage ("معلومات الحساب", text_35)
            // is that existing page -- avatar/name/phone/email/city/save +
            // password/settings section -- already reachable one extra tap
            // deep from AccountPage (account_page.dart's text_30 row). Routed
            // directly here instead, via push (not pushReplacement) so its
            // own back arrow (Navigator.pop) returns correctly to wherever
            // the drawer was opened from, matching every other push-based
            // drawer destination below.
            Navigator.push(context, MaterialPageRoute(builder: (context) => const AccountProfilePage()));
          } else if (title == AppLocalizations.of(context)!.text_23) {
            Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => AccountShellPage()));
          } else if (title == AppLocalizations.of(context)!.text_27) {
            Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => MyAuctionsShellPage()));
          } else if (title == AppLocalizations.of(context)!.text_2) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const AllAuctionsPage()));
          } else if (title == AppLocalizations.of(context)!.text_380) {
            AppModals.showContactModal(context);
          } else if (title == AppLocalizations.of(context)!.text_401) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const RequestsPage()));
          } else if (title == AppLocalizations.of(context)!.text_44) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const SettingsPage()));
          } else if (title == AppLocalizations.of(context)!.text_284) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const PrivacyPolicyPage()));
          } else if (title == AppLocalizations.of(context)!.text_385) {
            // MAZADPAY -- "الشروط والأحكام" drawer item had no matching
            // onTap branch at all, so tapping it silently did nothing. The
            // existing TermsPage (already used elsewhere, e.g. the
            // registration flow via create_profile_page.dart) is reused
            // here rather than duplicated -- pushed (not pushReplacement)
            // so its own back arrow/Navigator.pop returns correctly to
            // wherever the drawer was opened from, matching every other
            // push-based drawer destination.
            Navigator.push(context, MaterialPageRoute(builder: (context) => const TermsPage()));
          } else if (title == AppLocalizations.of(context)!.text_28) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const FavoritesPage()));
          } else if (title == AppLocalizations.of(context)!.text_386) {
            Navigator.push(context, MaterialPageRoute(builder: (context) => const SupportPage()));
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
              Expanded(
                child: Text(
                  title,
                  textAlign: TextAlign.end,
                  style: TextStyle(
                    fontSize: isCompact ? 14 : 16,
                    fontWeight: isSelected ? FontWeight.bold : FontWeight.w600,
                    color: isSelected ? primaryBlue : (isDarkMode ? Colors.white : Colors.black87),
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
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

  Widget _buildSocialIcon(
    FaIconData icon,
    Color color, {
    bool isYellow = false,
    required SocialPlatform platform,
  }) {
    return InkWell(
      onTap: () => _openSocialLink(context, platform),
      customBorder: const CircleBorder(),
      child: Container(
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
      ),
    );
  }

  // Note #3: native OS share sheet for MazadPay itself (not a specific
  // auction/win -- see auction_winner_page.dart's _shareWin for that
  // distinct flow). Reuses the exact same share_plus call shape (message +
  // sharePositionOrigin from the tapped widget's RenderBox, required on
  // iPad for the share sheet's popover anchor) already proven there. The
  // website URL matches support_page.dart's own websiteUrl constant (that
  // one is declared on its private _SupportPageState, not importable here).
  void _shareApp(BuildContext context) {
    final message = AppLocalizations.of(context)!.text_389;
    final box = context.findRenderObject() as RenderBox?;
    Share.share(
      '$message\nhttps://mazadpay.com/',
      sharePositionOrigin: box != null ? box.localToGlobal(Offset.zero) & box.size : null,
    );
  }

  // Note #3: safe external-URI launch for a footer social icon, mirroring
  // the same canLaunchUrl/launchUrl(externalApplication) + failure-snackbar
  // pattern already used by support_page.dart's _launchSafely and
  // home_page.dart's _openBannerTargetUrl. A platform with no configured
  // URL yet (Snapchat/Instagram, pending client-supplied links) safely
  // no-ops instead of showing a broken/misleading error.
  Future<void> _openSocialLink(BuildContext context, SocialPlatform platform) async {
    final uri = socialPlatformUri(platform);
    if (uri == null) return;
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
        return;
      }
    } catch (_) {
      // Fall through to the failure snackbar below.
    }
    if (!context.mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(AppLocalizations.of(context)!.text_409)),
    );
  }
}