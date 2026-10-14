import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/auction_provider_api.dart';
import '../services/auction_api.dart';
import '../services/bid_api.dart';
import '../utils/money_formatter.dart';
import '../pages/deposit_page.dart';

// Arabic message shown when the user is already the current highest bidder
// (Customer #38: no two consecutive bids from the same user -- replacing
// the old permanent "one bid per user ever" wording), shared between the
// proactive UI state (has_bid == true, shown before the user even tries)
// and the reactive 409 bid_already_placed error path below (defense-in-depth
// for stale local state) -- see bidPlacementErrorMessage.
const String kAlreadyBidMessageAr = 'أنت بالفعل صاحب أعلى مزايدة حالياً، يرجى الانتظار حتى يزايد شخص آخر';

/// Pure mapping from PlaceBid's raw error string (backend/internal/handlers/
/// response.go MapError codes, client feedback #19's bid_already_placed
/// included) to a safe, localized, user-facing message -- never surfaces the
/// raw internal error string. Extracted from the widget's catch block so it
/// can be tested directly without pumping a widget tree or mocking the
/// network layer (see test/bid_action_sheet_logic_test.dart).
String bidPlacementErrorMessage(String raw, String locale) {
  // MAZADPAY insufficient-balance bid UX: "insufficient_for_insurance" is
  // the actual code BidService.PlaceBid returns when the wallet balance is
  // below the auction's required insurance amount (see backend's MapError
  // case) -- distinct from the generic "insufficient_balance"/
  // "insufficient_funds" codes, which PlaceBid never actually returns
  // (those belong to deposit/withdrawal flows). Checked before the generic
  // branch below so it isn't shadowed by it.
  if (raw.contains('insufficient_for_insurance')) {
    return locale == 'ar'
        ? 'رصيدك غير كافٍ للمزايدة. يرجى شحن رصيدك أولًا.'
        : (locale == 'fr' ? "Solde insuffisant pour enchérir. Veuillez d'abord recharger votre solde." : 'Insufficient balance to bid. Please top up your balance first.');
  } else if (raw.contains('insufficient_funds') || raw.contains('insufficient_balance')) {
    return locale == 'ar'
        ? 'رصيدك غير كافٍ لإتمام هذه المزايدة'
        : (locale == 'fr' ? 'Solde insuffisant pour placer cette enchère' : 'Insufficient balance to place this bid');
  } else if (raw.contains('self_bid') || raw.contains('cannot_bid_own_auction')) {
    return locale == 'ar'
        ? 'لا يمكنك المزايدة على مزادك الخاص'
        : (locale == 'fr' ? 'Vous ne pouvez pas enchérir sur votre propre enchère' : 'You cannot bid on your own auction');
  } else if (raw.contains('cross_market_bid_not_allowed')) {
    return locale == 'ar'
        ? 'لا يمكن المزايدة على مزادات من سوق دولة أخرى'
        : (locale == 'fr' ? "Impossible d'enchérir sur une enchère d'un autre marché" : 'You cannot bid on an auction from another market');
  } else if (raw.contains('wallet_currency_mismatch')) {
    return locale == 'ar'
        ? 'تعذر إتمام العملية بسبب عدم تطابق العملة'
        : (locale == 'fr' ? 'Opération impossible : devise incompatible' : 'Unable to complete: currency mismatch');
  } else if (raw.contains('bid_too_low')) {
    return locale == 'ar'
        ? 'مبلغ المزايدة منخفض جدًا'
        : (locale == 'fr' ? "Le montant de l'enchère est trop bas" : 'Bid amount is too low');
  } else if (raw.contains('auction_ended') || raw.contains('auction_not_active')) {
    return locale == 'ar'
        ? 'هذا المزاد لم يعد نشطًا'
        : (locale == 'fr' ? "Cette enchère n'est plus active" : 'This auction is no longer active');
  } else if (raw.contains('bid_already_placed')) {
    // Customer #38: a user cannot place two consecutive bids in a row --
    // they may bid again once someone else has bid in between. Surfaces the
    // backend's 409 with the same friendly message shown proactively when
    // local state was stale (e.g. another device placed this user's bid
    // concurrently).
    return locale == 'ar'
        ? kAlreadyBidMessageAr
        : (locale == 'fr' ? "Vous avez déjà l'enchère la plus élevée, attendez qu'un autre utilisateur enchérisse" : 'You are already the highest bidder, wait for someone else to bid');
  }
  return locale == 'ar'
      ? 'تعذر إتمام المزايدة، حاول مرة أخرى'
      : (locale == 'fr' ? "Impossible de placer l'enchère, réessayez" : 'Could not place bid, please try again');
}

/// MAZADPAY insufficient-balance bid UX: pure predicate, separated from
/// bidPlacementErrorMessage so the widget can decide whether to show the
/// dedicated insufficient-balance dialog (with a "شحن الرصيد" deposit CTA)
/// instead of the plain error snackbar every other bid failure gets.
bool isInsufficientBalanceBidError(String raw) => raw.contains('insufficient_for_insurance');

