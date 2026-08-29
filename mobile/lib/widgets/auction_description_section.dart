import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The "الوصف" (description) card shown on the auction details page.
///
/// Extracted from `AuctionDetailsPage._buildDescriptionSection` so it can be
/// unit-tested in isolation from the rest of that page (which is coupled to
/// network calls, Riverpod providers, and timers). Renders nothing when the
/// description is empty/blank, matching the previous inline behavior.
class AuctionDescriptionSection extends StatelessWidget {
  final String description;
  final bool isDarkMode;

  const AuctionDescriptionSection({
    super.key,
    required this.description,
    required this.isDarkMode,
  });

  @override
  Widget build(BuildContext context) {
    if (description.trim().isEmpty) return const SizedBox.shrink();

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.03),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            AppLocalizations.of(context)!.label_description,
            style: const TextStyle(
              fontFamily: 'Plus Jakarta Sans',
              fontSize: 16,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 12),
          Text(
            description,
            textAlign: TextAlign.right,
            style: TextStyle(
              fontFamily: 'Plus Jakarta Sans',
              fontSize: 14,
              height: 1.6,
              color: isDarkMode ? Colors.white70 : Colors.black87,
            ),
          ),
        ],
      ),
    );
  }
}
