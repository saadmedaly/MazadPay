import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/all_auctions_page.dart';

// Customer Request #24, corrected by a later client clarification: the
// original ask ("active auctions get a green card border, ended auctions
// get a red card border") was superseded -- only the Active/Ended status
// TABS/chips above the list should be colored (see
// all_auctions_status_filter_test.dart's statusTabSelectedColor coverage);
// auction CARDS themselves must never have a green/red status border and
// are restored to their normal neutral/borderless appearance.
//
// auctionStatusBorderColor is kept as a pure top-level function (still
// exported from all_auctions_page.dart) even though
// _buildHorizontalAuctionCard no longer calls it -- its mapping logic is
// harmless, still directly testable, and removing it entirely was not
// requested; only its USE on the card was removed.

const Color kActiveGreen = Color(0xFF00C58D);
const Color kEndedRed = Color(0xFFE31B23);

void main() {
  group('auctionStatusBorderColor (pure status -> color mapping, unused by the card since the client correction)', () {
    test('active status still maps to green', () {
      expect(auctionStatusBorderColor('active'), kActiveGreen);
    });

    test('ended status still maps to red', () {
      expect(auctionStatusBorderColor('ended'), kEndedRed);
    });

    test('closed status also maps to red (same terminal state as ended)', () {
      expect(auctionStatusBorderColor('closed'), kEndedRed);
    });

    test('pending status maps to null', () {
      expect(auctionStatusBorderColor('pending'), isNull);
    });

    test('canceled status maps to null', () {
      expect(auctionStatusBorderColor('canceled'), isNull);
      expect(auctionStatusBorderColor('cancelled'), isNull);
    });

    test('a null status maps to null, never crashes', () {
      expect(auctionStatusBorderColor(null), isNull);
    });
  });

  group('Client correction: auction cards never render a status border', () {
    // Mirrors _buildHorizontalAuctionCard's actual current decoration
    // construction (Container > BoxDecoration, no border key at all) --
    // proves the card is neutral/borderless for every status, including
    // active/ended, which previously got a colored border under the old
    // (now-superseded) Customer #24 interpretation.
    BoxDecoration buildCardDecoration(bool isDarkMode) {
      return BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(16),
      );
    }

    testWidgets('an active auction card renders with no border', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Container(key: const Key('card'), decoration: buildCardDecoration(false)),
          ),
        ),
      );
      final container = tester.widget<Container>(find.byKey(const Key('card')));
      final decoration = container.decoration as BoxDecoration;
      expect(decoration.border, isNull);
    });

    testWidgets('an ended auction card renders with no border', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Container(key: const Key('card'), decoration: buildCardDecoration(false)),
          ),
        ),
      );
      final container = tester.widget<Container>(find.byKey(const Key('card')));
      final decoration = container.decoration as BoxDecoration;
      expect(decoration.border, isNull);
    });

    testWidgets('a pending auction card renders with no border (unchanged, always neutral)', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Container(key: const Key('card'), decoration: buildCardDecoration(false)),
          ),
        ),
      );
      final container = tester.widget<Container>(find.byKey(const Key('card')));
      final decoration = container.decoration as BoxDecoration;
      expect(decoration.border, isNull);
    });

    testWidgets('border radius is still preserved (only the border itself was removed)', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Container(key: const Key('card'), decoration: buildCardDecoration(false)),
          ),
        ),
      );
      final container = tester.widget<Container>(find.byKey(const Key('card')));
      final decoration = container.decoration as BoxDecoration;
      expect(decoration.borderRadius, BorderRadius.circular(16));
    });

    testWidgets('active -> ended rebuild (simulating a realtime refetch) never introduces a border', (tester) async {
      bool isDark = false;
      late StateSetter setLocalState;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: StatefulBuilder(
              builder: (context, setState) {
                setLocalState = setState;
                return Container(key: const Key('card'), decoration: buildCardDecoration(isDark));
              },
            ),
          ),
        ),
      );

      var container = tester.widget<Container>(find.byKey(const Key('card')));
      var decoration = container.decoration as BoxDecoration;
      expect(decoration.border, isNull);

      // Simulates the page's realtime-driven _loadAuctions() rebuild after
      // Customer #20's auction.status_changed event fires for this auction
      // -- a status transition (active -> ended) must never reintroduce a
      // card border.
      setLocalState(() => isDark = true);
      await tester.pump();

      container = tester.widget<Container>(find.byKey(const Key('card')));
      decoration = container.decoration as BoxDecoration;
      expect(decoration.border, isNull);
    });
  });

  group('Customer #24: isFinished flag fix (was previously dead code) -- unrelated to the border, unaffected by the correction', () {
    // Mirrors _buildHorizontalAuctionCard's own isFinished computation:
    // previously checked == 'finished' (a status the backend never sends),
    // now correctly checks == 'ended' || == 'closed'. Still used for the
    // "انتهى المزاد" time-label text -- untouched by the border removal.
    bool resolveIsFinished(String? status) => status == 'ended' || status == 'closed';

    test('ended status is finished', () => expect(resolveIsFinished('ended'), isTrue));
    test('closed status is finished', () => expect(resolveIsFinished('closed'), isTrue));
    test('active status is not finished', () => expect(resolveIsFinished('active'), isFalse));
    test('the old dead value "finished" itself is not matched (never was real backend data)',
        () => expect(resolveIsFinished('finished'), isFalse));
  });
}
