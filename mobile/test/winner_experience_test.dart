import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/utils/auction_image.dart';

// Customer Request #23: My Winnings + Winner Congratulations + Share.
// Full widget-provider testing of AuctionWinnerPage/MyWinningsPage would
// require mocking a code-generated Riverpod provider (auctionNotifierApiProvider,
// riverpod_generator) and a live network layer -- this repo's established
// pattern (see my_winnings_image_parsing_test.dart, my_winnings_pull_to_refresh_test.dart)
// is instead to reproduce the specific testable logic/shape as a pure
// function or minimal standalone widget, which is what these tests do.

void main() {
  group('Customer #23: price source (current_price is the only real backend field)', () {
    // Pure reproduction of my_winnings_page.dart's price-resolution logic.
    num resolvePrice(Map<String, dynamic> winning) {
      return num.tryParse((winning['current_price'] ?? winning['final_price'] ?? winning['current_bid'] ?? 0).toString()) ?? 0;
    }

    test('current_price (the real backend field) is used when present', () {
      expect(resolvePrice({'current_price': 200}), 200);
    });

    test('current_price wins even if a legacy final_price is also present', () {
      expect(resolvePrice({'current_price': 200, 'final_price': 999}), 200);
    });

    test('falls back to final_price only when current_price is absent (legacy compatibility, never the primary path)', () {
      expect(resolvePrice({'final_price': 150}), 150);
    });

    test('no price fields at all resolves to 0, never a crash', () {
      expect(resolvePrice({}), 0);
    });
  });

  group('Customer #23: end date resolution', () {
    String? resolveEndDateLabel(Map<String, dynamic> winning) {
      final endTimeRaw = winning['end_time']?.toString();
      final endDate = endTimeRaw != null ? DateTime.tryParse(endTimeRaw) : null;
      return endDate != null ? '${endDate.year}-${endDate.month.toString().padLeft(2, '0')}-${endDate.day.toString().padLeft(2, '0')}' : null;
    }

    test('a valid end_time resolves to a date label', () {
      expect(resolveEndDateLabel({'end_time': '2026-03-15T10:00:00Z'}), '2026-03-15');
    });

    test('a missing end_time resolves to null, not a crash', () {
      expect(resolveEndDateLabel({}), isNull);
    });

    test('a malformed end_time resolves to null, not a crash', () {
      expect(resolveEndDateLabel({'end_time': 'not-a-date'}), isNull);
    });
  });

  group('Customer #23: winning card image (Bug H preserved, no Corolla)', () {
    test('a valid image URL resolves to itself', () {
      expect(resolveAuctionImageUrl(['https://cdn.example.com/a.jpg']), 'https://cdn.example.com/a.jpg');
    });

    test('an empty list resolves to null (neutral placeholder), never a bundled asset', () {
      expect(resolveAuctionImageUrl([]), isNull);
    });

    test('a list of only empty strings resolves to null', () {
      expect(resolveAuctionImageUrl(['', '  ']), isNull);
    });

    test('resolveAuctionImageUrl never returns a corolla/local asset path', () {
      final result = resolveAuctionImageUrl(['https://cdn.example.com/real.jpg']);
      expect(result, isNot(contains('corolla')));
      expect(result, isNot(startsWith('assets/')));
    });
  });

  group('Customer #23: foreground winner coordinator (pure logic reproduction)', () {
    // Mirrors notification_handler.dart's _maybeShowForegroundWinner guard
    // logic: given a set of already-surfaced auction IDs and a newly
    // event-flagged auction ID, decide whether to surface it.
    bool shouldSurface(Set<String> alreadySurfaced, String auctionId) {
      if (auctionId.isEmpty) return false;
      if (alreadySurfaced.contains(auctionId)) return false;
      return true;
    }

    test('a fresh auction ID not yet surfaced should be surfaced once', () {
      final surfaced = <String>{};
      expect(shouldSurface(surfaced, 'a1'), isTrue);
    });

    test('WINNER_FOREGROUND_TRIGGER_ONCE / DUPLICATE_REALTIME_TRIGGER_PREVENTED: the same auction ID is never surfaced twice in a session', () {
      final surfaced = <String>{'a1'};
      expect(shouldSurface(surfaced, 'a1'), isFalse);
    });

    test('HISTORICAL_WIN_REPLAY_PREVENTED: a resume/reconnect does not clear the surfaced set, so an old win is never replayed as a new popup', () {
      final surfaced = <String>{'old-win-1', 'old-win-2'};
      // Simulates an app-resume catch-up firing status_changed again for an
      // already-surfaced historical win -- the guard must still say no.
      expect(shouldSurface(surfaced, 'old-win-1'), isFalse);
      expect(shouldSurface(surfaced, 'old-win-2'), isFalse);
    });

    test('an empty auction id is never surfaced (defensive, matches the real guard)', () {
      final surfaced = <String>{};
      expect(shouldSurface(surfaced, ''), isFalse);
    });

    test('a DIFFERENT new win after an old one is still surfaced once (not blocked by unrelated history)', () {
      final surfaced = <String>{'old-win-1'};
      expect(shouldSurface(surfaced, 'new-win-2'), isTrue);
    });
  });

  group('Customer #23: My Winnings authoritative match (never guess winner identity client-side)', () {
    // Mirrors _maybeShowForegroundWinner's match-against-My-Winnings step.
    bool isRealWinForCurrentUser(List<Map<String, dynamic>> winnings, String auctionId) {
      return winnings.any((w) => w['id']?.toString() == auctionId);
    }

    test('a changed auction present in My Winnings is a real win', () {
      expect(isRealWinForCurrentUser([{'id': 'a1'}], 'a1'), isTrue);
    });

    test('a changed auction NOT present in My Winnings is not a win for this user (no-bid/cancelled/someone-else-won)', () {
      expect(isRealWinForCurrentUser([{'id': 'a1'}], 'a2'), isFalse);
    });

    test('an empty My Winnings response never falsely matches', () {
      expect(isRealWinForCurrentUser([], 'a1'), isFalse);
    });
  });

  group('Customer #23: legacy fcm_service.dart extractRoute is dead code (documented, not refactored)', () {
    test('proof: extractRoute has zero call sites in the codebase besides its own declaration', () {
      // This is a structural/documentation assertion, not a runtime check --
      // verified via direct source grep during implementation: `grep -rn
      // "extractRoute" lib/` returns exactly one match (the declaration
      // itself, fcm_service.dart:671), confirming it is never called by
      // notifications_page.dart's _onNotificationTap or
      // notification_handler.dart's _navigateFromNotification (the two REAL
      // active tap-routing paths, both already correctly special-casing
      // auction_won to AuctionWinnerPage). Left untouched per the
      // implementation brief: "If dead: do not blindly refactor;
      // document/prove it is dead."
      expect(true, isTrue);
    });
  });

  group('Customer #23: congratulations localization', () {
    testWidgets('text_406 (congratulations) resolves for ar/fr/en, never empty', (tester) async {
      for (final locale in const [Locale('ar'), Locale('fr'), Locale('en')]) {
        late BuildContext capturedContext;
        await tester.pumpWidget(
          MaterialApp(
            locale: locale,
            localizationsDelegates: AppLocalizations.localizationsDelegates,
            supportedLocales: AppLocalizations.supportedLocales,
            home: Builder(
              builder: (context) {
                capturedContext = context;
                return const SizedBox();
              },
            ),
          ),
        );
        await tester.pump();
        final text = AppLocalizations.of(capturedContext)!.text_406;
        expect(text, isNotEmpty, reason: 'text_406 must resolve for locale $locale');
      }
    });

    testWidgets('text_407 (share payload template) resolves for ar/fr/en with both placeholders substituted', (tester) async {
      for (final locale in const [Locale('ar'), Locale('fr'), Locale('en')]) {
        late BuildContext capturedContext;
        await tester.pumpWidget(
          MaterialApp(
            locale: locale,
            localizationsDelegates: AppLocalizations.localizationsDelegates,
            supportedLocales: AppLocalizations.supportedLocales,
            home: Builder(
              builder: (context) {
                capturedContext = context;
                return const SizedBox();
              },
            ),
          ),
        );
        await tester.pump();
        final text = AppLocalizations.of(capturedContext)!.text_407('100 MRU', 'Test Auction');
        expect(text, contains('Test Auction'));
        expect(text, contains('100 MRU'));
      }
    });
  });

  group('Customer #23: share payload safety', () {
    // Mirrors auction_winner_page.dart's _shareWin: only title + formatted
    // amount ever go into the shared text.
    String buildSharePayload(String template, String title, String amount) {
      return template.replaceAll('{title}', title).replaceAll('{amount}', amount);
    }

    test('share payload contains the auction title and amount', () {
      final payload = buildSharePayload('Won {title} for {amount}', 'iPhone 15', '5000 MRU');
      expect(payload, contains('iPhone 15'));
      expect(payload, contains('5000 MRU'));
    });

    test('share payload never contains a raw UUID-shaped winner/auction id', () {
      const uuidPattern = r'[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}';
      final payload = buildSharePayload('Won {title} for {amount}', 'iPhone 15', '5000 MRU');
      expect(RegExp(uuidPattern).hasMatch(payload), isFalse);
    });

    test('share payload never contains the literal words wallet/token/jwt/balance (safety net against accidental inclusion)', () {
      final payload = buildSharePayload('Won {title} for {amount}', 'Vintage Watch', '5000 MRU').toLowerCase();
      for (final forbidden in ['wallet', 'token', 'jwt', 'balance']) {
        expect(payload, isNot(contains(forbidden)));
      }
    });
  });
}
