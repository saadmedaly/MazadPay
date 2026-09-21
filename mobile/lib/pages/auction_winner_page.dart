import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart' hide TextDirection;
import 'package:share_plus/share_plus.dart';

import '../providers/auction_provider_api.dart';
import '../models/auction.dart';
import '../utils/money_formatter.dart';
import '../utils/auction_image.dart';
import '../utils/whatsapp_launcher.dart';
import 'my_winnings_page.dart' show buildWinnerPaymentMessage;

class AuctionWinnerPage extends ConsumerWidget {
  final String auctionId;
  const AuctionWinnerPage({super.key, required this.auctionId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    // Customer feedback #11 final proof: this page previously read
    // auctionNotifierProvider (auction_provider.dart), whose build(id) IGNORES
    // the id and always returns hardcoded demo data ("Toyota Corolla 2018",
    // mock-seller-id) -- every winner would have seen the same fake auction
    // regardless of what they actually won. auctionNotifierApiProvider
    // (auction_provider_api.dart) is the real, live-fetching provider already
    // used by auction_details_page.dart for the identical id -> Auction
    // lookup (GET auction by id, with its own loading/error/data states).
    final auctionAsync = ref.watch(auctionNotifierApiProvider(auctionId));

    return auctionAsync.when(
      loading: () => Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        body: const Center(child: CircularProgressIndicator()),
      ),
      error: (err, stack) => Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          leading: IconButton(
            icon: const Icon(Icons.close),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ),
        body: Center(
          child: Text(AppLocalizations.of(context)!.error_loading_auction),
        ),
      ),
      data: (auction) => Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        body: Stack(
          children: [
            // Fireworks Background (Simulated with icons)
            _buildFireworksDecoration(context),

            SafeArea(
              child: SingleChildScrollView(
                physics: const BouncingScrollPhysics(),
                padding: const EdgeInsetsDirectional.only(bottom: 24),
                child: Column(
                  children: [
                    _buildHeader(context, auction),
                    const SizedBox(height: 20),
                    // Customer #23: the client's required phrase "مبروك، ربحت
                    // المزاد" is now sourced from AppLocalizations (text_406),
                    // localized for ar/fr/en, instead of a hardcoded
                    // Arabic-only string -- same single-sentence wording
                    // requirement as before (Phase B), just no longer
                    // language-locked.
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 24),
                      child: Text(
                        AppLocalizations.of(context)!.text_406,
                        textAlign: TextAlign.center,
                        style: TextStyle(fontFamily: 'Plus Jakarta Sans',
                          fontSize: 28,
                          fontWeight: FontWeight.w900,
                          color: isDarkMode ? Colors.white : Colors.black87,
                        ),
                      ),
                    ),
                    const SizedBox(height: 40),

                    _buildWinningAmountBox(auction, isDarkMode),
                    const SizedBox(height: 32),

                    _buildProductCard(auction, isDarkMode),
                    const SizedBox(height: 32),

                    _buildWinnerSummary(context, auction, isDarkMode),

                    const SizedBox(height: 40),
                    _buildFooterAction(context, auction, isDarkMode),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFireworksDecoration(BuildContext context) {
     final bool isRtl = Directionality.of(context) == TextDirection.rtl;
     return Stack(
       children: [
         Positioned.directional(textDirection: isRtl ? TextDirection.rtl : TextDirection.ltr, top: 100, start: 50, child: Icon(Icons.star, color: Colors.orange.withOpacity(0.3), size: 40)),
         Positioned.directional(textDirection: isRtl ? TextDirection.rtl : TextDirection.ltr, top: 150, end: 80, child: Icon(Icons.auto_awesome, color: Colors.blue.withOpacity(0.3), size: 50)),
         Positioned.directional(textDirection: isRtl ? TextDirection.rtl : TextDirection.ltr, top: 300, start: 30, child: Icon(Icons.favorite, color: Colors.red.withOpacity(0.2), size: 30)),
         Positioned.directional(textDirection: isRtl ? TextDirection.rtl : TextDirection.ltr, bottom: 200, end: 40, child: Icon(Icons.wb_sunny, color: Colors.yellow.withOpacity(0.3), size: 60)),
       ],
     );
  }

  Widget _buildHeader(BuildContext context, Auction auction) {
    return Padding(
      padding: const EdgeInsets.all(16.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          IconButton(
            icon: const Icon(Icons.share_outlined, color: Color(0xFF135BEC)),
            onPressed: () => _shareWin(context, auction),
          ),
          IconButton(
            icon: const Icon(Icons.close, color: Colors.grey),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ],
      ),
    );
  }

  // Customer #23: winner-share requirement. Only public/safe auction data
  // (title, settled amount, currency) goes into the shared text -- never a
  // winner user id, phone number, payment/wallet data, or any internal DB
  // id, per the implementation brief's explicit safe-content list.
  void _shareWin(BuildContext context, Auction auction) {
    final amount = MoneyFormatter.format(auction.currentPrice, auction.currencyCode);
    // gen-l10n alphabetizes generated positional params (amount, title) --
    // matching that exact generated order here, not the ARB source order.
    final message = AppLocalizations.of(context)!.text_407(amount, auction.title);
    final box = context.findRenderObject() as RenderBox?;
    Share.share(
      message,
      sharePositionOrigin: box != null ? box.localToGlobal(Offset.zero) & box.size : null,
    );
  }

  Widget _buildWinningAmountBox(Auction auction, bool isDarkMode) {
    // Customer #23: this was a hardcoded fake date ('2026/02/15') unrelated
    // to any real auction data, shown to every winner regardless of when
    // they actually won. auction.endTime is the real field (win date == the
    // moment the auction actually ended) -- payment_deadline is a distinct
    // concept (when payment is due) and is deliberately not used here to
    // avoid mislabeling a payment deadline as the win date.
    final winDateLabel = DateFormat('yyyy/MM/dd').format(auction.endTime.toLocal());
    return Column(
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 8),
          decoration: const BoxDecoration(
            color: Color(0xFFE31B23),
            borderRadius: BorderRadius.vertical(top: Radius.circular(12)),
          ),
          child: Text(
            winDateLabel,
            style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.white, fontWeight: FontWeight.bold),
          ),
        ),
        Container(
          width: 300,
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: const Color(0xFFFFCC00),
            borderRadius: BorderRadius.circular(16),
            boxShadow: [
              BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 20, offset: const Offset(0, 10)),
            ],
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
               const Icon(Icons.check_circle, color: Color(0xFF00C58D), size: 32),
               const SizedBox(width: 12),
               Text(
                 MoneyFormatter.format(auction.currentPrice, auction.currencyCode),
                 style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 28, fontWeight: FontWeight.w900, color: Colors.black),
               ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildProductCard(Auction auction, bool isDarkMode) {
     // Customer #23: harmonized onto the shared Bug H helper
     // (resolveAuctionImageUrl) instead of this page's own direct
     // imageUrls[0] indexing + local-asset-guessing branch -- same
     // contract as My Winnings: a real URL renders, otherwise the neutral
     // placeholder below, never a bundled fake product image.
     final imageUrl = resolveAuctionImageUrl(auction.imageUrls);
     return Container(
       padding: const EdgeInsets.all(8),
       decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 10, offset: const Offset(0, 4)),
          ],
       ),
       child: ClipRRect(
         borderRadius: BorderRadius.circular(12),
         child: imageUrl != null
             ? Image.network(
                 imageUrl,
                 width: 300,
                 height: 180,
                 fit: BoxFit.cover,
                 errorBuilder: (c, e, s) => Container(
                   width: 300,
                   height: 180,
                   color: Colors.grey[200],
                   child: const Icon(Icons.image_not_supported, color: Colors.grey),
                 ),
               )
             : Container(
                 width: 300,
                 height: 180,
                 color: Colors.grey[200],
                 child: const Icon(Icons.image_not_supported, color: Colors.grey),
               ),
       ),
     );
  }

  Widget _buildWinnerSummary(BuildContext context, Auction auction, bool isDarkMode) {
    return Container(
      width: 350,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: const Color(0xFF135BEC).withOpacity(0.1)),
      ),
      child: Row(
        children: [
           Container(
             width: 60, height: 60,
             decoration: BoxDecoration(
               color: const Color(0xFFFFF7E6),
               borderRadius: BorderRadius.circular(12),
             ),
             child: const Center(child: Icon(Icons.emoji_events, color: Color(0xFFFFCC00), size: 40)),
           ),
           const SizedBox(width: 16),
           Expanded(
             child: Column(
               crossAxisAlignment: CrossAxisAlignment.start,
               children: [
                  // Phase B final check (client feedback item 11): this used
                  // to show a hardcoded static name ("محمد احمد سيديا") and a
                  // generic "الفائز الاول بالمزاد" label to every winner
                  // regardless of who they actually are or what they won --
                  // fake content, not real backend data. Replaced with the
                  // real auction title (already loaded on this page via
                  // auctionNotifierProvider, the same data already fixed to
                  // carry a real backend-persisted winner_id in Phase B).
                  Text(
                    auction.title.isNotEmpty ? auction.title : AppLocalizations.of(context)!.text_84,
                    style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 18, fontWeight: FontWeight.bold),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  Text(AppLocalizations.of(context)!.text_84, style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.grey)),
                  // Customer #23: lot number, when present, per the client's
                  // expected minimum winner screen fields.
                  if (auction.lotNumber.isNotEmpty)
                    Text(
                      '#${auction.lotNumber}',
                      style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 12, color: Colors.grey[500]),
                    ),
               ],
             ),
           ),
        ],
      ),
    );
  }

  // MAZADPAY -- winner notification payment button bug: this button was a
  // placeholder ("Future payment gateway integration", never implemented),
  // so tapping "أكمل عملية الدفع" here did nothing at all. Fixed to behave
  // EXACTLY like the already-working My Winnings payment button
  // (my_winnings_page.dart's _openPaymentWhatsApp): same WhatsApp number,
  // same prefilled message builder (buildWinnerPaymentMessage, imported
  // from my_winnings_page.dart rather than reimplemented), same wa.me
  // primary launch / whatsapp:// native fallback / safe error handling
  // (launchMazadPayWhatsApp, the shared launcher already used elsewhere --
  // see utils/whatsapp_launcher.dart's own doc comment for why this exists).
  Widget _buildFooterAction(BuildContext context, Auction auction, bool isDarkMode) {
    final message = buildWinnerPaymentMessage(
      auctionTitle: auction.title,
      lotNumber: auction.lotNumber,
      formattedAmount: MoneyFormatter.format(auction.currentPrice, auction.currencyCode),
    );
    return Padding(
      padding: const EdgeInsets.all(24.0),
      child: SizedBox(
        width: double.infinity,
        height: 60,
        child: ElevatedButton(
          onPressed: () => launchMazadPayWhatsApp(context, message),
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFF4A7DFF),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          ),
          child: Text(
            AppLocalizations.of(context)!.text_85,
            style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white),
          ),
        ),
      ),
    );
  }
}
