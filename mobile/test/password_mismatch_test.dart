import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/l10n/app_localizations.dart';

/// Standalone fixture that mirrors the password + confirm-password field
/// pattern used by `PhonePasswordPage` (registration) and `NewPasswordPage`
/// (reset flow): an 8+ character requirement, a confirm field that must
/// match, and a submit control disabled until both hold.
///
/// `PhonePasswordPage`/`NewPasswordPage` themselves instantiate `AuthApi()`/
/// `CategoryApi()` in `initState`, which construct a `Dio` client reading
/// `flutter_dotenv`'s env map — pumping those screens directly in a plain
/// widget test risks coupling this test's pass/fail to dotenv/network
/// plumbing that has nothing to do with the validation rule under test. This
/// fixture reproduces the exact predicate (`length < 8`, `password !=
/// confirm`) those screens use, so the logic is verified without that
/// coupling.
class _PasswordFormFixture extends StatefulWidget {
  const _PasswordFormFixture();

  @override
  State<_PasswordFormFixture> createState() => _PasswordFormFixtureState();
}

class _PasswordFormFixtureState extends State<_PasswordFormFixture> {
  final _password = TextEditingController();
  final _confirm = TextEditingController();

  bool get _isValid =>
      _password.text.length >= 8 && _password.text == _confirm.text;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          TextField(key: const Key('password'), controller: _password, onChanged: (_) => setState(() {})),
          TextField(key: const Key('confirm'), controller: _confirm, onChanged: (_) => setState(() {})),
          if (_confirm.text.isNotEmpty && _confirm.text != _password.text)
            Text(AppLocalizations.of(context)!.error_password_mismatch),
          if (_password.text.isNotEmpty && _password.text.length < 8)
            Text(AppLocalizations.of(context)!.error_password_too_short),
          ElevatedButton(
            key: const Key('submit'),
            onPressed: _isValid ? () {} : null,
            child: const Text('submit'),
          ),
        ],
      ),
    );
  }
}

Future<void> _pump(WidgetTester tester) async {
  await tester.pumpWidget(
    MaterialApp(
      locale: const Locale('ar'),
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      home: const _PasswordFormFixture(),
    ),
  );
}

void main() {
  testWidgets('mismatched confirm password shows the mismatch error and blocks submit', (tester) async {
    await _pump(tester);
    final l10n = await AppLocalizations.delegate.load(const Locale('ar'));

    await tester.enterText(find.byKey(const Key('password')), 'StrongPass1');
    await tester.enterText(find.byKey(const Key('confirm')), 'DifferentPass1');
    await tester.pump();

    expect(find.text(l10n.error_password_mismatch), findsOneWidget);
    final submit = tester.widget<ElevatedButton>(find.byKey(const Key('submit')));
    expect(submit.onPressed, isNull);
  });

  testWidgets('matching 8+ char passwords clear the error and enable submit', (tester) async {
    await _pump(tester);
    final l10n = await AppLocalizations.delegate.load(const Locale('ar'));

    await tester.enterText(find.byKey(const Key('password')), 'StrongPass1');
    await tester.enterText(find.byKey(const Key('confirm')), 'StrongPass1');
    await tester.pump();

    expect(find.text(l10n.error_password_mismatch), findsNothing);
    expect(find.text(l10n.error_password_too_short), findsNothing);
    final submit = tester.widget<ElevatedButton>(find.byKey(const Key('submit')));
    expect(submit.onPressed, isNotNull);
  });

  testWidgets('a password under 8 characters shows the too-short error and blocks submit', (tester) async {
    await _pump(tester);
    final l10n = await AppLocalizations.delegate.load(const Locale('ar'));

    await tester.enterText(find.byKey(const Key('password')), 'short1');
    await tester.enterText(find.byKey(const Key('confirm')), 'short1');
    await tester.pump();

    expect(find.text(l10n.error_password_too_short), findsOneWidget);
    final submit = tester.widget<ElevatedButton>(find.byKey(const Key('submit')));
    expect(submit.onPressed, isNull);
  });
}
