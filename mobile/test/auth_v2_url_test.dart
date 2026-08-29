import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/services/auth_api.dart';

void main() {
  group('buildV2AuthUrl (API Versioning Phase 1)', () {
    test('replaces /v1/api with /v2/api in a typical base URL', () {
      final url = buildV2AuthUrl(
        'http://localhost:8082/v1/api',
        '/auth/register',
      );
      expect(url, 'http://localhost:8082/v2/api/auth/register');
    });

    test('works for the reset-password path too', () {
      final url = buildV2AuthUrl(
        'https://api.mazadpay.example/v1/api',
        '/auth/reset-password',
      );
      expect(url, 'https://api.mazadpay.example/v2/api/auth/reset-password');
    });

    test('never produces a /v1/api URL for the v2-routed paths', () {
      final url = buildV2AuthUrl(
        'http://localhost:8082/v1/api',
        '/auth/register',
      );
      expect(
        url.contains('/v1/api'),
        isFalse,
        reason:
            'a v2-routed call must never resolve back to the legacy /v1/api path',
      );
    });

    test(
      'falls back to appending /v2/api when the base has no /v1/api segment',
      () {
        final url = buildV2AuthUrl('http://localhost:8082', '/auth/register');
        expect(url, 'http://localhost:8082/v2/api/auth/register');
      },
    );
  });
}
