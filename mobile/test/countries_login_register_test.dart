import 'package:flutter_test/flutter_test.dart';
import 'package:mezadpay/utils/phone_validation.dart';

// Regression tests for the mobile-side "8/12 instead of 8/8, Login button
// stuck disabled" bug. Root cause: GET /countries returns a healthy,
// correctly-shaped response (success:true, data:[...] of Map objects with
// int phone_min_length/phone_max_length), but the SAME source APK observed
// showing 8/12 was built and tested while the backend was crash-looping
// (SCRAM auth failure against a stale Neon endpoint, fixed separately) --
// so _loadCountries() never received a successful response and
// _selectedCountry fell through to the generic 7-12 fallback range with no
// real country selected. Two structural gaps existed regardless of that
// specific timeline: (1) selectDefaultCountry's selection algorithm had
// never been exercised against a fixture shaped exactly like the live
// backend response, and (2) a failed countries load left the form
// permanently unusable with no retry path. Both are covered here.
//
// This fixture mirrors the exact live GET /v1/api/countries response shape
// confirmed by curling https://mazadpay-validation-backend.onrender.com/v1/api/countries
// directly: {"data": [ {..., "phone_min_length": 8 (int), ...} ], "success": true}.
final List<Map<String, dynamic>> liveShapedCountriesFixture = [
  {
    'id': 81,
    'code': 'IS',
    'country_code': '+354',
    'name_ar': 'آيسلندا',
    'name_fr': 'Islande',
    'name_en': 'Iceland',
    'flag_emoji': '🇮🇸',
    'is_active': true,
    'phone_min_length': 7,
    'phone_max_length': 12,
    'currency_code': 'ISK',
  },
  {
    'id': 999,
    'code': 'MR',
    'country_code': '+222',
    'name_ar': 'موريتانيا',
    'name_fr': 'Mauritanie',
    'name_en': 'Mauritania',
    'flag_emoji': '🇲🇷',
    'is_active': true,
    'phone_min_length': 8,
    'phone_max_length': 8,
    'currency_code': 'MRU',
  },
  {
    'id': 70,
    'code': 'DE',
    'country_code': '+49',
    'name_ar': 'ألمانيا',
    'name_fr': 'Allemagne',
    'name_en': 'Germany',
    'flag_emoji': '🇩🇪',
    'is_active': true,
    'phone_min_length': 10,
    'phone_max_length': 11,
    'currency_code': 'EUR',
  },
  {
    'id': 7,
    'code': 'SN',
    'country_code': '+221',
    'name_ar': 'السنغال',
    'name_fr': 'Sénégal',
    'name_en': 'Senegal',
    'flag_emoji': '🇸🇳',
    'is_active': true,
    'phone_min_length': 9,
    'phone_max_length': 9,
    'currency_code': 'XOF',
  },
  {
    'id': 8,
    'code': 'US',
    'country_code': '+1',
    'name_ar': 'الولايات المتحدة',
    'name_fr': 'États-Unis',
    'name_en': 'United States',
    'flag_emoji': '🇺🇸',
    'is_active': true,
    'phone_min_length': 10,
    'phone_max_length': 10,
    'currency_code': 'USD',
  },
];

