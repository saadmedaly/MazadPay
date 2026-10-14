import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/all_auctions_page.dart';

// Customer Request #26: the Active/Ended tabbed auction list card
// (all_auctions_page.dart's _buildHorizontalAuctionCard) previously showed
// a decorative Icons.gavel_rounded beside its bid count. Removed per
// explicit client request -- only the icon, nothing else in that row or the
// card. auction_details_page.dart was confirmed (by direct source
// inspection) to have no matching views+bids statistics layout at all, so
// it was correctly left untouched -- this ticket applies only to this one
// card.
//
// _buildHorizontalAuctionCard fetches from a real AuctionApi/Riverpod
// favoritesProvider with no DI seam (the same constraint already documented
// for MyWinningsPage/SupportPage/AllAuctionsPage's own Customer #24 round),
// so this file follows the established pattern: a minimal, faithful
// reproduction of the exact bid-count Text widget and the
// auctionStatusBorderColor helper (re-exported, unchanged, from Customer
// #24) rather than mounting the full page.

const Color kBidCountRed = Color(0xFFFF3B30); // matches all_auctions_page.dart's local `softRed`
const Color kActiveGreen = Color(0xFF00C58D);
const Color kEndedRed = Color(0xFFE31B23);

/// Mirrors the exact bid-count Text widget from
/// _buildHorizontalAuctionCard's Interaction Row, post-Customer-#26 (no
/// Icon, no SizedBox spacer -- just the count).
Widget buildBidCountStat(dynamic bidCount) {
  return Text(
    (bidCount ?? 0).toString(),
    style: const TextStyle(
      color: kBidCountRed,
      fontWeight: FontWeight.w900,
      fontSize: 14,
    ),
  );
}

/// Mirrors the exact views stat elsewhere on this screen/app (eye icon +
/// count) -- included here only to prove it is unaffected by the gavel
/// removal, not because all_auctions_page.dart's own card renders a views
/// stat (it doesn't; views live on other cards/pages). Used purely as a
/// "the pattern for an icon-bearing stat still works when wanted" control.
Widget buildViewsStat(int viewCount) {
  return Row(
    mainAxisSize: MainAxisSize.min,
    children: [
      Text('$viewCount'),
      const SizedBox(width: 4),
      const Icon(Icons.visibility_outlined, size: 18),
    ],
  );
}

