import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/utils/phone_validation.dart';

void main() {
  group('phoneMinLengthFor / phoneMaxLengthFor', () {
    test('reads phone_min_length/phone_max_length from the country map', () {
      final country = {'phone_min_length': 9, 'phone_max_length': 10};
      expect(phoneMinLengthFor(country), 9);
      expect(phoneMaxLengthFor(country), 10);
    });

    test('falls back to the default range when fields are null', () {
      expect(phoneMinLengthFor(null), kDefaultPhoneMinLength);
      expect(phoneMaxLengthFor(null), kDefaultPhoneMaxLength);
      expect(phoneMinLengthFor(<String, dynamic>{}), kDefaultPhoneMinLength);
      expect(phoneMaxLengthFor(<String, dynamic>{}), kDefaultPhoneMaxLength);
    });

    test('accepts numeric (non-int) values from JSON decoding', () {
      final country = {'phone_min_length': 8.0, 'phone_max_length': 12.0};
      expect(phoneMinLengthFor(country), 8);
      expect(phoneMaxLengthFor(country), 12);
    });
  });

  group('isPhoneLengthValid', () {
    final mauritania = {'phone_min_length': 8, 'phone_max_length': 8};
    final usCanada = {'phone_min_length': 10, 'phone_max_length': 10};

    test('accepts a phone within the country range', () {
      expect(isPhoneLengthValid('20123456', mauritania), isTrue);
      expect(isPhoneLengthValid('6135550123', usCanada), isTrue);
    });

    test('rejects a phone shorter than the minimum', () {
      expect(isPhoneLengthValid('2012345', mauritania), isFalse);
    });

    test('rejects a phone longer than the maximum', () {
      expect(isPhoneLengthValid('201234567', mauritania), isFalse);
    });

    test('uses the generic 7-12 digit fallback when country is null', () {
      expect(isPhoneLengthValid('1234567', null), isTrue); // 7 digits, min
      expect(isPhoneLengthValid('123456789012', null), isTrue); // 12, max
      expect(isPhoneLengthValid('123456', null), isFalse); // 6, too short
      expect(isPhoneLengthValid('1234567890123', null), isFalse); // 13, too long
    });
  });

  group('filterUsableCountries', () {
    test('keeps only entries with a non-null country_code', () {
      final countries = [
        {'code': 'MR', 'country_code': '+222'},
        {'code': 'ZZ'}, // missing country_code, should be dropped
        {'code': 'TN', 'country_code': '+216'},
        'not a map', // non-map entries should be dropped too
      ];
      final usable = filterUsableCountries(countries);
      expect(usable.length, 2);
      expect(usable.map((c) => c['code']), containsAll(['MR', 'TN']));
    });
  });

  group('countryMatchesQuery', () {
    final tunisia = {
      'name_ar': 'تونس',
      'name_fr': 'Tunisie',
      'name_en': 'Tunisia',
      'country_code': '+216',
      'code': 'TN',
    };

    test('empty query matches everything', () {
      expect(countryMatchesQuery(tunisia, ''), isTrue);
      expect(countryMatchesQuery(tunisia, '   '), isTrue);
    });

    test('matches case-insensitively on the English name', () {
      expect(countryMatchesQuery(tunisia, 'tun'), isTrue);
      expect(countryMatchesQuery(tunisia, 'TUN'), isTrue);
    });

    test('matches on dial code and ISO code', () {
      expect(countryMatchesQuery(tunisia, '+216'), isTrue);
      expect(countryMatchesQuery(tunisia, 'TN'), isTrue);
    });

    test('does not match an unrelated query', () {
      expect(countryMatchesQuery(tunisia, 'xyz'), isFalse);
    });
  });
}
