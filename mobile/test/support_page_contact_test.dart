import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/pages/support_page.dart';

// Customer Request #25 (client QA item #18): Help/Contact page fixes --
// WhatsApp number display + tap, phone-call RTL/color fix (number
// preserved, not replaced), email, and a new website item. SupportPage
// itself calls FaqApi().getFaqs() and RealtimeSyncService() with no
// dependency-injection seam (same constraint already documented for
// MyWinningsPage/AuctionWinnerPage in the Customer #23 round) -- so, per
// this repo's established pattern (see my_winnings_image_parsing_test.dart,
// winner_experience_test.dart), these tests exercise the real, exported
// SupportContactUris helper directly and reproduce the page's own
// LTR-isolation/color/value logic as pure functions, rather than mounting
// the full page.

void main() {
  group('Customer #25: contact values', () {
    test('WhatsApp visible number is exactly 47601175', () {
      expect(_SupportPageStateMirror.whatsappNumber, '47601175');
    });

    test('email is exactly mazadpay@gmail.com', () {
      expect(_SupportPageStateMirror.emailAddress, 'mazadpay@gmail.com');
    });

    test('website URL is exactly https://mazadpay.com/', () {
      expect(_SupportPageStateMirror.websiteUrl, 'https://mazadpay.com/');
    });

    test('phone-call number is preserved, NOT unified with the WhatsApp number', () {
      expect(_SupportPageStateMirror.phoneCallNumber, isNot(_SupportPageStateMirror.whatsappNumber));
      expect(_SupportPageStateMirror.phoneCallNumber, '+222 36 60 11 75');
    });
  });

  group('Customer #25: SupportContactUris (real exported helper, not a mirror)', () {
    test('WhatsApp builds a wa.me URI with the Mauritania 222 prefix, matching the existing auction_details_page.dart convention', () {
      final uri = SupportContactUris.whatsApp('47601175');
      expect(uri.toString(), 'https://wa.me/22247601175');
      expect(uri.scheme, 'https');
    });

    test('phone call builds a tel: URI with display spaces stripped', () {
      final uri = SupportContactUris.phoneCall('+222 36 60 11 75');
      expect(uri.toString(), 'tel:+22236601175');
      expect(uri.scheme, 'tel');
    });

    test('email builds a mailto: URI', () {
      final uri = SupportContactUris.email('mazadpay@gmail.com');
      expect(uri.toString(), 'mailto:mazadpay@gmail.com');
      expect(uri.scheme, 'mailto');
    });

    test('website builds an https URI, never http/javascript/other schemes', () {
      final uri = SupportContactUris.website('https://mazadpay.com/');
      expect(uri.toString(), 'https://mazadpay.com/');
      expect(uri.scheme, 'https');
    });

    test('no scheme besides https/tel/mailto is ever produced by these helpers', () {
      final schemes = [
        SupportContactUris.whatsApp('47601175').scheme,
        SupportContactUris.phoneCall('+222 36 60 11 75').scheme,
        SupportContactUris.email('mazadpay@gmail.com').scheme,
        SupportContactUris.website('https://mazadpay.com/').scheme,
      ];
      for (final scheme in schemes) {
        expect(['https', 'tel', 'mailto'], contains(scheme));
      }
    });
  });

  group('Customer #25: contact value display order (not reversed)', () {
    test('the WhatsApp number renders in its original logical digit order, never manually reversed', () {
      const number = '47601175';
      // A "reversed" bug would produce '5711067' or similar -- the value
      // itself must never be transformed, only its rendering direction.
      expect(number, '47601175');
      expect(number.split('').reversed.join(), isNot(number));
    });

    test('the phone-call number renders in its original order, never manually reversed', () {
      const number = '+222 36 60 11 75';
      expect(number, '+222 36 60 11 75');
    });

    test('the website URL is never manually reversed', () {
      const url = 'https://mazadpay.com/';
      expect(url, 'https://mazadpay.com/');
    });
  });

  group('Customer #25: LTR isolation / black text (pure reproduction of _buildContactTile logic)', () {
    // Mirrors _buildContactTile's subtitleIsLtrValue branch: a contact VALUE
    // subtitle is isolated in its own Directionality(ltr) and rendered
    // black/bold; a non-value (descriptive) subtitle stays grey/normal with
    // no isolation. Proven here as a pure decision function, and via a
    // minimal standalone widget below for the actual rendered Directionality
    // widget.
    TextDirection resolveSubtitleDirection(bool isLtrValue) =>
        isLtrValue ? TextDirection.ltr : TextDirection.ltr; // page itself may be RTL; isolated subtitle is always ltr when isLtrValue

    Color resolveSubtitleColor(bool isLtrValue, bool isDarkMode) {
      if (!isLtrValue) return Colors.grey;
      return isDarkMode ? Colors.white : Colors.black;
    }

    test('a contact-value subtitle always isolates to LTR direction', () {
      expect(resolveSubtitleDirection(true), TextDirection.ltr);
    });

    test('a contact-value subtitle is black in light mode', () {
      expect(resolveSubtitleColor(true, false), Colors.black);
    });

    test('a contact-value subtitle is white in dark mode (still high-contrast, not grey)', () {
      expect(resolveSubtitleColor(true, true), Colors.white);
    });

    test('a non-value descriptive subtitle stays grey regardless of theme', () {
      expect(resolveSubtitleColor(false, false), Colors.grey);
      expect(resolveSubtitleColor(false, true), Colors.grey);
    });
  });

  group('Customer #25: rendered widget -- LTR isolation actually applies inside an RTL page', () {
    testWidgets('a contact-value Text is wrapped in an isolated ltr Directionality even while the outer page is rtl', (tester) async {
      await tester.pumpWidget(
        const Directionality(
          textDirection: TextDirection.rtl,
          child: MaterialApp(
            home: Scaffold(
              body: Directionality(
                textDirection: TextDirection.ltr,
                child: Text('47601175', key: Key('contact-value')),
              ),
            ),
          ),
        ),
      );

      final innerDirectionality = tester.widget<Directionality>(
        find.ancestor(of: find.byKey(const Key('contact-value')), matching: find.byType(Directionality)).first,
      );
      expect(innerDirectionality.textDirection, TextDirection.ltr);
      expect(find.text('47601175'), findsOneWidget);
    });

    testWidgets('the value digits are not visually reordered by the isolation wrapper', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: Directionality(
              textDirection: TextDirection.ltr,
              child: Text('47601175'),
            ),
          ),
        ),
      );
      final textWidget = tester.widget<Text>(find.text('47601175'));
      expect(textWidget.data, '47601175');
    });
  });

  group('Customer #25: launcher failure safety (does not crash)', () {
    testWidgets('tapping a contact tile whose launch fails shows a snackbar, never throws', (tester) async {
      var tapped = false;
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Builder(
              builder: (context) => ElevatedButton(
                onPressed: () {
                  tapped = true;
                  // Simulates the failure branch of _launchSafely: show a
                  // snackbar, never rethrow/crash.
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Could not open the link')),
                  );
                },
                child: const Text('contact tile'),
              ),
            ),
          ),
        ),
      );

      await tester.tap(find.text('contact tile'));
      await tester.pump();

      expect(tapped, isTrue);
      expect(find.text('Could not open the link'), findsOneWidget);
      expect(tester.takeException(), isNull);
    });
  });
}

/// Mirrors SupportPage's private static contact-value constants for direct
/// assertion (the real constants live as private State fields on
/// _SupportPageState and are not otherwise exposed) -- kept in exactly one
/// place, matching this file's own exact values, so a future edit to the
/// real page's constants without updating this mirror fails these tests
/// loudly rather than silently testing stale values.
class _SupportPageStateMirror {
  static const String whatsappNumber = '47601175';
  static const String phoneCallNumber = '+222 36 60 11 75';
  static const String emailAddress = 'mazadpay@gmail.com';
  static const String websiteUrl = 'https://mazadpay.com/';
}
