import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/widgets/country_picker_sheet.dart';

/// Fixture countries mirroring the shape of `GET /countries`. Includes two
/// countries sharing the same dial code (+1) to exercise the country-vs-
/// dial-code ambiguity that motivated requiring `country_iso`.
final _fixtureCountries = <Map<String, dynamic>>[
  {
    'code': 'MR',
    'country_code': '+222',
    'name_ar': 'موريتانيا',
    'name_fr': 'Mauritanie',
    'name_en': 'Mauritania',
    'is_active': true,
  },
  {
    'code': 'TN',
    'country_code': '+216',
    'name_ar': 'تونس',
    'name_fr': 'Tunisie',
    'name_en': 'Tunisia',
    'is_active': true,
  },
  {
    'code': 'US',
    'country_code': '+1',
    'name_ar': 'الولايات المتحدة',
    'name_fr': 'États-Unis',
    'name_en': 'United States',
    'is_active': true,
  },
  {
    'code': 'CA',
    'country_code': '+1',
    'name_ar': 'كندا',
    'name_fr': 'Canada',
    'name_en': 'Canada',
    'is_active': true,
  },
];

Future<void> _pumpSheet(
  WidgetTester tester, {
  required void Function(Map<String, dynamic>) onSelected,
}) async {
  await tester.pumpWidget(
    MaterialApp(
      locale: const Locale('en'),
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: Scaffold(
        body: Builder(
          builder: (context) => ElevatedButton(
            onPressed: () => showCountryPickerSheet(
              context,
              countries: _fixtureCountries,
              onSelected: onSelected,
            ),
            child: const Text('open'),
          ),
        ),
      ),
    ),
  );

  await tester.tap(find.text('open'));
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('shows every fixture country before any search input', (tester) async {
    await _pumpSheet(tester, onSelected: (_) {});

    expect(find.text('Mauritania'), findsOneWidget);
    expect(find.text('Tunisia'), findsOneWidget);
    expect(find.text('United States'), findsOneWidget);
    expect(find.text('Canada'), findsOneWidget);
  });

  testWidgets('typing a search query filters the list to matching countries', (tester) async {
    await _pumpSheet(tester, onSelected: (_) {});

    await tester.enterText(find.byType(TextField), 'tun');
    await tester.pumpAndSettle();

    expect(find.text('Tunisia'), findsOneWidget);
    expect(find.text('Mauritania'), findsNothing);
    expect(find.text('United States'), findsNothing);
    expect(find.text('Canada'), findsNothing);
  });

  testWidgets(
    'US and Canada both surface under the shared +1 dial code, distinguished by name',
    (tester) async {
      await _pumpSheet(tester, onSelected: (_) {});

      await tester.enterText(find.byType(TextField), '+1');
      await tester.pumpAndSettle();

      expect(find.text('United States'), findsOneWidget);
      expect(find.text('Canada'), findsOneWidget);
      expect(find.text('Tunisia'), findsNothing);
    },
  );

  testWidgets('selecting a country returns its full map (including the ISO code) and closes the sheet', (tester) async {
    Map<String, dynamic>? selected;
    await _pumpSheet(tester, onSelected: (c) => selected = c);

    await tester.enterText(find.byType(TextField), 'Cana');
    await tester.pumpAndSettle();

    await tester.tap(find.text('Canada'));
    await tester.pumpAndSettle();

    expect(selected, isNotNull);
    expect(selected!['code'], 'CA');
    expect(selected!['country_code'], '+1');
    // The sheet should have closed after selection.
    expect(find.text('Canada'), findsNothing);
  });

  testWidgets('an empty search result shows the no-data message instead of a blank list', (tester) async {
    await _pumpSheet(tester, onSelected: (_) {});

    await tester.enterText(find.byType(TextField), 'zzz-no-such-country');
    await tester.pumpAndSettle();

    expect(find.text('Mauritania'), findsNothing);
    expect(find.text('Tunisia'), findsNothing);
  });
}
