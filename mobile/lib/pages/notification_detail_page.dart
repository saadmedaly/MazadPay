import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:intl/intl.dart' hide TextDirection;
import 'package:mezadpay/l10n/app_localizations.dart';

/// Customer Request #22: generic detail screen for a notification that has
/// content but no real entity/target to navigate to (admin broadcast
/// notifications today; any future notification that similarly has no
/// existing specialized navigation case). Deliberately the smallest possible
/// screen -- title, full body, optional image, date/time -- not a redesign
/// of the Notifications list itself.
class NotificationDetailPage extends StatelessWidget {
  final String title;
  final String? body;
  final String? imageUrl;
  final DateTime createdAt;

  const NotificationDetailPage({
    super.key,
    required this.title,
    this.body,
    this.imageUrl,
    required this.createdAt,
  });

  @override
  Widget build(BuildContext context) {
    final isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final locale = Localizations.localeOf(context).languageCode;

    // Client feedback #22, requirement D/E: RTL for Arabic, correct LTR
    // layout for French/English -- Directionality here mirrors the pattern
    // already used across this app's other content pages rather than
    // inventing a new one.
    final textDirection = locale == 'ar' ? TextDirection.rtl : TextDirection.ltr;

    return Directionality(
      textDirection: textDirection,
      child: Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          leading: IconButton(
            icon: Icon(
              locale == 'ar' ? Icons.arrow_forward : Icons.arrow_back,
              color: isDarkMode ? Colors.white : Colors.black,
            ),
            onPressed: () => Navigator.of(context).pop(),
          ),
          title: Text(
            AppLocalizations.of(context)?.text_48 ?? 'Notification',
            style: TextStyle(
              color: isDarkMode ? Colors.white : Colors.black,
              fontWeight: FontWeight.w600,
              fontSize: 18,
            ),
          ),
        ),
        body: SingleChildScrollView(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Requirement A: image_url null/empty -> no image box at all.
              // Requirement C: broken image -> neutral placeholder, never a
              // crash -- mirrors the same CachedNetworkImage errorWidget
              // pattern already used throughout this app (Bug H).
              if (imageUrl != null && imageUrl!.trim().isNotEmpty) ...[
                ClipRRect(
                  borderRadius: BorderRadius.circular(16),
                  child: CachedNetworkImage(
                    imageUrl: imageUrl!,
                    width: double.infinity,
                    height: 220,
                    fit: BoxFit.cover,
                    placeholder: (context, url) => Container(
                      width: double.infinity,
                      height: 220,
                      color: Colors.grey[300],
                      child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),
                    ),
                    errorWidget: (context, url, error) => Container(
                      width: double.infinity,
                      height: 220,
                      color: Colors.grey[300],
                      child: const Icon(Icons.image_not_supported, color: Colors.grey, size: 40),
                    ),
                  ),
                ),
                const SizedBox(height: 20),
              ],
              // Requirement F: plain Text only, never HTML/markup execution.
              Text(
                title,
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: isDarkMode ? Colors.white : Colors.black,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                DateFormat('yyyy-MM-dd HH:mm').format(createdAt.toLocal()),
                style: TextStyle(
                  fontSize: 13,
                  color: isDarkMode ? Colors.grey[400] : Colors.grey[600],
                ),
              ),
              const SizedBox(height: 16),
              if (body != null && body!.trim().isNotEmpty)
                Text(
                  body!,
                  style: TextStyle(
                    fontSize: 16,
                    height: 1.6,
                    color: isDarkMode ? Colors.grey[300] : Colors.grey[800],
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