/// Whether the bid action should be blocked, per Customer #38 (no two
/// CONSECUTIVE bids from the same user -- replacing the old permanent
/// "one bid per user ever" rule from client feedback #19). has_bid now means
/// "is this user the CURRENT last/highest bidder" (server-authoritative, see
/// backend AuctionService.GetBidStatus / auctions.last_bidder_id) -- it
/// returns to false again once someone else outbids this user, re-enabling
/// the button. The function itself is unchanged (still a direct passthrough)
/// since only has_bid's server-side meaning changed, not this mapping.
bool isRepeatBidBlocked(bool hasBid) => hasBid;

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
  final AuctionApi _auctionApi = AuctionApi();
  // hasAlreadyBid (Customer #38): whether this user is the CURRENT
  // last/highest bidder -- checked proactively via the existing
  // GET /auctions/:id/bid-status endpoint so the bid control can be
  // disabled before the user even tries, and automatically re-enabled once
  // someone else outbids them (realtime/next status check flips has_bid
  // back to false). The backend's 409 bid_already_placed rejection (handled
  // below) remains authoritative regardless -- this is purely a UX head
  // start, never the real guard.
  bool _hasAlreadyBid = false;
  bool _checkingBidStatus = true;

  @override
  void initState() {
    super.initState();
    _bidAmount = widget.currentPrice + widget.minIncrement;
    _checkBidStatus();
  }

  Future<void> _checkBidStatus() async {
    final response = await _auctionApi.getBidStatus(widget.auctionId);
    if (!mounted) return;
    setState(() {
      _hasAlreadyBid = response.success && response.data?['has_bid'] == true;
      _checkingBidStatus = false;
    });
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
        if (_hasAlreadyBid) ...[
          const SizedBox(height: 12),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: Colors.orange.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: Colors.orange.withValues(alpha: 0.3)),
            ),
            child: Row(
              children: [
                const Icon(Icons.info_outline, color: Colors.orange, size: 20),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    kAlreadyBidMessageAr,
                    style: TextStyle(fontFamily: 'Plus Jakarta Sans', fontSize: 13, color: Colors.orange[800]),
                  ),
                ),
              ],
            ),
          ),
        ],
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
        onPressed: (_isLoading || _checkingBidStatus || isRepeatBidBlocked(_hasAlreadyBid)) ? null : () async {
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
                final locale = Localizations.localeOf(context).languageCode;
                if (raw.contains('bid_already_placed')) {
                  setState(() => _hasAlreadyBid = true);
                }

                // MAZADPAY insufficient-balance bid UX: only the genuine
                // insufficient_for_insurance case gets the dedicated
                // dialog + deposit CTA -- every other bid failure (auction
                // ended, bid too low, wallet disabled, owner cannot bid,
                // etc.) keeps its existing plain snackbar behavior
                // unchanged, via bidPlacementErrorMessage.
                if (isInsufficientBalanceBidError(raw)) {
                  showDialog(
                    context: context,
                    builder: (dialogContext) => _InsufficientBalanceDialog(
                      onDeposit: () {
                        Navigator.of(dialogContext).pop();
                        Navigator.of(context).push(
                          MaterialPageRoute(builder: (_) => const DepositPage()),
                        );
                      },
                    ),
                  );
                } else {
                  final errorMessage = bidPlacementErrorMessage(raw, locale);
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

/// MAZADPAY insufficient-balance bid UX: shown only for the genuine
/// insufficient_for_insurance case (see isInsufficientBalanceBidError),
/// styled to match the app's existing SuccessDialog conventions (rounded
/// Dialog, transparent barrier background, circular icon badge) but with
/// two actions instead of one -- "شحن الرصيد" pushes the existing
/// DepositPage (no auction-specific params, matching the plain top-up
/// entry points already used elsewhere: account_page.dart, notifications_page.dart),
/// "إلغاء" just dismisses. No bid is auto-resubmitted after depositing --
/// the user returns to the auction page and can tap bid again themselves.
class _InsufficientBalanceDialog extends StatelessWidget {
  final VoidCallback onDeposit;

  const _InsufficientBalanceDialog({required this.onDeposit});

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      elevation: 0,
      backgroundColor: Colors.transparent,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 80,
              height: 80,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: Colors.red.withValues(alpha: 0.1),
              ),
              child: const Center(
                child: Icon(Icons.account_balance_wallet_outlined, color: Colors.red, size: 40),
              ),
            ),
            const SizedBox(height: 24),
            Text(
              AppLocalizations.of(context)!.text_423,
              style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.black),
              textAlign: TextAlign.center,
              textDirection: TextDirection.rtl,
            ),
            const SizedBox(height: 8),
            Text(
              AppLocalizations.of(context)!.text_424,
              style: TextStyle(fontSize: 14, color: Colors.grey[700]),
              textAlign: TextAlign.center,
              textDirection: TextDirection.rtl,
            ),
            const SizedBox(height: 24),
            SizedBox(
              width: double.infinity,
              height: 48,
              child: ElevatedButton(
                onPressed: onDeposit,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF0081FF),
                  elevation: 0,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                ),
                child: Text(
                  AppLocalizations.of(context)!.text_425,
                  style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            const SizedBox(height: 8),
            SizedBox(
              width: double.infinity,
              height: 48,
              child: TextButton(
                onPressed: () => Navigator.of(context).pop(),
                child: Text(
                  AppLocalizations.of(context)!.text_426,
                  style: const TextStyle(color: Colors.grey, fontSize: 16, fontWeight: FontWeight.w600),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}