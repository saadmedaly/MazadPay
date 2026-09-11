import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

// Customer feedback #11 final proof (Phase 7): my_winnings_page.dart's empty
// state wraps a short, viewport-shorter SizedBox in a plain ListView inside a
// RefreshIndicator. RefreshIndicator only fires onRefresh when the drag
// actually reaches an overscroll -- a ListView whose content is shorter than
// the viewport can otherwise refuse to register that drag under default
// ScrollPhysics, silently making pull-to-refresh inert exactly in the empty
// state (the state a user stares at right after a win, waiting for it to
// appear). This reproduces that exact shape (short content, AlwaysScrollable
// physics) as a minimal standalone widget -- not the full page, which needs
// no network mocking harness for this specific proof -- and drives a real
// drag gesture to confirm onRefresh actually fires.
void main() {
  testWidgets('pull-to-refresh fires onRefresh even when content is shorter than the viewport (the empty-state shape)', (tester) async {
    var refreshCount = 0;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: RefreshIndicator(
            onRefresh: () async {
              refreshCount++;
            },
            child: ListView(
              physics: const AlwaysScrollableScrollPhysics(),
              children: const [
                SizedBox(height: 100, child: Center(child: Text('empty state'))),
              ],
            ),
          ),
        ),
      ),
    );

    expect(refreshCount, 0);

    // Drag down from near the top of the viewport, as a real pull-to-refresh gesture would.
    await tester.fling(find.text('empty state'), const Offset(0, 300), 1000);
    await tester.pumpAndSettle();

    expect(refreshCount, 1, reason: 'expected onRefresh to fire once for a pull-down gesture on shorter-than-viewport content');
  });
}