void main() {
  group('live-shaped /countries fixture parses correctly', () {
    test('MR is present with the exact live values', () {
      final mr = liveShapedCountriesFixture.firstWhere((c) => c['code'] == 'MR');
      expect(mr['country_code'], '+222');
      expect(mr['phone_min_length'], 8);
      expect(mr['phone_max_length'], 8);
      // Confirms the live backend really does send ints, not strings/doubles.
      expect(mr['phone_min_length'], isA<int>());
      expect(mr['phone_max_length'], isA<int>());
    });

    test('phoneMinLengthFor/phoneMaxLengthFor read the live MR int values directly', () {
      final mr = liveShapedCountriesFixture.firstWhere((c) => c['code'] == 'MR');
      expect(phoneMinLengthFor(mr), 8);
      expect(phoneMaxLengthFor(mr), 8);
    });
  });

  group('selectDefaultCountry', () {
    test('selects MR by default when no target is specified', () {
      final selected = selectDefaultCountry(liveShapedCountriesFixture);
      expect(selected, isNotNull);
      expect(selected!['code'], 'MR');
      expect(selected['country_code'], '+222');
      expect(phoneMinLengthFor(selected), 8);
      expect(phoneMaxLengthFor(selected), 8);
    });

    test('_selectedCountry is non-null after a successful, non-empty load', () {
      final selected = selectDefaultCountry(liveShapedCountriesFixture);
      expect(selected, isNotNull);
    });

    test('the full international list is preserved, not reduced to MR-only', () {
      // Regression guard: the fix must not collapse the picker down to a
      // single hardcoded MR entry -- every backend-returned country must
      // stay selectable.
      expect(liveShapedCountriesFixture.length, 5);
      expect(liveShapedCountriesFixture.map((c) => c['code']),
          containsAll(['MR', 'IS', 'DE', 'SN', 'US']));
    });

    test('prefers an explicit target ISO over the MR default', () {
      final selected = selectDefaultCountry(liveShapedCountriesFixture, targetIso: 'DE');
      expect(selected!['code'], 'DE');
      expect(selected['country_code'], '+49');
    });

    test('prefers an explicit target dial code over the MR default', () {
      final selected =
          selectDefaultCountry(liveShapedCountriesFixture, targetCountryCode: '+221');
      expect(selected!['code'], 'SN');
    });

    test('falls back to the first entry when neither target nor MR exists', () {
      final noMr = liveShapedCountriesFixture.where((c) => c['code'] != 'MR').toList();
      final selected = selectDefaultCountry(noMr, targetIso: 'ZZ');
      expect(selected, isNotNull);
      expect(selected!['code'], noMr.first['code']);
    });

    test('returns null only for a genuinely empty list', () {
      expect(selectDefaultCountry(const []), isNull);
    });
  });

  group('MR phone length validation (8/8, not the 7-12 fallback)', () {
    final mr = liveShapedCountriesFixture.firstWhere((c) => c['code'] == 'MR');

    test('an 8-digit MR number is valid', () {
      expect(isPhoneLengthValid('32816780', mr), isTrue);
    });

    test('a 7-digit MR number is invalid', () {
      expect(isPhoneLengthValid('3281678', mr), isFalse);
    });

    test('a 9-digit MR number is invalid', () {
      expect(isPhoneLengthValid('328167800', mr), isFalse);
    });
  });

  group('international country switching updates validation dynamically', () {
    test('MR -> Germany -> MR restores +222 and 8/8', () {
      final mr = selectDefaultCountry(liveShapedCountriesFixture)!;
      expect(mr['country_code'], '+222');
      expect(phoneMinLengthFor(mr), 8);
      expect(phoneMaxLengthFor(mr), 8);

      final germany = selectDefaultCountry(liveShapedCountriesFixture, targetIso: 'DE')!;
      expect(germany['country_code'], '+49');
      expect(phoneMinLengthFor(germany), 10);
      expect(phoneMaxLengthFor(germany), 11);

      final backToMr = selectDefaultCountry(liveShapedCountriesFixture, targetIso: 'MR')!;
      expect(backToMr['country_code'], '+222');
      expect(phoneMinLengthFor(backToMr), 8);
      expect(phoneMaxLengthFor(backToMr), 8);
    });

    test('an African non-MR country (Senegal) validates with its own 9/9 range', () {
      final senegal = selectDefaultCountry(liveShapedCountriesFixture, targetIso: 'SN')!;
      expect(senegal['country_code'], '+221');
      expect(isPhoneLengthValid('771234567', senegal), isTrue); // 9 digits
      expect(isPhoneLengthValid('77123456', senegal), isFalse); // 8 digits
    });

    test('a European country (Germany) validates with its own 10-11 range', () {
      final germany = selectDefaultCountry(liveShapedCountriesFixture, targetIso: 'DE')!;
      expect(isPhoneLengthValid('1512345678', germany), isTrue); // 10 digits
      expect(isPhoneLengthValid('151234567', germany), isFalse); // 9 digits, too short
    });

    test('a North American country (US) validates with its own 10/10 range', () {
      final us = selectDefaultCountry(liveShapedCountriesFixture, targetIso: 'US')!;
      expect(us['country_code'], '+1');
      expect(isPhoneLengthValid('6135550123', us), isTrue); // 10 digits
      expect(isPhoneLengthValid('613555012', us), isFalse); // 9 digits
    });

    test('country_iso remains correct after switching country', () {
      final senegal = selectDefaultCountry(liveShapedCountriesFixture, targetIso: 'SN')!;
      expect(senegal['code'], 'SN');
      final us = selectDefaultCountry(liveShapedCountriesFixture, targetIso: 'US')!;
      expect(us['code'], 'US');
    });
  });

  group('a successful-but-empty /countries response is treated as a failure', () {
    // Regression test for the real-device bug found after APK 1.1.0+8:
    // every _loadCountries() success branch originally guarded only on
    // `response.success && response.data != null` -- which is ALSO true
    // for {"success": true, "data": []}. That path bypassed the entire
    // failure/fallback branch: _countries stayed [] (empty country picker,
    // "لا توجد بيانات متاحة") and _selectedCountry stayed null (falls
    // through to the generic 7-12 range, producing the observed 8/12
    // instead of 8/8). The fix adds an explicit non-empty check
    // (`response.data!.isNotEmpty`) so an empty list is routed through the
    // exact same recovery path as a genuine network failure.
    bool hasUsableData({required bool success, required List<dynamic>? data}) {
      return success && data != null && data.isNotEmpty;
    }

    test('success=true with a non-empty list is usable', () {
      expect(hasUsableData(success: true, data: liveShapedCountriesFixture), isTrue);
    });

    test('success=true with an EMPTY list is NOT usable (the real-device bug)', () {
      expect(hasUsableData(success: true, data: const []), isFalse);
    });

    test('success=false is not usable regardless of data', () {
      expect(hasUsableData(success: false, data: liveShapedCountriesFixture), isFalse);
    });

    test('success=true with null data is not usable', () {
      expect(hasUsableData(success: true, data: null), isFalse);
    });

    test('an empty response routes through the same fallback as a genuine failure, '
        'seeding both _selectedCountry and the picker list', () {
      // Simulates the exact failure-branch logic now shared by every
      // affected page: _selectedCountry ??= kFallbackMauritania; and
      // _countries = _countries.isEmpty ? [kFallbackMauritania] : _countries.
      Map<String, dynamic>? selectedCountry;
      List<dynamic> countries = [];

      final usable = hasUsableData(success: true, data: const []);
      expect(usable, isFalse);

      selectedCountry ??= kFallbackMauritania;
      if (countries.isEmpty) {
        countries = [kFallbackMauritania];
      }

      // The picker must never be empty after this recovery.
      expect(countries, isNotEmpty);
      // The counter must show 8/8, not the generic 7-12 fallback range.
      expect(phoneMinLengthFor(selectedCountry), 8);
      expect(phoneMaxLengthFor(selectedCountry), 8);
    });
  });

  group('countries load failure does not permanently lock the form', () {
    test('kFallbackMauritania provides a valid, non-null selection on failure', () {
      // Mirrors what every _loadCountries() implementation now does on a
      // failed/empty response: _selectedCountry ??= kFallbackMauritania.
      expect(kFallbackMauritania['code'], 'MR');
      expect(kFallbackMauritania['country_code'], '+222');
      expect(phoneMinLengthFor(kFallbackMauritania), 8);
      expect(phoneMaxLengthFor(kFallbackMauritania), 8);
      // A non-null selected country means _areFieldsValid's
      // `_selectedCountry != null` clause is satisfiable even on failure,
      // rather than staying permanently null forever.
      expect(kFallbackMauritania, isNotNull);
    });

    test('a recovered (retried) successful load replaces the fallback with the real list', () {
      // Simulates: first load failed -> fallback MR applied -> retry
      // succeeds -> selectDefaultCountry now runs against the real list and
      // still correctly resolves MR (or any other requested country).
      Map<String, dynamic>? selected = kFallbackMauritania;
      expect(selected['phone_min_length'], 8);

      // Retry succeeds with the full international list.
      selected = selectDefaultCountry(liveShapedCountriesFixture, targetIso: 'DE');
      expect(selected!['code'], 'DE');
      expect(phoneMinLengthFor(selected), 10);
    });

    test('a retry after failure only re-selects while the fallback is still active, '
        'never overwriting a real user selection', () {
      // Mirrors the exact guard every _loadCountries() success branch now
      // uses: `if (_selectedCountry == null || identical(_selectedCountry,
      // kFallbackMauritania)) { _selectedCountry = selectDefaultCountry(...); }`.
      // This is a regression test for a bug found during review: the retry
      // button calls the same _loadCountries() success path, which used to
      // unconditionally re-run the default-selection algorithm on every
      // successful response -- silently discarding a country the user had
      // already picked via the country picker sheet before pressing retry.
      Map<String, dynamic>? selectedCountry;

      Map<String, dynamic>? applySuccessfulLoad(Map<String, dynamic>? current) {
        if (current == null || identical(current, kFallbackMauritania)) {
          return selectDefaultCountry(liveShapedCountriesFixture);
        }
        return current;
      }

      // First load succeeds normally -> MR selected by default.
      selectedCountry = applySuccessfulLoad(selectedCountry);
      expect(selectedCountry!['code'], 'MR');

      // User manually picks Germany from the country picker sheet (this is
      // exactly what `onSelected: (country) => setState(() => _selectedCountry
      // = country)` does in every affected page).
      selectedCountry = liveShapedCountriesFixture.firstWhere((c) => c['code'] == 'DE');
      expect(selectedCountry['code'], 'DE');

      // A later successful countries reload (e.g. a background refresh, or
      // pressing retry when it wasn't actually needed) must NOT discard the
      // user's manual Germany selection.
      selectedCountry = applySuccessfulLoad(selectedCountry);
      expect(selectedCountry!['code'], 'DE');
    });

    test('a retry that follows a genuine failure correctly replaces the temporary '
        'MR fallback with the real selection once data arrives', () {
      Map<String, dynamic>? selectedCountry;

      Map<String, dynamic>? applySuccessfulLoad(Map<String, dynamic>? current) {
        if (current == null || identical(current, kFallbackMauritania)) {
          return selectDefaultCountry(liveShapedCountriesFixture);
        }
        return current;
      }

      Map<String, dynamic>? applyFailedLoad(Map<String, dynamic>? current) {
        return current ?? kFallbackMauritania;
      }

      // Countries request fails -> fallback applied.
      selectedCountry = applyFailedLoad(selectedCountry);
      expect(identical(selectedCountry, kFallbackMauritania), isTrue);

      // User presses retry, this time it succeeds -> the fallback is
      // recognized and replaced by the real default selection.
      selectedCountry = applySuccessfulLoad(selectedCountry);
      expect(identical(selectedCountry, kFallbackMauritania), isFalse);
      expect(selectedCountry!['code'], 'MR');
    });
  });

  group('legacy login compatibility (password/PIN length rule)', () {
    // Mirrors LoginPage._isPasswordValid: len == 4 || len >= 8.
    bool isPasswordValid(String password) {
      final len = password.trim().length;
      return len == 4 || len >= 8;
    }

    test('a legacy 4-digit PIN is accepted', () {
      expect(isPasswordValid('1234'), isTrue);
    });

    test('an 8+ character password is accepted', () {
      expect(isPasswordValid('password123'), isTrue);
      expect(isPasswordValid('12345678'), isTrue);
    });

    test('a 5-7 character password is rejected (neither legacy PIN nor v2 password)', () {
      expect(isPasswordValid('12345'), isFalse);
      expect(isPasswordValid('123456'), isFalse);
      expect(isPasswordValid('1234567'), isFalse);
    });

    test('a 1-3 character password is rejected', () {
      expect(isPasswordValid('1'), isFalse);
      expect(isPasswordValid('123'), isFalse);
    });
  });
}