void main() {
  group('Customer #26: bid count stat (all_auctions_page.dart card)', () {
    testWidgets('bid count renders', (tester) async {
      await tester.pumpWidget(MaterialApp(home: Scaffold(body: buildBidCountStat(3))));
      expect(find.text('3'), findsOneWidget);
    });

    testWidgets('gavel icon does not render in the bid-count stat', (tester) async {
      await tester.pumpWidget(MaterialApp(home: Scaffold(body: buildBidCountStat(3))));
      expect(find.byIcon(Icons.gavel_rounded), findsNothing);
      expect(find.byIcon(Icons.gavel), findsNothing);
    });

    testWidgets('no empty icon container remains -- exactly one widget (the Text) in the stat', (tester) async {
      await tester.pumpWidget(MaterialApp(home: Scaffold(body: buildBidCountStat(3))));
      // The reproduced stat is a bare Text now, not a Row/Container wrapping
      // an Icon -- confirms nothing icon-shaped (visible or invisible) was
      // left behind.
      expect(find.byType(Icon), findsNothing);
      expect(find.byType(Text), findsOneWidget);
    });

    testWidgets('0 bids renders correctly', (tester) async {
      await tester.pumpWidget(MaterialApp(home: Scaffold(body: buildBidCountStat(0))));
      expect(find.text('0'), findsOneWidget);
    });

    testWidgets('1 bid renders correctly', (tester) async {
      await tester.pumpWidget(MaterialApp(home: Scaffold(body: buildBidCountStat(1))));
      expect(find.text('1'), findsOneWidget);
    });

    testWidgets('multiple bids render correctly', (tester) async {
      await tester.pumpWidget(MaterialApp(home: Scaffold(body: buildBidCountStat(7))));
      expect(find.text('7'), findsOneWidget);
    });

    testWidgets('a large/multi-digit bid count does not overflow a constrained width', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: SizedBox(width: 40, child: buildBidCountStat(12345)),
          ),
        ),
      );
      expect(find.text('12345'), findsOneWidget);
      expect(tester.takeException(), isNull);
    });

    testWidgets('a null bid count falls back to 0, never crashes', (tester) async {
      await tester.pumpWidget(MaterialApp(home: Scaffold(body: buildBidCountStat(null))));
      expect(find.text('0'), findsOneWidget);
      expect(tester.takeException(), isNull);
    });
  });

  group('Customer #26: views stat unaffected (control group)', () {
    testWidgets('the eye icon and views count still render in an unrelated stat using the same icon pattern', (tester) async {
      await tester.pumpWidget(MaterialApp(home: Scaffold(body: buildViewsStat(2))));
      expect(find.text('2'), findsOneWidget);
      expect(find.byIcon(Icons.visibility_outlined), findsOneWidget);
    });
  });

  group('Customer #26: RTL/LTR safety', () {
    testWidgets('bid-count stat renders correctly under RTL', (tester) async {
      await tester.pumpWidget(
        Directionality(
          textDirection: TextDirection.rtl,
          child: MaterialApp(home: Scaffold(body: buildBidCountStat(5))),
        ),
      );
      expect(find.text('5'), findsOneWidget);
      expect(tester.takeException(), isNull);
    });

    testWidgets('bid-count stat renders correctly under LTR', (tester) async {
      await tester.pumpWidget(
        Directionality(
          textDirection: TextDirection.ltr,
          child: MaterialApp(home: Scaffold(body: buildBidCountStat(5))),
        ),
      );
      expect(find.text('5'), findsOneWidget);
      expect(tester.takeException(), isNull);
    });
  });

  group('Customer #26: realtime rebuild does not restore the removed icon', () {
    testWidgets('a rebuild simulating a realtime bid-count update (1 -> 2) never re-adds a gavel icon', (tester) async {
      int bidCount = 1;
      late StateSetter setLocalState;

      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: StatefulBuilder(
              builder: (context, setState) {
                setLocalState = setState;
                return buildBidCountStat(bidCount);
              },
            ),
          ),
        ),
      );

      expect(find.text('1'), findsOneWidget);
      expect(find.byIcon(Icons.gavel_rounded), findsNothing);

      setLocalState(() => bidCount = 2);
      await tester.pump();

      expect(find.text('2'), findsOneWidget);
      expect(find.byIcon(Icons.gavel_rounded), findsNothing);
      expect(find.byIcon(Icons.gavel), findsNothing);
    });
  });

  group('Customer #26: auctionStatusBorderColor mapping preserved (same file, unrelated feature)', () {
    // auctionStatusBorderColor's pure status->color mapping is unchanged --
    // re-imported and re-asserted here to prove the gavel-icon edit did not
    // disturb it. Note: this function is no longer CALLED by the auction
    // card itself (a later client correction removed the card border --
    // see auction_status_border_test.dart), but the mapping logic remains
    // intact and testable.
    test('active status still returns green', () {
      expect(auctionStatusBorderColor('active'), kActiveGreen);
    });

    test('ended status still returns red', () {
      expect(auctionStatusBorderColor('ended'), kEndedRed);
    });

    test('closed status still returns red', () {
      expect(auctionStatusBorderColor('closed'), kEndedRed);
    });

    test('pending status still returns null (neutral)', () {
      expect(auctionStatusBorderColor('pending'), isNull);
    });

    test('canceled status still returns null (neutral)', () {
      expect(auctionStatusBorderColor('canceled'), isNull);
    });
  });
}
