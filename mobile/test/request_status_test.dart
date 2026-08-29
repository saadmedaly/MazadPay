import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/requests_page.dart';

void main() {
  group('classifyRequestStatus', () {
    test('maps approved/active to the approved badge', () {
      expect(classifyRequestStatus('approved', isAuctionTab: true).statusKey, 'approved');
      expect(classifyRequestStatus('active', isAuctionTab: true).statusKey, 'approved');
    });

    test('maps draft/rejected/unknown to their own badges', () {
      expect(classifyRequestStatus('draft', isAuctionTab: true).statusKey, 'draft');
      expect(classifyRequestStatus('rejected', isAuctionTab: true).statusKey, 'rejected');
      expect(classifyRequestStatus('something_else', isAuctionTab: true).statusKey, 'pending');
      expect(classifyRequestStatus(null, isAuctionTab: true).statusKey, 'pending');
    });

    test('is case-insensitive on the raw status', () {
      expect(classifyRequestStatus('DRAFT', isAuctionTab: true).statusKey, 'draft');
      expect(classifyRequestStatus('Rejected', isAuctionTab: true).statusKey, 'rejected');
    });

    test('offers edit-and-resubmit only for auction-tab draft/rejected requests', () {
      expect(classifyRequestStatus('draft', isAuctionTab: true).canEditResubmit, isTrue);
      expect(classifyRequestStatus('rejected', isAuctionTab: true).canEditResubmit, isTrue);
      expect(classifyRequestStatus('approved', isAuctionTab: true).canEditResubmit, isFalse);
      expect(classifyRequestStatus('pending', isAuctionTab: true).canEditResubmit, isFalse);
    });

    test('never offers edit-and-resubmit on the banner tab (no update endpoint)', () {
      expect(classifyRequestStatus('draft', isAuctionTab: false).canEditResubmit, isFalse);
      expect(classifyRequestStatus('rejected', isAuctionTab: false).canEditResubmit, isFalse);
    });
  });
}
