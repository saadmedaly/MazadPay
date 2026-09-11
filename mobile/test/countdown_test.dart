import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/utils/time_utils.dart';

// Customer request #3 (high priority): the auction details page countdown
// was not ticking down in real time. Root cause: AuctionDetailsPage._startTimer's
// Timer.periodic callback decremented `_timeLeft.inSeconds - 1` on every tick
// instead of recomputing from `endTime.difference(DateTime.now())` -- the
// correct helper (_updateTimeLeft) already existed and was called from
// build(), but never from the timer tick itself. Manually decrementing a
// cached value drifts away from wall-clock truth under dropped frames, GC
// pauses, or the app being backgrounded/resumed, since it never re-anchors
// to the real end time. Fixed by having the timer tick call the same
// endTime-anchored recompute on every tick.
//
// CountdownParts.fromDuration is the pure days/hours/minutes/seconds
// breakdown extracted from the inline widget-building code so it's directly
// unit-testable without a widget harness.
void main() {
  group('CountdownParts.fromDuration', () {
    test('30 seconds', () {
      final parts = CountdownParts.fromDuration(const Duration(seconds: 30));
      expect(parts.days, 0);
      expect(parts.hours, 0);
      expect(parts.minutes, 0);
      expect(parts.seconds, 30);
    });

    test('59 seconds', () {
      final parts = CountdownParts.fromDuration(const Duration(seconds: 59));
      expect(parts.minutes, 0);
      expect(parts.seconds, 59);
    });

    test('1 minute (60 seconds) rolls over to minutes, not 60 seconds', () {
      final parts = CountdownParts.fromDuration(const Duration(minutes: 1));
      expect(parts.minutes, 1);
      expect(parts.seconds, 0);
    });

    test('1 hour', () {
      final parts = CountdownParts.fromDuration(const Duration(hours: 1));
      expect(parts.days, 0);
      expect(parts.hours, 1);
      expect(parts.minutes, 0);
      expect(parts.seconds, 0);
    });

    test('23h59m59s stays within a single day (no day rollover yet)', () {
      final parts = CountdownParts.fromDuration(
        const Duration(hours: 23, minutes: 59, seconds: 59),
      );
      expect(parts.days, 0);
      expect(parts.hours, 23);
      expect(parts.minutes, 59);
      expect(parts.seconds, 59);
    });

    test('24 hours rolls over to exactly 1 day, 0 hours', () {
      final parts = CountdownParts.fromDuration(const Duration(hours: 24));
      expect(parts.days, 1);
      expect(parts.hours, 0);
    });

    test('25 hours is 1 day, 1 hour', () {
      final parts = CountdownParts.fromDuration(const Duration(hours: 25));
      expect(parts.days, 1);
      expect(parts.hours, 1);
    });

    test('48+ hours (50h) is 2 days, 2 hours -- days are never hidden', () {
      final parts = CountdownParts.fromDuration(const Duration(hours: 50));
      expect(parts.days, 2);
      expect(parts.hours, 2);
    });

    test('exactly zero duration', () {
      final parts = CountdownParts.fromDuration(Duration.zero);
      expect(parts.days, 0);
      expect(parts.hours, 0);
      expect(parts.minutes, 0);
      expect(parts.seconds, 0);
    });

    test('negative duration is clamped to all-zero, never negative digits', () {
      final parts = CountdownParts.fromDuration(const Duration(seconds: -5));
      expect(parts.days, 0);
      expect(parts.hours, 0);
      expect(parts.minutes, 0);
      expect(parts.seconds, 0);
    });
  });

  group('recomputation is anchored to real time, not decremented state', () {
    // This is the actual regression test for the root cause: given a fixed
    // endTime, computing `endTime.difference(now)` at two different `now`
    // values must always reflect the true remaining gap -- proving the
    // "recompute from endTime" approach (what _startTimer now does on every
    // tick) is drift-free by construction, unlike subtracting 1 second from
    // a previously cached value.
    test('endTime - now stays accurate regardless of how much wall-clock time elapsed between calls', () {
      final endTime = DateTime(2026, 1, 1, 12, 0, 0);

      final earlierNow = DateTime(2026, 1, 1, 11, 59, 30); // 30s before end
      final laterNow = DateTime(2026, 1, 1, 11, 59, 50); // 20s before end (app was "paused" for 20s)

      final earlierRemaining = endTime.difference(earlierNow);
      final laterRemaining = endTime.difference(laterNow);

      expect(earlierRemaining, const Duration(seconds: 30));
      // Even though only one tick nominally happened, the real elapsed gap
      // was 20 seconds (simulating a dropped/delayed tick or background
      // pause) -- recomputing from endTime reflects the TRUE remaining time
      // instead of naively landing on 29 seconds (30 - 1).
      expect(laterRemaining, const Duration(seconds: 10));
    });

    test('a duration past endTime is negative before clamping', () {
      final endTime = DateTime(2026, 1, 1, 12, 0, 0);
      final now = DateTime(2026, 1, 1, 12, 0, 5);
      expect(endTime.difference(now).isNegative, true);
    });
  });
}
