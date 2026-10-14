import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/widgets/bid_action_sheet.dart';

// Customer #38: a user cannot place two CONSECUTIVE bids on the same
// auction -- they may bid again once someone else has bid in between
// (replacing the old permanent "one bid per user ever" rule from client
// feedback #19). These tests cover the pure decision logic extracted from
// BidActionSheet (bidPlacementErrorMessage, isRepeatBidBlocked) directly,
// without pumping a widget tree or mocking the network/Riverpod layer -- the
// smallest practical seam for behavior that is otherwise entirely UI-glue
// around a real HTTP call.
void main() {
  group('isRepeatBidBlocked (Customer #38: has_bid = "currently last bidder")', () {
    test('has_bid = false -> the bid action is available', () {
      expect(isRepeatBidBlocked(false), isFalse);
    });

    test('has_bid = true (caller is the current last/highest bidder) -> blocked', () {
      expect(isRepeatBidBlocked(true), isTrue);
    });

    // has_bid now means "is the caller the CURRENT last/highest bidder"
    // (server-authoritative, auctions.last_bidder_id), not "has ever bid" --
    // it flips back to false once someone else outbids the caller, which is
    // exactly what re-enables the button without any other mobile code
    // change (see backend AuctionService.GetBidStatus).
    test('blocking is keyed only on has_bid, which the backend now re-derives per request', () {
      const isCurrentlyLastBidder = true;
      expect(isRepeatBidBlocked(isCurrentlyLastBidder), isTrue);
    });
  });

  group('bidPlacementErrorMessage (Customer #38: bid_already_placed)', () {
    test('bid_already_placed (Arabic) maps to the exact friendly message', () {
      final message = bidPlacementErrorMessage('bid_already_placed', 'ar');
      expect(message, kAlreadyBidMessageAr);
      expect(message, 'أنت بالفعل صاحب أعلى مزايدة حالياً، يرجى الانتظار حتى يزايد شخص آخر');
    });

    test('bid_already_placed (French) maps to a distinct friendly message', () {
      final message = bidPlacementErrorMessage('bid_already_placed', 'fr');
      expect(message, isNot(kAlreadyBidMessageAr));
      expect(message, contains("l'enchère la plus élevée"));
    });

    test('bid_already_placed (English) maps to a distinct friendly message', () {
      final message = bidPlacementErrorMessage('bid_already_placed', 'en');
      expect(message, isNot(kAlreadyBidMessageAr));
      expect(message, contains('highest bidder'));
    });

    test('a raw error containing the full backend exception text still matches', () {
      // The widget's catch block passes e.toString(), not the bare code --
      // e.g. "Exception: bid_already_placed" -- so the mapping must use
      // substring matching, not exact equality (mirrors production usage).
      final message = bidPlacementErrorMessage('Exception: bid_already_placed', 'ar');
      expect(message, kAlreadyBidMessageAr);
    });

    test('other existing error codes remain correctly mapped (no regression)', () {
      expect(bidPlacementErrorMessage('bid_too_low', 'ar'), contains('منخفض'));
      expect(bidPlacementErrorMessage('cannot_bid_own_auction', 'ar'), contains('مزادك الخاص'));
      expect(bidPlacementErrorMessage('auction_ended', 'ar'), contains('لم يعد نشطًا'));
      expect(bidPlacementErrorMessage('insufficient_balance', 'ar'), contains('غير كافٍ'));
    });

    test('an unrecognized error falls back to the generic retry message', () {
      final message = bidPlacementErrorMessage('some_never_seen_code', 'ar');
      expect(message, 'تعذر إتمام المزايدة، حاول مرة أخرى');
    });

    test('bid_already_placed takes precedence and is never confused with bid_too_low or auction_ended', () {
      // Sanity check that the new branch doesn't accidentally fall through
      // an existing contains() check due to substring overlap.
      final message = bidPlacementErrorMessage('bid_already_placed', 'ar');
      expect(message, isNot(contains('منخفض'))); // not bid_too_low's message
      expect(message, isNot(contains('نشطًا'))); // not auction_ended's message
    });
  });

  // MAZADPAY insufficient-balance bid UX: BidService.PlaceBid returns the
  // "insufficient_for_insurance" code (backend/internal/errors/errors.go,
  // backend/internal/handlers/response.go MapError) when the wallet balance
  // is below the auction's required insurance amount -- distinct from the
  // generic "insufficient_balance"/"insufficient_funds" codes, which
  // PlaceBid never actually returns (those belong to deposit/withdrawal
  // flows). isInsufficientBalanceBidError is the pure predicate the widget
  // uses to decide whether to show the dedicated deposit-CTA dialog instead
  // of the plain error snackbar every other bid failure gets.
  group('isInsufficientBalanceBidError / insufficient_for_insurance (MAZADPAY insufficient-balance bid UX)', () {
    test('1. balance = 0 -> raw "insufficient_for_insurance" is detected as insufficient-balance', () {
      expect(isInsufficientBalanceBidError('insufficient_for_insurance'), isTrue);
    });

    test('1b. the widget catch-block form (Exception.toString(), code prefixed) still matches', () {
      // providers/auction_provider_api.dart throws Exception('$code: $message'),
      // so e.toString() is "Exception: insufficient_for_insurance: <arabic message>".
      const raw = 'Exception: insufficient_for_insurance: رصيدك غير كافٍ لتغطية مبلغ التأمين';
      expect(isInsufficientBalanceBidError(raw), isTrue);
      expect(bidPlacementErrorMessage(raw, 'ar'), 'رصيدك غير كافٍ للمزايدة. يرجى شحن رصيدك أولًا.');
    });

    test('2. balance less than required amount -> same code, same detection (backend does not distinguish 0 vs partial)', () {
      // The backend check is `wallet.Balance.LessThan(auction.InsuranceAmount)`
      // -- both a zero balance and a smaller-than-required balance hit the
      // exact same branch and return the same code, so mobile-side handling
      // is identical for both; no separate code exists to distinguish them.
      expect(isInsufficientBalanceBidError('insufficient_for_insurance'), isTrue);
    });

    test('bidPlacementErrorMessage returns the exact preferred Arabic/French/English copy', () {
      expect(bidPlacementErrorMessage('insufficient_for_insurance', 'ar'),
          'رصيدك غير كافٍ للمزايدة. يرجى شحن رصيدك أولًا.');
      expect(bidPlacementErrorMessage('insufficient_for_insurance', 'fr'), contains('Solde insuffisant'));
      expect(bidPlacementErrorMessage('insufficient_for_insurance', 'en'), contains('Insufficient balance'));
    });

    test('4. bid_too_low is NOT detected as an insufficient-balance error', () {
      expect(isInsufficientBalanceBidError('bid_too_low'), isFalse);
    });

    test('5. auction_ended is NOT detected as an insufficient-balance error', () {
      expect(isInsufficientBalanceBidError('auction_ended'), isFalse);
      expect(isInsufficientBalanceBidError('auction_not_active'), isFalse);
    });

    test('6. wallet_disabled is NOT detected as insufficient balance (checked earlier in PlaceBid, different failure category)', () {
      expect(isInsufficientBalanceBidError('wallet_disabled'), isFalse);
    });

    test('insurance_not_set (auction misconfigured, not a user-balance issue) is NOT detected as insufficient balance', () {
      expect(isInsufficientBalanceBidError('insurance_not_set'), isFalse);
    });

    test('the generic insufficient_balance/insufficient_funds codes (never actually returned by PlaceBid) are NOT matched by the specific predicate', () {
      // These codes belong to deposit/withdrawal flows; PlaceBid never
      // returns them, but the predicate is still exact-code-scoped so it
      // would never accidentally fire for them if it ever did.
      expect(isInsufficientBalanceBidError('insufficient_balance'), isFalse);
      expect(isInsufficientBalanceBidError('insufficient_funds'), isFalse);
    });

    test('cannot_bid_own_auction, cross_market_bid_not_allowed, wallet_currency_mismatch are NOT detected as insufficient balance', () {
      expect(isInsufficientBalanceBidError('cannot_bid_own_auction'), isFalse);
      expect(isInsufficientBalanceBidError('cross_market_bid_not_allowed'), isFalse);
      expect(isInsufficientBalanceBidError('wallet_currency_mismatch'), isFalse);
    });

    test('8. a generic/unrecognized network or server error is NOT detected as insufficient balance (existing generic behavior preserved)', () {
      expect(isInsufficientBalanceBidError('some_never_seen_code'), isFalse);
      expect(isInsufficientBalanceBidError('Exception: SocketException'), isFalse);
      expect(bidPlacementErrorMessage('Exception: SocketException', 'ar'),
          'تعذر إتمام المزايدة، حاول مرة أخرى');
    });
  });
}
