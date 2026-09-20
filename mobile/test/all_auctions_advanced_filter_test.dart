import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/all_auctions_page.dart';

// Note #1 (client feedback): the Active Auctions screen's advanced filter
// sheet (price range + sort: newest/price_asc/price_desc/ending_soon,
// Apply, Reset) had two real bugs found by source inspection:
//
// 1. Backend: GET /auctions never read or applied min_price/max_price/
//    sort_by at all (confirmed via AuctionHandler.List and
//    auction_repo.go's FindAll, which had a hardcoded
//    "ORDER BY is_featured DESC, created_at DESC" and no price WHERE
//    clause) -- fixed in backend/internal/handlers/auction_handler.go and
//    backend/internal/repository/auction_repo.go, with its own regression
//    tests in backend/internal/services/integrationtest/
//    auction_advanced_filter_test.go.
//
// 2. Mobile: the filter sheet mutated the PARENT State's own
//    _priceRange/_minPrice/_maxPrice/_sortBy fields directly through the
//    modal's setModalState -- so touching a slider or sort option changed
//    those fields immediately, even before Apply, and Reset only reset the
//    sheet's own visual state without ever reloading the list. Fixed by
//    having the sheet use local draft copies, committed to the real parent
//    fields (and triggering a reload) only on Apply or Reset.
//
// These tests cover the extracted pure logic (resolveAdvancedFilterPriceParams,
// auctionMatchesPriceRange, sortAuctionsBy) that the fixed widget code now
// calls -- AllAuctionsPage itself has no DI seam for its live AuctionApi
// (same constraint already documented for other pages in this app), so the
// widget's own Apply/Reset button handlers are not pumped directly here.
void main() {
  group('resolveAdvancedFilterPriceParams (price range -> query params)', () {
    test('the untouched default range (0, 1000000) sends no filter at all', () {
      final params = resolveAdvancedFilterPriceParams(const RangeValues(0, 1000000));
      expect(params.minPrice, isNull);
      expect(params.maxPrice, isNull);
    });

    test('a narrowed range sends both bounds', () {
      final params = resolveAdvancedFilterPriceParams(const RangeValues(100, 5000));
      expect(params.minPrice, 100);
      expect(params.maxPrice, 5000);
    });

    test('only the min moved: max stays unsent (still the default upper bound)', () {
      final params = resolveAdvancedFilterPriceParams(const RangeValues(250, 1000000));
      expect(params.minPrice, 250);
      expect(params.maxPrice, isNull);
    });

    test('only the max moved: min stays unsent (still 0)', () {
      final params = resolveAdvancedFilterPriceParams(const RangeValues(0, 8000));
      expect(params.minPrice, isNull);
      expect(params.maxPrice, 8000);
    });
  });

  group('auctionMatchesPriceRange (client-side belt-and-suspenders price check)', () {
    test('a price within [min, max] matches', () {
      expect(auctionMatchesPriceRange(500, minPrice: 100, maxPrice: 1000), isTrue);
    });

    test('a price below min does not match', () {
      expect(auctionMatchesPriceRange(50, minPrice: 100, maxPrice: 1000), isFalse);
    });

    test('a price above max does not match', () {
      expect(auctionMatchesPriceRange(5000, minPrice: 100, maxPrice: 1000), isFalse);
    });

    test('null min/max means no bound on that side', () {
      expect(auctionMatchesPriceRange(1, minPrice: null, maxPrice: null), isTrue);
      expect(auctionMatchesPriceRange(999999, minPrice: null, maxPrice: null), isTrue);
    });

    test('a price exactly at the boundary matches (inclusive)', () {
      expect(auctionMatchesPriceRange(100, minPrice: 100, maxPrice: 1000), isTrue);
      expect(auctionMatchesPriceRange(1000, minPrice: 100, maxPrice: 1000), isTrue);
    });
  });

  group('sortAuctionsBy (each sort mode)', () {
    final auctions = [
      {
        'id': 'a1',
        'current_price': 300,
        'end_time': DateTime(2026, 1, 10),
        'created_at': DateTime(2026, 1, 1),
      },
      {
        'id': 'a2',
        'current_price': 100,
        'end_time': DateTime(2026, 1, 3),
        'created_at': DateTime(2026, 1, 5),
      },
      {
        'id': 'a3',
        'current_price': 700,
        'end_time': DateTime(2026, 1, 6),
        'created_at': DateTime(2026, 1, 9),
      },
    ];

    test('price_asc orders lowest current_price first', () {
      final sorted = sortAuctionsBy(auctions, 'price_asc');
      expect(sorted.map((a) => a['id']), ['a2', 'a1', 'a3']);
    });

    test('price_desc orders highest current_price first', () {
      final sorted = sortAuctionsBy(auctions, 'price_desc');
      expect(sorted.map((a) => a['id']), ['a3', 'a1', 'a2']);
    });

    test('ending_soon orders nearest end_time first', () {
      final sorted = sortAuctionsBy(auctions, 'ending_soon');
      expect(sorted.map((a) => a['id']), ['a2', 'a3', 'a1']);
    });

    test('newest orders most recently created first', () {
      final sorted = sortAuctionsBy(auctions, 'newest');
      expect(sorted.map((a) => a['id']), ['a3', 'a2', 'a1']);
    });

    test('an unrecognized sort value falls back to newest, matching the backend default', () {
      final sorted = sortAuctionsBy(auctions, 'not_a_real_sort_mode');
      expect(sorted.map((a) => a['id']), ['a3', 'a2', 'a1']);
    });

    test('sortAuctionsBy never mutates the original list (Apply must be able to re-derive from source data)', () {
      final original = List<Map<String, dynamic>>.from(auctions);
      sortAuctionsBy(auctions, 'price_asc');
      expect(auctions, original);
    });
  });

  group('combined filter + sort', () {
    final auctions = [
      {'id': 'cheap', 'current_price': 50, 'end_time': DateTime(2026, 2, 1), 'created_at': DateTime(2026, 1, 1)},
      {'id': 'mid', 'current_price': 500, 'end_time': DateTime(2026, 2, 2), 'created_at': DateTime(2026, 1, 2)},
      {'id': 'high', 'current_price': 5000, 'end_time': DateTime(2026, 2, 3), 'created_at': DateTime(2026, 1, 3)},
    ];

    test('a price filter narrows the set before sort is applied, matching what Apply sends together', () {
      final params = resolveAdvancedFilterPriceParams(const RangeValues(100, 1000));
      final filtered = auctions
          .where((a) => auctionMatchesPriceRange(
                a['current_price'] as num,
                minPrice: params.minPrice,
                maxPrice: params.maxPrice,
              ))
          .toList();
      final sorted = sortAuctionsBy(filtered, 'price_desc');

      expect(sorted.map((a) => a['id']), ['mid']);
    });

    test('a filter that excludes everything combined with any sort mode yields an empty (no-result) list', () {
      final params = resolveAdvancedFilterPriceParams(const RangeValues(100000, 200000));
      final filtered = auctions
          .where((a) => auctionMatchesPriceRange(
                a['current_price'] as num,
                minPrice: params.minPrice,
                maxPrice: params.maxPrice,
              ))
          .toList();
      final sorted = sortAuctionsBy(filtered, 'ending_soon');

      expect(sorted, isEmpty);
    });
  });

  group('Reset restores the exact default values Apply would send for an untouched filter', () {
    test('the default RangeValues(0, 1000000) resolves to no price filter at all', () {
      const defaultRange = RangeValues(0, 1000000);
      final params = resolveAdvancedFilterPriceParams(defaultRange);
      expect(params.minPrice, isNull);
      expect(params.maxPrice, isNull);
    });

    test('"newest" sort mode is the same default sortAuctionsBy uses for an unrecognized value', () {
      final auctions = [
        {'id': 'a', 'current_price': 1, 'end_time': DateTime(2026), 'created_at': DateTime(2026, 1, 1)},
        {'id': 'b', 'current_price': 1, 'end_time': DateTime(2026), 'created_at': DateTime(2026, 1, 2)},
      ];
      expect(sortAuctionsBy(auctions, 'newest'), sortAuctionsBy(auctions, ''));
    });
  });
}
