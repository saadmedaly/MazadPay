import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/utils/description_validation.dart';

void main() {
  group('description length limits (10-5000 chars)', () {
    test('too short below the minimum', () {
      expect(isDescriptionTooShort('short'), isTrue); // 5 chars
      expect(isDescriptionTooShort(''), isTrue);
    });

    test('not too short at or above the minimum', () {
      expect(isDescriptionTooShort('1234567890'), isFalse); // exactly 10
      expect(isDescriptionTooShort('a valid description here'), isFalse);
    });

    test('trims whitespace before measuring the minimum', () {
      expect(isDescriptionTooShort('   short   '), isTrue);
      expect(isDescriptionTooShort('  1234567890  '), isFalse);
    });

    test('too long above the maximum', () {
      final tooLong = 'a' * (kDescriptionMaxLength + 1);
      expect(isDescriptionTooLong(tooLong), isTrue);
    });

    test('not too long at or below the maximum', () {
      final atMax = 'a' * kDescriptionMaxLength;
      expect(isDescriptionTooLong(atMax), isFalse);
    });

    test('isDescriptionValid requires both bounds satisfied', () {
      expect(isDescriptionValid('a' * 10), isTrue);
      expect(isDescriptionValid('a' * 9), isFalse);
      expect(isDescriptionValid('a' * (kDescriptionMaxLength + 1)), isFalse);
      expect(isDescriptionValid('a' * kDescriptionMaxLength), isTrue);
    });
  });
}
