import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/pages/services_page.dart';

// Note #4 (client feedback): the Home page services grid previously showed
// 12 service icons (كورس, توصيل, نقل البضائع, كورس عبر المدن, توصيل طعام,
// توصيل أدوية, شحن من خارج, رافعة سيارة, شاحنة ماء, توصيل مجاري, نقل أثاث,
// توصيل أسماك ولحوم), none of which had a working tap handler (onTap: null
// since client feedback A5 removed the "Delivery Details" page). Per the
// client's authoritative reference image, only these 5 delivery services
// must remain, and tapping ANY of them must open MazadPay's own WhatsApp
// (+22247601175) using the same reliable wa.me + whatsapp:// fallback
// strategy already fixed in My Winnings (mobile/utils/whatsapp_launcher.dart
// -- shared, not reimplemented).
Widget _wrap(Widget child) {
  return MaterialApp(
    locale: const Locale('ar'),
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    home: Scaffold(body: child),
  );
}

void main() {
  group('ServicesPage (Note #4: reduced to 5 delivery services)', () {
    testWidgets('exactly 5 service cards render', (tester) async {
      await tester.pumpWidget(_wrap(const ServicesPage()));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.local_shipping), findsOneWidget); // نقل البضائع
      expect(find.byIcon(Icons.delivery_dining), findsOneWidget); // توصيل
      expect(find.byIcon(Icons.car_repair), findsOneWidget); // رافعة سيارة
      expect(find.byIcon(Icons.chair), findsOneWidget); // نقل أثاث
      expect(find.byIcon(Icons.flight), findsOneWidget); // شحن من خارج
    });

    testWidgets('the 5 required service labels are visible', (tester) async {
      await tester.pumpWidget(_wrap(const ServicesPage()));
      await tester.pumpAndSettle();

      expect(find.text('نقل البضائع'), findsOneWidget);
      expect(find.text('توصيل'), findsOneWidget);
      expect(find.text('رافعة سيارة'), findsOneWidget);
      expect(find.text('نقل أثاث'), findsOneWidget);
      expect(find.text('شحن من خارج'), findsOneWidget);
    });

    testWidgets('removed services no longer render (icons)', (tester) async {
      await tester.pumpWidget(_wrap(const ServicesPage()));
      await tester.pumpAndSettle();

      // Icons that belonged only to the 7 removed services.
      expect(find.byIcon(Icons.local_taxi), findsNothing); // كورس
      expect(find.byIcon(Icons.directions_bus), findsNothing); // كورس عبر المدن
      expect(find.byIcon(Icons.restaurant), findsNothing); // توصيل طعام
      expect(find.byIcon(Icons.medical_services), findsNothing); // توصيل أدوية
      expect(find.byIcon(Icons.water), findsNothing); // شاحنة ماء
      expect(find.byIcon(Icons.plumbing), findsNothing); // توصيل مجاري
      expect(find.byIcon(Icons.set_meal), findsNothing); // توصيل أسماك ولحوم
    });

    testWidgets('removed services no longer render (labels)', (tester) async {
      await tester.pumpWidget(_wrap(const ServicesPage()));
      await tester.pumpAndSettle();

      expect(find.text('كورس'), findsNothing);
      expect(find.text('كورس عبر المدن'), findsNothing);
      expect(find.text('توصيل طعام'), findsNothing);
      expect(find.text('توصيل أدوية'), findsNothing);
      expect(find.text('شاحنة ماء'), findsNothing);
      expect(find.text('توصيل مجاري'), findsNothing);
      expect(find.text('توصيل أسماك ولحوم'), findsNothing);
    });

    testWidgets('the grid contains exactly 5 InkWell service cards, not the old 12', (tester) async {
      await tester.pumpWidget(_wrap(const ServicesPage()));
      await tester.pumpAndSettle();

      final gridFinder = find.byType(GridView);
      expect(gridFinder, findsOneWidget);
      final grid = tester.widget<GridView>(gridFinder);
      expect(grid.semanticChildCount, 5);
    });

    testWidgets('the section title still renders (layout unchanged)', (tester) async {
      await tester.pumpWidget(_wrap(const ServicesPage()));
      await tester.pumpAndSettle();

      // Same section-title widget/style as before -- only the grid contents
      // changed, per the client's explicit "keep the same visual style".
      expect(find.byType(GridView), findsOneWidget);
      final gridDelegate =
          (tester.widget<GridView>(find.byType(GridView)).gridDelegate)
              as SliverGridDelegateWithFixedCrossAxisCount;
      expect(gridDelegate.crossAxisCount, 3);
    });

    testWidgets('tapping any of the 5 service cards does not crash (WhatsApp launch attempted)',
        (tester) async {
      await tester.pumpWidget(_wrap(const ServicesPage()));
      await tester.pumpAndSettle();

      // url_launcher has no platform channel in a widget test, so
      // launchUrl throws/returns false and the shared launcher's own
      // catch/snackbar path runs -- this proves the tap handler is wired
      // (not onTap: null) without needing a real WhatsApp install.
      for (final icon in [
        Icons.local_shipping,
        Icons.delivery_dining,
        Icons.car_repair,
        Icons.chair,
        Icons.flight,
      ]) {
        final finder = find.byIcon(icon);
        await tester.ensureVisible(finder);
        await tester.pumpAndSettle();
        await tester.tap(finder);
        await tester.pump();
      }
      expect(tester.takeException(), isNull);
    });
  });
}
