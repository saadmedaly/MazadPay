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
