import 'dart:async';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/auction_provider_api.dart';
import '../models/auction.dart';

class AuctionHistoryPage extends ConsumerStatefulWidget {
  final String auctionId;
  final Auction? auctionData;
  const AuctionHistoryPage({super.key, required this.auctionId, this.auctionData});

  @override
  ConsumerState<AuctionHistoryPage> createState() => _AuctionHistoryPageState();
}

class _AuctionHistoryPageState extends ConsumerState<AuctionHistoryPage> {
  Timer? _refreshTimer;

  @override
  void initState() {
    super.initState();
    // Rafraîchissement automatique toutes les 5 secondes (temps réel)
    _refreshTimer = Timer.periodic(const Duration(seconds: 5), (_) {
      ref.invalidate(auctionHistoryApiProvider(widget.auctionId));
    });
  }

  @override
  void dispose() {
    _refreshTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final historyAsync = ref.watch(auctionHistoryApiProvider(widget.auctionId));
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    // Si on a déjà les données de l'enchère, pas besoin d'appel API supplémentaire
    final Widget headerWidget;
    if (widget.auctionData != null) {
      headerWidget = _buildProductHeader(context, widget.auctionData!, isDarkMode);
    } else {
      final auctionAsync = ref.watch(auctionNotifierApiProvider(widget.auctionId));
      headerWidget = auctionAsync.when(
        data: (auction) => _buildProductHeader(context, auction, isDarkMode),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => Text(AppLocalizations.of(context)!.error_loading_auction),
      );
    }

    return Scaffold(
      backgroundColor: isDarkMode
          ? const Color(0xFF121212)
          : const Color(0xFFFBFBFB),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: true,
        title: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              AppLocalizations.of(context)!.text_75,
              style: TextStyle(
                fontFamily: 'Plus Jakarta Sans',
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: isDarkMode ? Colors.white : Colors.black,
              ),
            ),
            const SizedBox(width: 8),
            // Indicateur temps réel
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: Colors.red,
                borderRadius: BorderRadius.circular(4),
              ),
              child: const Text(
                'LIVE',
                style: TextStyle(
                  color: Colors.white,
                  fontSize: 10,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ],
        ),
        leading: IconButton(
          icon: Icon(
            Icons.arrow_back_ios,
            color: isDarkMode ? Colors.white : Colors.black,
            size: 20,
          ),
          onPressed: () => Navigator.of(context).pop(),
        ),
        actions: [
          IconButton(
            icon: Icon(
              Icons.refresh,
              color: isDarkMode ? Colors.white : Colors.black,
            ),
            onPressed: () {
              ref.invalidate(auctionHistoryApiProvider(widget.auctionId));
            },
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(auctionHistoryApiProvider(widget.auctionId));
          await ref.read(auctionHistoryApiProvider(widget.auctionId).future);
        },
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              headerWidget,
              const SizedBox(height: 32),
              historyAsync.when(
                data: (history) => _buildBidList(context, history, isDarkMode),
                loading: () => const Center(child: CircularProgressIndicator()),
                error: (_, __) => Text(_getLocalizedError(context)),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _getLocalizedError(BuildContext context) {
    final loc = AppLocalizations.of(context)!;
    try {
      return loc.error_loading_auctions;
    } catch (_) {
      return 'Error loading bids';
    }
  }

  String _getNoBidsText(BuildContext context) {
    final locale = Localizations.localeOf(context).languageCode;
    switch (locale) {
      case 'ar':
        return 'لا توجد مزايدات حتى الآن';
      case 'fr':
        return 'Aucune enchère pour le moment';
      case 'en':
      default:
        return 'No bids yet';
    }
  }

  String _getHiddenPhoneText(BuildContext context) {
    final locale = Localizations.localeOf(context).languageCode;
    switch (locale) {
      case 'ar':
        return 'الهاتف مخفي';
      case 'fr':
        return 'Téléphone masqué';
      case 'en':
      default:
        return 'Phone hidden';
    }
  }

  String _getCurrencySymbol(BuildContext context) {
    final locale = Localizations.localeOf(context).languageCode;
    switch (locale) {
      case 'ar':
        return 'أوقية';
      case 'fr':
      case 'en':
      default:
        return 'MRU';
    }
  }

  Widget _buildProductHeader(
    BuildContext context,
    Auction auction,
    bool isDarkMode,
  ) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Lot number badge
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            decoration: BoxDecoration(
              color: const Color(0xFF0084FF),
              borderRadius: BorderRadius.circular(6),
            ),
            child: Text(
              'Lot #${auction.lotNumber}',
              style: const TextStyle(
                fontFamily: 'Plus Jakarta Sans',
                color: Colors.white,
                fontSize: 12,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      auction.title,
                      style: TextStyle(
                        fontFamily: 'Plus Jakarta Sans',
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: isDarkMode ? Colors.white : Colors.black,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 12,
                      runSpacing: 4,
                      children: [
                        Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              Icons.people_outline,
                              size: 16,
                              color: Colors.grey[600],
                            ),
                            const SizedBox(width: 4),
                            Flexible(
                              child: Text(
                                '${auction.bidderCount}',
                                style: TextStyle(
                                  fontFamily: 'Plus Jakarta Sans',
                                  fontSize: 12,
                                  color: Colors.grey[600],
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          ],
                        ),
                        Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              Icons.visibility_outlined,
                              size: 16,
                              color: Colors.grey[600],
                            ),
                            const SizedBox(width: 4),
                            Flexible(
                              child: Text(
                                '${auction.views}',
                                style: TextStyle(
                                  fontFamily: 'Plus Jakarta Sans',
                                  fontSize: 12,
                                  color: Colors.grey[600],
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Text(
                      '${auction.currentPrice.toStringAsFixed(0)} ${_getCurrencySymbol(context)}',
                      style: const TextStyle(
                        fontFamily: 'Plus Jakarta Sans',
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0081FF),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 16),
              ClipRRect(
                borderRadius: BorderRadius.circular(12),
                child: _buildAuctionImage(auction.imageUrls.isNotEmpty ? auction.imageUrls[0] : 'assets/corolla.png'),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildBidList(
    BuildContext context,
    List<BidEntry> history,
    bool isDarkMode,
  ) {
    if (history.isEmpty) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            AppLocalizations.of(context)!.text_80,
            style: TextStyle(
              fontFamily: 'Plus Jakarta Sans',
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: Colors.grey[400],
            ),
          ),
          const SizedBox(height: 24),
          Center(
            child: Column(
              children: [
                Icon(
                  Icons.gavel_outlined,
                  size: 48,
                  color: Colors.grey[400],
                ),
                const SizedBox(height: 12),
                Text(
                  _getNoBidsText(context),
                  style: TextStyle(
                    fontFamily: 'Plus Jakarta Sans',
                    color: Colors.grey[600],
                    fontSize: 14,
                  ),
                ),
              ],
            ),
          ),
        ],
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          '${AppLocalizations.of(context)!.text_80} (${history.length})',
          style: TextStyle(
            fontFamily: 'Plus Jakarta Sans',
            fontSize: 16,
            fontWeight: FontWeight.bold,
            color: Colors.grey[400],
          ),
        ),
        const SizedBox(height: 16),
        ListView.separated(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: history.length,
          separatorBuilder: (context, index) => const Divider(height: 32),
          itemBuilder: (context, index) {
            final bid = history[index];
            final isWinner = index == 0;
            return Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: isWinner
                    ? Colors.orange.withOpacity(0.1)
                    : (isDarkMode ? const Color(0xFF1D1D1D) : Colors.white),
                borderRadius: BorderRadius.circular(12),
                border: isWinner
                    ? Border.all(color: Colors.orange.withOpacity(0.3))
                    : null,
              ),
              child: Row(
                children: [
                  // Rang du bid
                  Container(
                    width: 32,
                    height: 32,
                    decoration: BoxDecoration(
                      color: isWinner ? Colors.orange : Colors.grey[300],
                      shape: BoxShape.circle,
                    ),
                    child: Center(
                      child: Text(
                        '${index + 1}',
                        style: TextStyle(
                          fontFamily: 'Plus Jakarta Sans',
                          color: isWinner ? Colors.white : Colors.grey[700],
                          fontWeight: FontWeight.bold,
                          fontSize: 14,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  // Info utilisateur
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          bid.bidderName,
                          style: TextStyle(
                            fontFamily: 'Plus Jakarta Sans',
                            fontWeight: FontWeight.bold,
                            fontSize: 14,
                            color: isDarkMode ? Colors.white : Colors.black,
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          bid.phoneNumber.isNotEmpty
                              ? bid.phoneNumber
                              : _getHiddenPhoneText(context),
                          style: TextStyle(
                            fontFamily: 'Plus Jakarta Sans',
                            color: Colors.grey[600],
                            fontSize: 12,
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          _getTimeAgo(bid.timestamp),
                          style: TextStyle(
                            fontFamily: 'Plus Jakarta Sans',
                            color: Colors.grey[500],
                            fontSize: 11,
                          ),
                        ),
                      ],
                    ),
                  ),
                  // Montant
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      if (isWinner)
                        const Icon(
                          Icons.workspace_premium,
                          color: Colors.orange,
                          size: 20,
                        ),
                      const SizedBox(height: 4),
                      Text(
                        '${bid.amount.toStringAsFixed(0)} ${_getCurrencySymbol(context)}',
                        style: TextStyle(
                          fontFamily: 'Plus Jakarta Sans',
                          fontWeight: FontWeight.bold,
                          fontSize: 15,
                          color: const Color(0xFF0081FF),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
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

  Widget _buildAuctionImage(String imagePath) {
    final isNetworkImage = imagePath.startsWith('http');
    if (isNetworkImage) {
      return Image.network(
        imagePath,
        width: 100,
        height: 100,
        fit: BoxFit.cover,
        errorBuilder: (c, e, s) => Image.asset(
          'assets/corolla.png',
          width: 100,
          height: 100,
          fit: BoxFit.cover,
        ),
        loadingBuilder: (context, child, loadingProgress) {
          if (loadingProgress == null) return child;
          return Container(
            width: 100,
            height: 100,
            color: Colors.grey[200],
            child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),
          );
        },
      );
    } else {
      return Image.asset(
        imagePath,
        width: 100,
        height: 100,
        fit: BoxFit.cover,
        errorBuilder: (c, e, s) => Container(
          width: 100,
          height: 100,
          color: Colors.grey[300],
          child: const Icon(Icons.image, color: Colors.grey),
        ),
      );
    }
  }
}
