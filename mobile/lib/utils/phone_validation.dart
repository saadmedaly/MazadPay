/// Generic, country-agnostic phone helpers.
///
/// The backend now validates phone numbers server-side (E.164 / libphonenumber)
/// for ALL countries, not just Mauritania. These helpers only provide a soft
/// client-side hint using the `phone_min_length`/`phone_max_length` fields
/// returned by `GET /countries` (when present), falling back to a generic
/// 7-12 digit range when they are absent. The server remains authoritative.
library;

/// Default fallback range (national number digit length) used when the
/// selected country doesn't carry `phone_min_length`/`phone_max_length`.
const int kDefaultPhoneMinLength = 7;
const int kDefaultPhoneMaxLength = 12;

/// Local Mauritania fallback, used ONLY when `GET /countries` genuinely
/// fails (network error, timeout, malformed response) so the login/register
/// forms don't get stuck showing the generic 7-12 range with no selected
/// country at all. Matches the real values the backend returns for MR
/// (confirmed via GET /countries: code=MR, country_code=+222,
/// phone_min_length=8, phone_max_length=8) -- this is a display/validation
/// fallback only, never used to bypass server-side validation, and never
/// substituted when the real country list loads successfully (which always
/// takes priority and includes every other country normally).
final Map<String, dynamic> kFallbackMauritania = {
  'code': 'MR',
  'country_code': '+222',
  'phone_min_length': 8,
  'phone_max_length': 8,
};

int phoneMinLengthFor(Map<String, dynamic>? country) {
  final val = country?['phone_min_length'];
  if (val is int) return val;
  if (val is num) return val.toInt();
  return kDefaultPhoneMinLength;
}

int phoneMaxLengthFor(Map<String, dynamic>? country) {
  final val = country?['phone_max_length'];
  if (val is int) return val;
  if (val is num) return val.toInt();
  return kDefaultPhoneMaxLength;
}

/// Returns true when [phone] (digits only, no dial code) has a length within
/// the selected country's expected national-number length range.
bool isPhoneLengthValid(String phone, Map<String, dynamic>? country) {
  final len = phone.trim().length;
  final min = phoneMinLengthFor(country);
  final max = phoneMaxLengthFor(country);
  return len >= min && len <= max;
}

/// Filters a country list (from `GET /countries`) to entries usable in a
/// picker: must have a non-null `country_code`.
List<dynamic> filterUsableCountries(List<dynamic> countries) {
  return countries.where((c) => c is Map && c['country_code'] != null).toList();
}

/// Pure selection logic shared by every page that loads `GET /countries` and
/// needs a sensible default selected country (Login, Register, forgot-
/// password). Priority: an explicit target ISO code, then an explicit target
/// dial code (both used when arriving from a prior screen that already
/// picked a country), then Mauritania (this app's default market), then
/// simply the first entry in the list. Returns null only when [countries] is
/// empty. Extracted as a pure function (no BuildContext/setState) so this
/// exact selection algorithm is unit-testable without a widget harness.
Map<String, dynamic>? selectDefaultCountry(
  List<dynamic> countries, {
  String? targetIso,
  String? targetCountryCode,
}) {
  if (targetIso != null) {
    for (final c in countries) {
      if (c is Map && c['code'] == targetIso) return c.cast<String, dynamic>();
    }
  }
  if (targetCountryCode != null) {
    for (final c in countries) {
      if (c is Map && c['country_code'] == targetCountryCode) {
        return c.cast<String, dynamic>();
      }
    }
  }
  for (final c in countries) {
    if (c is Map && (c['country_code'] == '+222' || c['code'] == 'MR')) {
      return c.cast<String, dynamic>();
    }
  }
  if (countries.isNotEmpty && countries.first is Map) {
    return (countries.first as Map).cast<String, dynamic>();
  }
  return null;
}

/// Case-insensitive substring match across the localized names and dial code,
/// used by the searchable country picker.
bool countryMatchesQuery(Map<String, dynamic> country, String query) {
  if (query.trim().isEmpty) return true;
  final q = query.trim().toLowerCase();
  final candidates = [
    country['name_ar'],
    country['name_fr'],
    country['name_en'],
    country['country_code'],
    country['code'],
  ];
  return candidates.any(
    (c) => c != null && c.toString().toLowerCase().contains(q),
  );
}
