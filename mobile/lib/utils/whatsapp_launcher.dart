import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

/// MazadPay's own official WhatsApp number (matches support_page.dart's
/// canonical whatsappNumber / my_winnings_page.dart's
/// kMazadPayWhatsAppNumber -- the same value already confirmed there to
/// agree with login_page.dart and app_modals.dart). Not duplicated as a
/// separate constant per page; this is the single source new callers
/// (Note #4: services_page.dart) should use.
const String kMazadPayWhatsAppNumber = '47601175';

/// wa.me link for MazadPay's own WhatsApp with a prefilled message.
/// Mirrors support_page.dart's SupportContactUris.whatsApp / the primary
/// URI my_winnings_page.dart already builds for the exact same number.
Uri buildMazadPayWhatsAppUri(String message) {
  return Uri.parse('https://wa.me/222$kMazadPayWhatsAppNumber')
      .replace(queryParameters: {'text': message});
}

/// REAL DEVICE BUG fallback (Customer #30, my_winnings_page.dart): on
/// Android 11+ (targetSdk 30+), canLaunchUrl()/launchUrl() for the
/// https://wa.me/... link can fail even with WhatsApp installed unless the
/// host app declares package-visibility `<queries>` for that exact intent.
/// This native `whatsapp://send` deep link is the second, independent path
/// to the same conversation/prefilled message, tried only if the primary
/// wa.me launch fails.
Uri buildMazadPayWhatsAppNativeUri(String message) {
  return Uri.parse('whatsapp://send').replace(queryParameters: {
    'phone': '222$kMazadPayWhatsAppNumber',
    'text': message,
  });
}

/// Note #4 (client feedback): the SAME reliable WhatsApp-launch strategy
/// already fixed in My Winnings (my_winnings_page.dart's
/// _openPaymentWhatsApp) -- safe wa.me primary launch, whatsapp:// native
/// fallback on failure, a snackbar only if both fail, never a crash. Shared
/// here rather than reimplemented per the ticket's explicit "do not create
/// another launch implementation unnecessarily" instruction.
Future<void> launchMazadPayWhatsApp(BuildContext context, String message) async {
  final primaryUri = buildMazadPayWhatsAppUri(message);
  final nativeUri = buildMazadPayWhatsAppNativeUri(message);
  try {
    final launched = await launchUrl(primaryUri, mode: LaunchMode.externalApplication);
    if (launched) return;
    final fallbackLaunched = await launchUrl(nativeUri, mode: LaunchMode.externalApplication);
    if (fallbackLaunched) return;
    if (!context.mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('تعذّر فتح واتساب. تأكد من تثبيته على جهازك.')),
    );
  } catch (_) {
    try {
      final fallbackLaunched = await launchUrl(nativeUri, mode: LaunchMode.externalApplication);
      if (fallbackLaunched) return;
    } catch (_) {
      // fall through to the error snackbar below
    }
    if (!context.mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('تعذّر فتح واتساب. تأكد من تثبيته على جهازك.')),
    );
  }
}
