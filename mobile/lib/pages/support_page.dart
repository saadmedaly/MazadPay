import 'dart:async';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:mezadpay/services/faq_api.dart';
import 'package:mezadpay/services/realtime_sync_service.dart';
import 'package:url_launcher/url_launcher.dart';

/// Customer Request #25: pure helpers for the contact URIs this page opens,
/// extracted so they can be unit-tested without a live url_launcher call
/// (mirrors the existing wa.me pattern already used in
/// auction_details_page.dart's _openWhatsApp).
class SupportContactUris {
  /// Mauritania international WhatsApp deep link. Reuses the exact same
  /// "222" + local-number concatenation already established in
  /// auction_details_page.dart's _supportPhone/_openWhatsApp -- no new
  /// international-format convention invented here.
  static Uri whatsApp(String localNumber) => Uri.parse('https://wa.me/222$localNumber');

  // Strips display-only spaces (the visible value keeps them for
  // readability, e.g. "+222 36 60 11 75") -- the tel: URI itself must be a
  // clean number.
  static Uri phoneCall(String rawNumber) => Uri.parse('tel:${rawNumber.replaceAll(' ', '')}');

  static Uri email(String address) => Uri.parse('mailto:$address');

  static Uri website(String url) => Uri.parse(url);
}

class SupportPage extends StatefulWidget {
  const SupportPage({super.key});

  @override
  State<SupportPage> createState() => _SupportPageState();
}

class _SupportPageState extends State<SupportPage> {
  // Customer Request #25 (client QA item #18): the current canonical contact
  // values -- confirmed to already match login_page.dart's "forgot PIN"
  // contact modal and app_modals.dart's contact sheet (both independently
  // use '47601175'/'mazadpay@gmail.com'), unlike this page's own prior
  // stale values ('+222 36 60 11 75' / 'support@mazadpay.mr'), which nothing
  // else in the app agreed with. The phone-call number is deliberately kept
  // as its own distinct pre-existing value per explicit product decision --
  // NOT unified with the WhatsApp number.
  static const String whatsappNumber = '47601175';
  static const String phoneCallNumber = '+222 36 60 11 75';
  static const String emailAddress = 'mazadpay@gmail.com';
  static const String websiteUrl = 'https://mazadpay.com/';

  final FaqApi _faqApi = FaqApi();
  List<dynamic> _faqs = [];
  bool _isLoadingFaqs = true;
  StreamSubscription? _realtimeEventSub;
  StreamSubscription? _realtimeCatchUpSub;

  @override
  void initState() {
    super.initState();
    _loadFaqs();
    // Customer #20: support_page.dart previously had no refresh mechanism at
    // all (no RefreshIndicator, no realtime) -- an admin FAQ edit was only
    // visible after leaving and re-entering this page or restarting the
    // app. Refetches FAQ only (Phase 7: targeted invalidation), never the
    // whole page/app.
    _realtimeEventSub = RealtimeSyncService().events.listen((event) {
      if (!mounted) return;
      if (event.type == RealtimeEventType.faqUpdated) {
        _loadFaqs();
      }
    });
    _realtimeCatchUpSub = RealtimeSyncService().catchUpSignal.listen((_) {
      if (!mounted) return;
      _loadFaqs();
    });
  }

  @override
  void dispose() {
    _realtimeEventSub?.cancel();
    _realtimeCatchUpSub?.cancel();
    super.dispose();
  }

