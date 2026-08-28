import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/widgets/auction_description_section.dart';

Future<void> _pump(WidgetTester tester, {required String description}) async {
  await tester.pumpWidget(
    MaterialApp(
      locale: const Locale('ar'),
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: Scaffold(
        body: AuctionDescriptionSection(
          description: description,
          isDarkMode: false,
        ),
      ),
    ),
  );
}

void main() {
  testWidgets('renders the description label and text when description is non-empty', (tester) async {
    await _pump(tester, description: 'A detailed, non-empty auction description.');

    final l10n = await AppLocalizations.delegate.load(const Locale('ar'));
    expect(find.text(l10n.label_description), findsOneWidget);
    expect(find.text('A detailed, non-empty auction description.'), findsOneWidget);
  });

  testWidgets('renders nothing (SizedBox.shrink) when description is empty', (tester) async {
    await _pump(tester, description: '');

    final l10n = await AppLocalizations.delegate.load(const Locale('ar'));
    expect(find.text(l10n.label_description), findsNothing);
    expect(find.byType(Container), findsNothing);
  });

  testWidgets('renders nothing when description is only whitespace', (tester) async {
    await _pump(tester, description: '   ');

    final l10n = await AppLocalizations.delegate.load(const Locale('ar'));
    expect(find.text(l10n.label_description), findsNothing);
  });
}
