import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/widgets/bid_action_sheet.dart';

// Client feedback #19: a user may successfully bid on a given auction at
// most once, ever -- even after being outbid. These tests cover the pure
// decision logic extracted from BidActionSheet (bidPlacementErrorMessage,
// isRepeatBidBlocked) directly, without pumping a widget tree or mocking
// the network/Riverpod layer -- the smallest practical seam for behavior
// that is otherwise entirely UI-glue around a real HTTP call.
void main() {
  group('isRepeatBidBlocked (client feedback #19)', () {
    test('has_bid = false -> the normal bid action remains available', () {
      expect(isRepeatBidBlocked(false), isFalse);
    });

    test('has_bid = true -> the repeat-bid action is blocked', () {
      expect(isRepeatBidBlocked(true), isTrue);
    });

    // Existing "highest bidder" state (is_user_highest_bidder /
    // is_highest_bid) is a DIFFERENT concept from has_bid: an outbid user
    // is no longer highest but must still be blocked from re-bidding.
    // isRepeatBidBlocked deliberately takes only hasBid as input -- there is
    // no code path anywhere that could substitute a "highest bidder" flag
    // for it, which this test documents explicitly.
    test('blocking is keyed only on has_bid, never on highest-bidder status', () {
      // A user who is NOT the current highest bidder but HAS bid before
      // must still be blocked -- has_bid alone determines this.
      const hasBidButNotHighest = true;
      expect(isRepeatBidBlocked(hasBidButNotHighest), isTrue);
    });
  });

  group('bidPlacementErrorMessage (client feedback #19: bid_already_placed)', () {
    test('bid_already_placed (Arabic) maps to the exact friendly message', () {
      final message = bidPlacementErrorMessage('bid_already_placed', 'ar');
      expect(message, kAlreadyBidMessageAr);
      expect(message, 'لقد قمت بالمزايدة على هذا المزاد مسبقًا');
    });

    test('bid_already_placed (French) maps to a distinct friendly message', () {
      final message = bidPlacementErrorMessage('bid_already_placed', 'fr');
      expect(message, isNot(kAlreadyBidMessageAr));
      expect(message, contains('déjà enchéri'));
    });

    test('bid_already_placed (English) maps to a distinct friendly message', () {
      final message = bidPlacementErrorMessage('bid_already_placed', 'en');
      expect(message, isNot(kAlreadyBidMessageAr));
      expect(message, contains('already bid'));
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
