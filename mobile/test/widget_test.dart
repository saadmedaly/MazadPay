import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/main.dart';

void main() {
  testWidgets('App starts with splash screen', (WidgetTester tester) async {
    // Build our app and trigger a frame.
    await tester.pumpWidget(const MazadApp());

    // Verify that SplashPage is shown (checking for MazadPay branding)
    expect(find.text('MazadPay'), findsOneWidget);
  });
}
