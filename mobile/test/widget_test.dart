import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/widgets/mazad_pay_logo.dart';

void main() {
  testWidgets('MazadPay logo renders correctly', (WidgetTester tester) async {
    // Build the logo widget in isolation, without MazadApp startup side effects.
    await tester.pumpWidget(
      MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        home: const Scaffold(
          body: MazadPayLogo(),
        ),
      ),
    );

    expect(find.byType(MazadPayLogo), findsOneWidget);
  });
}
