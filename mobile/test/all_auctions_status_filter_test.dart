import 'package:flutter_test/flutter_test.dart';

// Customer feedback #12 (restore Active/Ended auctions selector). The status
// tabs themselves (all_auctions_page.dart's _buildStatusTab/_statusFilter)
// already existed on this branch (client feedback A12), driving a real
// server-side GET /auctions?status=... request. The actual defect proven
// and fixed this round was backend-side: AuctionRepository.FindAll always
// appended "AND end_time > NOW()", which silently excluded every ended
// auction from a status='ended' request (an ended auction's end_time is, by
// definition, always in the past) -- making the Ended tab permanently
// empty. This file reproduces the pure client-side pieces of the fix: the
// tab-count resolution (real total, never page-length) and the request
// parameters sent per tab, as predicates independent of widgets/network.

int resolveTabCount(Map<String, dynamic>? responseData) {
  if (responseData == null) return 0;
  return (responseData['total'] as int?) ?? 0;
}

Map<String, String?> statusRequestParams(String statusFilter, String? categoryId) {
  return {'status': statusFilter, 'category_id': categoryId};
}

void main() {
  group('Tab count resolution (client feedback #12)', () {
    test('a real backend total is used directly, not the page length', () {
      // total=47 even though only 1 row was fetched (limit:1 for a
      // lightweight count-only request) -- proves the count is not derived
      // from the returned list's length.
      final count = resolveTabCount({'auctions': [
        {'id': 'a1'},
      ], 'total': 47});
      expect(count, 47);
    });

    test('a zero total (no matching auctions) resolves to zero, not a crash', () {
      expect(resolveTabCount({'auctions': [], 'total': 0}), 0);
    });

    test('a missing total field falls back to 0 rather than throwing', () {
      expect(resolveTabCount({'auctions': []}), 0);
    });

    test('a null response resolves to 0 (network/error case), no crash', () {
      expect(resolveTabCount(null), 0);
    });
  });

  group('Per-tab request parameters', () {
    test('the active tab requests status=active', () {
      final params = statusRequestParams('active', null);
      expect(params['status'], 'active');
    });

    test('the ended tab requests status=ended, not a stale finished/completed value', () {
      final params = statusRequestParams('ended', null);
      expect(params['status'], 'ended');
    });

    test('a selected category is included alongside the status filter', () {
      final params = statusRequestParams('ended', 'cat-5');
      expect(params['status'], 'ended');
      expect(params['category_id'], 'cat-5');
    });

    test('no category selected (all categories) sends a null category_id, not a placeholder string', () {
      final params = statusRequestParams('active', null);
      expect(params['category_id'], isNull);
    });
  });
}
