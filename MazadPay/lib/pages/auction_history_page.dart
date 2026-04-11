import 'package:flutter_gen/gen_l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import '../providers/auction_provider.dart';
import '../models/auction.dart';

class AuctionHistoryPage extends ConsumerWidget {
  final String auctionId;
  const AuctionHistoryPage({super.key, required this.auctionId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final auction = ref.watch(auctionNotifierProvider(auctionId));
    final history = ref.watch(auctionHistoryProvider(auctionId));
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          title: Text(
            AppLocalizations.of(context)!.text_75,
            style: GoogleFonts.plusJakartaSans(
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
              _buildProductHeader(context, auction, isDarkMode),
              const SizedBox(height: 32),
              _buildSummaryHeader(isDarkMode),
              const SizedBox(height: 24),
              _buildBidList(context, history, isDarkMode),
            ],
          ),
        ),
      );
  }

  Widget _buildProductHeader(BuildContext context, Auction auction, bool isDarkMode) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 10, offset: const Offset(0, 4)),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(auction.title, style: GoogleFonts.plusJakartaSans(fontSize: 14, fontWeight: FontWeight.bold)),
                const SizedBox(height: 8),
                Text('${auction.currentPrice.toStringAsFixed(0)} أوقية جديدة', 
                    style: GoogleFonts.plusJakartaSans(fontSize: 18, fontWeight: FontWeight.bold, color: const Color(0xFF0081FF))),
                const SizedBox(height: 4),
                Text(AppLocalizations.of(context)!.text_76, style: GoogleFonts.plusJakartaSans(fontSize: 12, color: Colors.grey)),
              ],
            ),
          ),
          const SizedBox(width: 16),
          ClipRRect(
            borderRadius: BorderRadius.circular(12),
            child: Image.asset(auction.imageUrls[0], width: 100, height: 100, fit: BoxFit.cover),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryHeader(bool isDarkMode) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(AppLocalizations.of(context)!.text_77, style: GoogleFonts.plusJakartaSans(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.grey[400])),
        const SizedBox(height: 16),
        _buildSummaryRow(AppLocalizations.of(context)!.text_78, '15', isDarkMode),
        const SizedBox(height: 12),
        _buildSummaryRow(AppLocalizations.of(context)!.text_79, '12', isDarkMode),
      ],
    );
  }

  Widget _buildSummaryRow(String label, String value, bool isDarkMode) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.grey.withOpacity(0.1)),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.w600)),
          Text(value, style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.bold)),
        ],
      ),
    );
  }

  Widget _buildBidList(BuildContext context, List<BidEntry> history, bool isDarkMode) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(AppLocalizations.of(context)!.text_80, style: GoogleFonts.plusJakartaSans(fontSize: 16, fontWeight: FontWeight.bold, color: Colors.grey[400])),
        const SizedBox(height: 16),
        ListView.separated(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: history.length,
          separatorBuilder: (context, index) => const Divider(height: 32),
          itemBuilder: (context, index) {
            final bid = history[index];
            return Row(
              children: [
                Text('${_getTimeAgo(bid.timestamp)}', style: GoogleFonts.plusJakartaSans(color: Colors.grey[600], fontSize: 13, fontWeight: FontWeight.bold)),
                const Spacer(),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Row(
                      children: [
                        if (index == 0) const Icon(Icons.workspace_premium, color: Colors.orange, size: 20),
                        const SizedBox(width: 4),
                        Text('${bid.amount.toStringAsFixed(0)} أوقية', style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.bold, fontSize: 15)),
                      ],
                    ),
                    const SizedBox(height: 4),
                    Text('${bid.bidderName} (${bid.phoneNumber})', 
                        style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 12)),
                  ],
                ),
              ],
            );
          },
        ),
      ],
    );
  }

  String _getTimeAgo(DateTime dt) {
    final diff = DateTime.now().difference(dt);
    if (diff.inHours > 0) return '${diff.inHours}h : ${diff.inMinutes % 60}m';
    return '${diff.inMinutes}m : ${diff.inSeconds % 60}s';
  }
}