  // Customer #25: safe external-URI launch shared by WhatsApp/phone/email/
  // website taps -- canLaunchUrl/launchUrl guarded exactly like the existing
  // auction_details_page.dart._openWhatsApp pattern (external app mode, a
  // snackbar on failure, never a crash/unhandled exception).
  Future<void> _launchSafely(Uri uri, String failureMessage) async {
    try {
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } else {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(failureMessage)));
      }
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(failureMessage)));
    }
  }

  // Customer #20 hardening (out-of-order response guard): _loadFaqs can now
  // be triggered concurrently by a realtime faq.updated event and app-resume
  // catch-up.
  int _faqsRequestGeneration = 0;

  Future<void> _loadFaqs() async {
    final myGeneration = ++_faqsRequestGeneration;
    final response = await _faqApi.getFaqs();
    if (mounted && myGeneration == _faqsRequestGeneration) {
      setState(() {
        if (response.success && response.data != null) {
          _faqs = response.data!;
        }
        _isLoadingFaqs = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final isArabic = Localizations.localeOf(context).languageCode == 'ar';

    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          title: Text(
            AppLocalizations.of(context)!.text_309,
            style: TextStyle(fontFamily: 'Plus Jakarta Sans', 
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
              Text(
                AppLocalizations.of(context)!.text_310,
                style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 24, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              Text(
                AppLocalizations.of(context)!.text_311,
                style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.grey[600]),
              ),
              const SizedBox(height: 32),
              _buildContactTile(
                icon: Icons.chat_bubble_outline,
                title: AppLocalizations.of(context)!.text_312,
                subtitle: whatsappNumber,
                subtitleIsLtrValue: true,
                color: const Color(0xFF25D366),
                isDarkMode: isDarkMode,
                onTap: () => _launchSafely(
                  SupportContactUris.whatsApp(whatsappNumber),
                  AppLocalizations.of(context)!.text_409,
                ),
              ),
              const SizedBox(height: 16),
              _buildContactTile(
                icon: Icons.phone_in_talk_outlined,
                title: AppLocalizations.of(context)!.text_314,
                subtitle: phoneCallNumber,
                subtitleIsLtrValue: true,
                color: const Color(0xFF0081FF),
                isDarkMode: isDarkMode,
                onTap: () => _launchSafely(
                  SupportContactUris.phoneCall(phoneCallNumber),
                  AppLocalizations.of(context)!.text_409,
                ),
              ),
              const SizedBox(height: 16),
              _buildContactTile(
                icon: Icons.alternate_email,
                title: AppLocalizations.of(context)!.text_41,
                subtitle: emailAddress,
                subtitleIsLtrValue: true,
                color: Colors.orange,
                isDarkMode: isDarkMode,
                onTap: () => _launchSafely(
                  SupportContactUris.email(emailAddress),
                  AppLocalizations.of(context)!.text_409,
                ),
              ),
              const SizedBox(height: 16),
              _buildContactTile(
                icon: Icons.language,
                title: AppLocalizations.of(context)!.text_408,
                subtitle: websiteUrl,
                subtitleIsLtrValue: true,
                color: const Color(0xFF6C5CE7),
                isDarkMode: isDarkMode,
                onTap: () => _launchSafely(
                  SupportContactUris.website(websiteUrl),
                  AppLocalizations.of(context)!.text_409,
                ),
              ),
              const SizedBox(height: 40),
              Text(
                AppLocalizations.of(context)!.text_207,
                style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 16),
              if (_isLoadingFaqs)
                const Center(child: CircularProgressIndicator())
              else if (_faqs.isEmpty)
                Text(
                  "No FAQs available",
                  style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.grey),
                )
              else
                ..._faqs.map((faq) {
                  final question = isArabic 
                      ? (faq['question_ar'] ?? '') 
                      : (faq['question_fr'] ?? faq['question_ar'] ?? '');
                  final answer = isArabic 
                      ? (faq['answer_ar'] ?? '') 
                      : (faq['answer_fr'] ?? faq['answer_ar'] ?? '');
                  
                  return _buildFaqItem(
                    context, 
                    question, 
                    answer, 
                    isDarkMode
                  );
                }),
            ],
          ),
        ),
      );
  }

  // Customer #25: subtitleIsLtrValue marks a subtitle that is a contact
  // VALUE (phone number, email, URL) rather than descriptive text --
  // wrapped in its own Directionality(TextDirection.ltr) so digits/URLs
  // render in the correct logical left-to-right order even while the
  // surrounding page stays RTL for Arabic (only this one Text is isolated,
  // never the whole page/tile). Rendered in solid black (bold, per the
  // client's explicit "black text" requirement) instead of the previous
  // grey descriptive-text style, which only ever applied to non-value
  // subtitles (e.g. "Fast direct response") and is preserved for onTap==null
  // tiles if any are ever added again.
  Widget _buildContactTile({
    required IconData icon,
    required String title,
    required String subtitle,
    required Color color,
    required bool isDarkMode,
    bool subtitleIsLtrValue = false,
    VoidCallback? onTap,
  }) {
    final subtitleText = Text(
      subtitle,
      style: TextStyle(
        fontFamily: 'Plus Jakarta Sans',
        color: subtitleIsLtrValue ? (isDarkMode ? Colors.white : Colors.black) : Colors.grey,
        fontWeight: subtitleIsLtrValue ? FontWeight.w600 : FontWeight.normal,
        fontSize: subtitleIsLtrValue ? 14 : 12,
      ),
    );

    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.withOpacity(0.1)),
        boxShadow: [
          BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 10, offset: const Offset(0, 4)),
        ],
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(color: color.withOpacity(0.1), shape: BoxShape.circle),
            child: Icon(icon, color: color, size: 24),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontWeight: FontWeight.bold, fontSize: 16)),
                const SizedBox(height: 4),
                subtitleIsLtrValue
                    ? Directionality(textDirection: TextDirection.ltr, child: subtitleText)
                    : subtitleText,
              ],
            ),
          ),
          const Icon(Icons.arrow_forward_ios, size: 16, color: Colors.grey),
        ],
      ),
      ),
    );
  }

  Widget _buildFaqItem(BuildContext context, String question, String answer, bool isDarkMode) {
    return Container(
      margin: const EdgeInsetsDirectional.only(bottom: 12),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(12),
      ),
      child: ExpansionTile(
        title: Text(question, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 14, fontWeight: FontWeight.w600)),
        trailing: const Icon(Icons.keyboard_arrow_down, size: 20, color: Color(0xFF0081FF)),
        childrenPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        expandedAlignment: Alignment.centerLeft,
        children: [
          Text(
            answer,
            style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 13, color: Colors.grey[600], height: 1.5),
          ),
        ],
      ),
    );
  }
}

