import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/utils/whatsapp_launcher.dart';

// Note #4 (client feedback): shared WhatsApp-launch helper used by
// services_page.dart's 5 delivery-service cards. Reuses the exact number
// and dual-launch (wa.me primary + whatsapp:// native fallback) strategy
// already established/fixed in my_winnings_page.dart's
// buildWinnerPaymentWhatsAppUri/buildWinnerPaymentWhatsAppNativeUri -- these
// tests mirror that file's own coverage of the pure URI builders.
void main() {
  group('WhatsApp number (Note #4: client-supplied +22247601175)', () {
    test('kMazadPayWhatsAppNumber matches the client-supplied number without the 222 country code', () {
      // Client supplied +22247601175 -- 222 is the Mauritania country code,
      // 47601175 is the local number these builders prepend "222" to.
      expect(kMazadPayWhatsAppNumber, '47601175');
      expect('222$kMazadPayWhatsAppNumber', '22247601175');
    });
  });

  group('buildMazadPayWhatsAppUri (primary wa.me launch)', () {
    test('builds a wa.me link with the full international number', () {
      final uri = buildMazadPayWhatsAppUri('test message');
      expect(uri.scheme, 'https');
      expect(uri.host, 'wa.me');
      expect(uri.path, '/22247601175');
    });

    test('includes the prefilled message as the text query parameter', () {
      final uri = buildMazadPayWhatsAppUri('مرحباً، أرغب في الاستفسار عن خدمة "توصيل".');
      expect(uri.queryParameters['text'], 'مرحباً، أرغب في الاستفسار عن خدمة "توصيل".');
    });
  });

  group('buildMazadPayWhatsAppNativeUri (whatsapp:// fallback)', () {
    test('builds a native whatsapp:// deep link with the full international number', () {
      final uri = buildMazadPayWhatsAppNativeUri('test message');
      expect(uri.scheme, 'whatsapp');
      expect(uri.host, 'send');
      expect(uri.queryParameters['phone'], '22247601175');
    });

    test('includes the prefilled message as the text query parameter', () {
      final uri = buildMazadPayWhatsAppNativeUri('مرحباً، أرغب في الاستفسار عن خدمة "نقل أثاث".');
      expect(uri.queryParameters['text'], 'مرحباً، أرغب في الاستفسار عن خدمة "نقل أثاث".');
    });
  });
}
