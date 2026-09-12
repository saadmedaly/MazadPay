import 'package:flutter_test/flutter_test.dart';

// Customer feedback #4 (select 100/500 MRU before payment). DepositPage
// previously had a single hardcoded _depositAmount = 500.0 with no way to
// know it was showing the right amount for an auction subscription request.
// The fee is never user-selectable -- it comes from the auction request's
// server-stamped subscription_fee (derived from the category's fee_tier,
// backend/internal/models/auction.go Category.SubscriptionFee()). This
// reproduces the pure decision logic (DepositPage's _depositAmount getter
// and the navigation decision in create_ad_form_page.dart) as widget-free
// predicates.

double resolveDepositAmount({double? subscriptionFee, double fallback = 500.0}) {
  return subscriptionFee ?? fallback;
}

bool isAuctionSubscriptionFee({String? auctionRequestId, double? subscriptionFee}) {
  return auctionRequestId != null && subscriptionFee != null;
}

// Mirrors DepositPage._submit()'s deposit-request body construction
// (financial-integrity round): auction_request_id is included ONLY when the
// page was reached from the auction-subscription flow. The backend then
// derives the authoritative amount from that id and ignores 'amount' -- but
// mobile must still never OMIT it for that flow, or the deposit would
// silently fall back to trusting the client-supplied amount.
Map<String, dynamic> buildDepositRequestBody({
  required double amount,
  required String gateway,
  String? auctionRequestId,
}) {
  return {
    'amount': amount,
    'gateway': gateway,
    'payment_method': gateway,
    if (auctionRequestId != null) 'auction_request_id': auctionRequestId,
  };
}

// Mirrors create_ad_form_page.dart's post-submit decision: only navigate to
// the payment page when the request was actually submitted for review
// (status == 'pending') and the server returned both an id and a fee.
bool shouldNavigateToDepositAfterSubmit({
  required String status,
  String? requestId,
  double? subscriptionFee,
}) {
  return status == 'pending' && requestId != null && subscriptionFee != null;
}

void main() {
  group('DepositPage amount resolution', () {
    test('an auction request subscription fee (100) is used directly', () {
      expect(resolveDepositAmount(subscriptionFee: 100.0), 100.0);
    });

    test('an auction request subscription fee (500, premium category) is used directly', () {
      expect(resolveDepositAmount(subscriptionFee: 500.0), 500.0);
    });

    test('no subscription fee (generic wallet top-up entry point) falls back to the legacy constant', () {
      expect(resolveDepositAmount(subscriptionFee: null), 500.0);
    });
  });

  group('isAuctionSubscriptionFee (controls whether the fee card/tagged note appear)', () {
    test('both id and fee present -> true', () {
      expect(isAuctionSubscriptionFee(auctionRequestId: 'r1', subscriptionFee: 100.0), isTrue);
    });

    test('neither present (generic wallet deposit) -> false', () {
      expect(isAuctionSubscriptionFee(auctionRequestId: null, subscriptionFee: null), isFalse);
    });

    test('fee missing even with an id present -> false (never show a fee that was not authoritatively provided)', () {
      expect(isAuctionSubscriptionFee(auctionRequestId: 'r1', subscriptionFee: null), isFalse);
    });
  });

  group('Navigation to DepositPage after submitting an auction request', () {
    test('submitting for review (pending) with a real id+fee navigates to payment', () {
      expect(
        shouldNavigateToDepositAfterSubmit(status: 'pending', requestId: 'r1', subscriptionFee: 100.0),
        isTrue,
      );
    });

    test('saving a draft never navigates to payment (not under review yet)', () {
      expect(
        shouldNavigateToDepositAfterSubmit(status: 'draft', requestId: 'r1', subscriptionFee: 100.0),
        isFalse,
      );
    });

    test('a pending submission missing the fee in the response does not navigate (fails safe, no guessed amount)', () {
      expect(
        shouldNavigateToDepositAfterSubmit(status: 'pending', requestId: 'r1', subscriptionFee: null),
        isFalse,
      );
    });

    test('a pending submission missing the id does not navigate', () {
      expect(
        shouldNavigateToDepositAfterSubmit(status: 'pending', requestId: null, subscriptionFee: 100.0),
        isFalse,
      );
    });
  });

  group('Deposit request body (financial-integrity round)', () {
    test('the auction-subscription flow sends auction_request_id', () {
      final body = buildDepositRequestBody(amount: 500.0, gateway: 'bankily', auctionRequestId: 'req-123');
      expect(body['auction_request_id'], 'req-123');
      expect(body.containsKey('auction_request_id'), isTrue);
    });

    test('the generic wallet deposit never sends auction_request_id', () {
      final body = buildDepositRequestBody(amount: 250.0, gateway: 'bankily');
      expect(body.containsKey('auction_request_id'), isFalse);
    });
  });
}
