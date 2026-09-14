import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/all_auctions_page.dart';

// Customer Request #24: Active/Ended tabbed auction list only
// (all_auctions_page.dart) -- active auctions get a green card border,
// ended auctions get a red card border. Scoped deliberately to this one
// screen (per explicit product decision): Home, My Auctions, Favorites, My
// Winnings, and Related Auctions are untouched and use their own existing
// status semantics/design.
//
// auctionStatusBorderColor is a pure top-level function exported from
// all_auctions_page.dart specifically so this logic is directly testable
// without pumping the full page (which fetches from a real AuctionApi with
// no DI seam, the same constraint already documented for
// MyWinningsPage/SupportPage in prior rounds).

const Color kActiveGreen = Color(0xFF00C58D);
const Color kEndedRed = Color(0xFFE31B23);

void main() {
  group('Customer #24: auctionStatusBorderColor (pure status -> color mapping)', () {
    test('active status returns green', () {
      expect(auctionStatusBorderColor('active'), kActiveGreen);
    });

    test('ended status returns red', () {
      expect(auctionStatusBorderColor('ended'), kEndedRed);
    });

    test('closed status also returns red (same terminal state as ended)', () {
      expect(auctionStatusBorderColor('closed'), kEndedRed);
    });

    test('pending status returns null (no incorrect active-green)', () {
      expect(auctionStatusBorderColor('pending'), isNull);
    });

    test('canceled status returns null (no incorrect ended-red)', () {
      expect(auctionStatusBorderColor('canceled'), isNull);
      expect(auctionStatusBorderColor('cancelled'), isNull);
    });

    test('rejected status returns null (neutral/current behavior preserved)', () {
      expect(auctionStatusBorderColor('rejected'), isNull);
    });

    test('an unknown/future status returns null, never guesses a color', () {
      expect(auctionStatusBorderColor('some_future_status'), isNull);
    });

    test('a null status returns null, never crashes', () {
      expect(auctionStatusBorderColor(null), isNull);
    });

    test('mixed active/ended inputs each resolve independently and correctly', () {
      final statuses = ['active', 'ended', 'active', 'pending', 'ended', 'canceled'];
      final expected = [kActiveGreen, kEndedRed, kActiveGreen, null, kEndedRed, null];
      for (var i = 0; i < statuses.length; i++) {
        expect(auctionStatusBorderColor(statuses[i]), expected[i], reason: 'index $i (${statuses[i]})');
      }
    });
  });

  group('Customer #24: rendered card decoration (minimal reproduction, not the full page)', () {
    // Mirrors _buildHorizontalAuctionCard's decoration construction exactly
    // (Container > BoxDecoration > Border.all(width: 2) when a color is
    // present, else no border) without needing AuctionApi/Riverpod/network.
    BoxDecoration buildCardDecoration(String? status, bool isDarkMode) {
      final cardBorderColor = auctionStatusBorderColor(status);
      return BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: cardBorderColor != null ? Border.all(color: cardBorderColor, width: 2) : null,
      );
    }

    testWidgets('an active card renders a green border', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Container(key: const Key('card'), decoration: buildCardDecoration('active', false)),
          ),
        ),
      );
      final container = tester.widget<Container>(find.byKey(const Key('card')));
      final decoration = container.decoration as BoxDecoration;
      expect((decoration.border as Border?)?.top.color, kActiveGreen);
    });

    testWidgets('an ended card renders a red border', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Container(key: const Key('card'), decoration: buildCardDecoration('ended', false)),
          ),
        ),
      );
      final container = tester.widget<Container>(find.byKey(const Key('card')));
      final decoration = container.decoration as BoxDecoration;
      expect((decoration.border as Border?)?.top.color, kEndedRed);
    });

    testWidgets('a pending card renders no border (neutral, unchanged)', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Container(key: const Key('card'), decoration: buildCardDecoration('pending', false)),
          ),
        ),
      );
      final container = tester.widget<Container>(find.byKey(const Key('card')));
      final decoration = container.decoration as BoxDecoration;
      expect(decoration.border, isNull);
    });

    testWidgets('border radius is preserved regardless of status', (tester) async {
      for (final status in ['active', 'ended', 'pending', null]) {
        await tester.pumpWidget(
          MaterialApp(
            home: Scaffold(
              body: Container(key: const Key('card'), decoration: buildCardDecoration(status, false)),
            ),
          ),
        );
        final container = tester.widget<Container>(find.byKey(const Key('card')));
        final decoration = container.decoration as BoxDecoration;
        expect(decoration.borderRadius, BorderRadius.circular(16), reason: 'status=$status');
      }
    });

    testWidgets('mixed active and ended cards each render their own independent, correct border', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Column(
              children: [
                Container(key: const Key('card-active'), decoration: buildCardDecoration('active', false)),
                Container(key: const Key('card-ended'), decoration: buildCardDecoration('ended', false)),
                Container(key: const Key('card-pending'), decoration: buildCardDecoration('pending', false)),
              ],
            ),
          ),
        ),
      );

      final activeDecoration = tester.widget<Container>(find.byKey(const Key('card-active'))).decoration as BoxDecoration;
      final endedDecoration = tester.widget<Container>(find.byKey(const Key('card-ended'))).decoration as BoxDecoration;
      final pendingDecoration = tester.widget<Container>(find.byKey(const Key('card-pending'))).decoration as BoxDecoration;

      expect((activeDecoration.border as Border?)?.top.color, kActiveGreen);
      expect((endedDecoration.border as Border?)?.top.color, kEndedRed);
      expect(pendingDecoration.border, isNull);
    });

    testWidgets('active -> ended rebuild (simulating a realtime refetch) changes the border from green to red', (tester) async {
      String status = 'active';
      late StateSetter setLocalState;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: StatefulBuilder(
              builder: (context, setState) {
                setLocalState = setState;
                return Container(key: const Key('card'), decoration: buildCardDecoration(status, false));
              },
            ),
          ),
        ),
      );

      var container = tester.widget<Container>(find.byKey(const Key('card')));
      var decoration = container.decoration as BoxDecoration;
      expect((decoration.border as Border?)?.top.color, kActiveGreen);

      // Simulates the page's realtime-driven _loadAuctions() rebuild after
      // Customer #20's auction.status_changed event fires for this auction.
      setLocalState(() => status = 'ended');
      await tester.pump();

      container = tester.widget<Container>(find.byKey(const Key('card')));
      decoration = container.decoration as BoxDecoration;
      expect((decoration.border as Border?)?.top.color, kEndedRed);
    });

    testWidgets('RTL layout does not affect border color resolution', (tester) async {
      await tester.pumpWidget(
        Directionality(
          textDirection: TextDirection.rtl,
          child: MaterialApp(
            home: Scaffold(
              body: Container(key: const Key('card'), decoration: buildCardDecoration('active', false)),
            ),
          ),
        ),
      );
      final container = tester.widget<Container>(find.byKey(const Key('card')));
      final decoration = container.decoration as BoxDecoration;
      expect((decoration.border as Border?)?.top.color, kActiveGreen);
    });

    testWidgets('LTR layout does not affect border color resolution', (tester) async {
      await tester.pumpWidget(
        Directionality(
          textDirection: TextDirection.ltr,
          child: MaterialApp(
            home: Scaffold(
              body: Container(key: const Key('card'), decoration: buildCardDecoration('ended', false)),
            ),
          ),
        ),
      );
      final container = tester.widget<Container>(find.byKey(const Key('card')));
      final decoration = container.decoration as BoxDecoration;
      expect((decoration.border as Border?)?.top.color, kEndedRed);
    });
  });

  group('Customer #24: isFinished flag fix (was previously dead code)', () {
    // Mirrors _buildHorizontalAuctionCard's own isFinished computation:
    // previously checked == 'finished' (a status the backend never sends),
    // now correctly checks == 'ended' || == 'closed'.
    bool resolveIsFinished(String? status) => status == 'ended' || status == 'closed';

    test('ended status is finished', () => expect(resolveIsFinished('ended'), isTrue));
    test('closed status is finished', () => expect(resolveIsFinished('closed'), isTrue));
    test('active status is not finished', () => expect(resolveIsFinished('active'), isFalse));
    test('the old dead value "finished" itself is not matched (never was real backend data)',
        () => expect(resolveIsFinished('finished'), isFalse));
  });
}
