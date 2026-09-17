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
}
