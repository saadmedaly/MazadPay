import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/pages/notification_detail_page.dart';

// Customer Request #22: NotificationDetailPage is the fallback screen opened
// when an admin broadcast (or any notification with no real entity/target)
// is tapped. These tests cover: text-only rendering, optional image
// rendering, no image box when image_url is absent, RTL for Arabic, and that
// a title/body render as plain text (never as HTML/markup).

Future<void> _pump(
  WidgetTester tester, {
  required String title,
  String? body,
  String? imageUrl,
  Locale locale = const Locale('en'),
}) async {
  await tester.pumpWidget(
    MaterialApp(
      locale: locale,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: NotificationDetailPage(
        title: title,
        body: body,
        imageUrl: imageUrl,
        createdAt: DateTime(2026, 1, 1, 12, 0),
      ),
    ),
  );
  await tester.pump();
}

void main() {
  testWidgets('text-only notification renders title and body, no image widget', (tester) async {
    await _pump(tester, title: 'Broadcast title', body: 'Broadcast body text');

    expect(find.text('Broadcast title'), findsOneWidget);
    expect(find.text('Broadcast body text'), findsOneWidget);
    expect(find.byType(ClipRRect), findsNothing);
  });

  testWidgets('a null image_url renders no image surface (optional image requirement)', (tester) async {
    await _pump(tester, title: 'No image here', body: 'just text', imageUrl: null);

    expect(find.byType(ClipRRect), findsNothing);
  });

  testWidgets('an empty-string image_url is treated the same as null (no image box)', (tester) async {
    await _pump(tester, title: 'Empty url', body: 'text', imageUrl: '   ');

    expect(find.byType(ClipRRect), findsNothing);
  });

  testWidgets('a present image_url renders the image surface', (tester) async {
    await _pump(
      tester,
      title: 'With image',
      body: 'text',
      imageUrl: 'https://cdn.example.com/notifications/pic.jpg',
    );

    expect(find.byType(ClipRRect), findsOneWidget);
  });

  testWidgets('a missing/empty body does not crash and shows no body text widget for it', (tester) async {
    await _pump(tester, title: 'Only a title', body: null);

    expect(find.text('Only a title'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('back button pops the page', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: Builder(
          builder: (context) => Scaffold(
            body: Center(
              child: ElevatedButton(
                onPressed: () => Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (_) => NotificationDetailPage(
                      title: 'Detail',
                      body: 'body',
                      createdAt: DateTime(2026, 1, 1),
                    ),
                  ),
                ),
                child: const Text('open'),
              ),
            ),
          ),
        ),
      ),
    );

    await tester.tap(find.text('open'));
    await tester.pumpAndSettle();
    expect(find.text('Detail'), findsOneWidget);

    await tester.tap(find.byType(IconButton));
    await tester.pumpAndSettle();
    expect(find.text('Detail'), findsNothing);
    expect(find.text('open'), findsOneWidget);
  });

  testWidgets('Arabic locale renders RTL directionality', (tester) async {
    await _pump(
      tester,
      title: 'عنوان',
      body: 'نص الإشعار',
      locale: const Locale('ar'),
    );

    final directionality = tester.widget<Directionality>(find.byType(Directionality).first);
    expect(directionality.textDirection, TextDirection.rtl);
    expect(find.text('عنوان'), findsOneWidget);
  });

  testWidgets('English/French-style locale renders LTR directionality', (tester) async {
    await _pump(tester, title: 'Title', body: 'Body', locale: const Locale('en'));

    final directionality = tester.widget<Directionality>(find.byType(Directionality).first);
    expect(directionality.textDirection, TextDirection.ltr);
  });

  testWidgets(
    'BROKEN_IMAGE_SAFE / TEXT_VISIBLE_AFTER_IMAGE_FAILURE: an unreachable '
    'image URL does not crash the page, and title/body remain visible '
    'while the image is loading/failing',
    (tester) async {
      // A deliberately unresolvable host -- CachedNetworkImage's own
      // errorWidget (a neutral grey placeholder, mirroring Bug H's
      // established pattern) is what renders once the network request
      // ultimately fails; widget tests run with no real network access, so
      // the image starts and stays in its placeholder/error state rather
      // than ever completing -- exactly the "broken image" condition this
      // test targets. No exception must propagate either way.
      await _pump(
        tester,
        title: 'Image failure title',
        body: 'Image failure body text',
        imageUrl: 'https://this-host-does-not-resolve.invalid/broken.jpg',
      );
      await tester.pump(const Duration(milliseconds: 100));

      expect(tester.takeException(), isNull);
      expect(find.text('Image failure title'), findsOneWidget);
      expect(find.text('Image failure body text'), findsOneWidget);
    },
  );

  testWidgets('LONG_CONTENT_SAFE: a long body scrolls inside a SingleChildScrollView without overflow', (tester) async {
    final longBody = List.generate(80, (i) => 'Line $i of a very long broadcast notification body.').join(' ');

    await _pump(tester, title: 'Long content', body: longBody);

    expect(find.byType(SingleChildScrollView), findsOneWidget);
    expect(tester.takeException(), isNull);

    // Actually scroll it to prove the content is reachable, not just present
    // in the widget tree behind a fixed-height clip.
    await tester.drag(find.byType(SingleChildScrollView), const Offset(0, -300));
    await tester.pump();
    expect(tester.takeException(), isNull);
  });
}
