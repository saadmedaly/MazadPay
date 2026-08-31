import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/auction_provider_api.dart';
import '../services/bid_api.dart';
import '../utils/money_formatter.dart';

class BidActionSheet extends ConsumerStatefulWidget {
  final String auctionId;
  final double currentPrice;
  final double minIncrement;
  final int bidCount;
  final String timeLeft;
  /// ISO-4217 currency code of this specific auction (migration 000046,
  /// Phase 2). Always the auction's own currency, never the viewer's
  /// assumed currency -- cross-market auctions are already inaccessible
  /// per Phase 1, but display must still reflect the auction's currency.
  final String? currencyCode;
  const BidActionSheet({
    super.key,
    required this.auctionId,
    required this.currentPrice,
    required this.minIncrement,
    required this.bidCount,
    required this.timeLeft,
    this.currencyCode,
  });

  @override
  ConsumerState<BidActionSheet> createState() => _BidActionSheetState();
}

class _BidActionSheetState extends ConsumerState<BidActionSheet> {
  int _step = 1; // 1: Increase amount, 2: Final confirm
  double _bidAmount = 0.0;
  bool _isLoading = false;
  final BidApi _bidApi = BidApi();

  @override
  void initState() {
    super.initState();
    _bidAmount = widget.currentPrice + widget.minIncrement;
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Container(
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: const BorderRadius.vertical(top: Radius.circular(32)),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey[300], borderRadius: BorderRadius.circular(2))),
          const SizedBox(height: 24),
          _step == 1 ? _buildStep1(isDarkMode) : _buildStep2(isDarkMode),
          const SizedBox(height: 32),
          _buildActionButton(isDarkMode),
          const SizedBox(height: 16),
        ],
      ),
    );
  }

  Widget _buildStep1(bool isDarkMode) {
    return Column(
      children: [
        Text(AppLocalizations.of(context)!.text_368, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 18, fontWeight: FontWeight.bold)),
        const SizedBox(height: 24),
        
        // Header info box
        Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: isDarkMode ? Colors.black26 : const Color(0xFFF9FAFB),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: Colors.grey.withOpacity(0.1)),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
               _buildInfoColumn('${widget.bidCount}', AppLocalizations.of(context)!.text_369, Colors.red),
               _buildVerticalDivider(),
               _buildInfoColumn(MoneyFormatter.format(widget.currentPrice, widget.currencyCode), AppLocalizations.of(context)!.text_370, Colors.black),
               _buildVerticalDivider(),
               _buildInfoColumn(widget.timeLeft, AppLocalizations.of(context)!.text_371, Colors.red),
            ],
          ),
        ),
        
        const SizedBox(height: 32),
        
        // Amount selector
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
          decoration: BoxDecoration(
            color: isDarkMode ? Colors.black26 : const Color(0xFFF2F4F7),
            borderRadius: BorderRadius.circular(16),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _buildIconButton(Icons.add, () => setState(() => _bidAmount += widget.minIncrement)),
              Expanded(
                child: Center(
                  child: Text(
                    MoneyFormatter.format(_bidAmount, widget.currencyCode),
                    style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 24, fontWeight: FontWeight.bold),
                  ),
                ),
              ),
              _buildIconButton(Icons.remove, () {
                if (_bidAmount > widget.currentPrice + widget.minIncrement) {
                  setState(() => _bidAmount -= widget.minIncrement);
                }
              }),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildStep2(bool isDarkMode) {
    return Column(
      children: [
        Text(AppLocalizations.of(context)!.text_368, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 18, fontWeight: FontWeight.bold)),
        const SizedBox(height: 32),
        Text(MoneyFormatter.format(_bidAmount, widget.currencyCode),
            style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 28, fontWeight: FontWeight.bold, color: const Color(0xFF0081FF))),
        const SizedBox(height: 16),
        Text(AppLocalizations.of(context)!.text_372, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 16, color: Colors.grey)),
        const SizedBox(height: 32),
      ],
    );
  }

  Widget _buildActionButton(bool isDarkMode) {
    return SizedBox(
      width: double.infinity,
      height: 60,
      child: ElevatedButton(
        onPressed: _isLoading ? null : () async {
          if (_step == 1) {
            setState(() => _step = 2);
          } else {
            // Confirm Bid using Optimistic UI
            setState(() => _isLoading = true);
            debugPrint('Placing optimistic bid for auction ID: ${widget.auctionId}, amount: $_bidAmount');
            
            try {
              // Appeler la méthode optimiste du provider
              await ref.read(auctionNotifierApiProvider(widget.auctionId).notifier)
                  .placeBidOptimistically(_bidAmount);
              
              if (mounted) {
                Navigator.of(context).pop();
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(AppLocalizations.of(context)!.text_373),
                    backgroundColor: const Color(0xFF00C58D),
                  ),
                );
              }
            } catch (e) {
              if (mounted) {
                setState(() => _isLoading = false);
                
                // Safe, localized, non-leaking messages for the actual Phase 1
                // error codes (backend/internal/handlers/response.go MapError) --
                // never surface the raw internal error string to the user.
                final raw = e.toString();
                String errorMessage;
                final locale = Localizations.localeOf(context).languageCode;
                if (raw.contains('insufficient_funds') || raw.contains('insufficient_balance')) {
                  errorMessage = locale == 'ar'
                      ? 'رصيدك غير كافٍ لإتمام هذه المزايدة'
                      : (locale == 'fr' ? 'Solde insuffisant pour placer cette enchère' : 'Insufficient balance to place this bid');
                } else if (raw.contains('self_bid') || raw.contains('cannot_bid_own_auction')) {
                  errorMessage = locale == 'ar'
                      ? 'لا يمكنك المزايدة على مزادك الخاص'
                      : (locale == 'fr' ? 'Vous ne pouvez pas enchérir sur votre propre enchère' : 'You cannot bid on your own auction');
                } else if (raw.contains('cross_market_bid_not_allowed')) {
                  errorMessage = locale == 'ar'
                      ? 'لا يمكن المزايدة على مزادات من سوق دولة أخرى'
                      : (locale == 'fr' ? "Impossible d'enchérir sur une enchère d'un autre marché" : 'You cannot bid on an auction from another market');
                } else if (raw.contains('wallet_currency_mismatch')) {
                  errorMessage = locale == 'ar'
                      ? 'تعذر إتمام العملية بسبب عدم تطابق العملة'
                      : (locale == 'fr' ? 'Opération impossible : devise incompatible' : 'Unable to complete: currency mismatch');
                } else if (raw.contains('bid_too_low')) {
                  errorMessage = locale == 'ar'
                      ? 'مبلغ المزايدة منخفض جدًا'
                      : (locale == 'fr' ? "Le montant de l'enchère est trop bas" : 'Bid amount is too low');
                } else if (raw.contains('auction_ended') || raw.contains('auction_not_active')) {
                  errorMessage = locale == 'ar'
                      ? 'هذا المزاد لم يعد نشطًا'
                      : (locale == 'fr' ? "Cette enchère n'est plus active" : 'This auction is no longer active');
                } else {
                  errorMessage = locale == 'ar'
                      ? 'تعذر إتمام المزايدة، حاول مرة أخرى'
                      : (locale == 'fr' ? "Impossible de placer l'enchère, réessayez" : 'Could not place bid, please try again');
                }

                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(errorMessage),
                    backgroundColor: Colors.red,
                    duration: const Duration(seconds: 3),
                  ),
                );
              }
            }
          }
        },
        style: ElevatedButton.styleFrom(
          backgroundColor: const Color(0xFF0081FF),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        ),
        child: _isLoading
            ? const CircularProgressIndicator(color: Colors.white)
            : Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.gavel, color: Colors.white),
                  const SizedBox(width: 8),
                  Text(_step == 1 ? AppLocalizations.of(context)!.text_72 : AppLocalizations.of(context)!.text_374, 
                      style: TextStyle(fontFamily: 'Plus Jakarta Sans', color: Colors.white, fontWeight: FontWeight.bold)),
                ],
              ),
      ),
    );
  }

  Widget _buildInfoColumn(String value, String label, Color valueColor) {
    return Column(
      children: [
        Text(value, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 18, fontWeight: FontWeight.bold, color: valueColor)),
        Text(label, style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 12, color: Colors.grey)),
      ],
    );
  }

  Widget _buildVerticalDivider() {
    return Container(width: 1, height: 40, color: Colors.grey.withOpacity(0.2));
  }

  Widget _buildIconButton(IconData icon, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: 48,
        height: 48,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          boxShadow: [
            BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 5),
          ],
        ),
        child: Icon(icon, color: const Color(0xFF0081FF)),
      ),
    );
  }
}