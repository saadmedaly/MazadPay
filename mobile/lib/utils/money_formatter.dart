/// Shared money/currency formatter for MazadPay Phase 2 (country-scoped currency).
///
/// Displays an amount with its ISO-4217 currency code (e.g. "100 MRU",
/// "100 TND", "100 MAD", "100 XOF") -- NO exchange-rate conversion, NO
/// locale-derived currency guessing. The currency always comes from the
/// API object's own `currency_code` (or `EffectiveCurrencyCode()` fallback
/// server-side); this formatter only renders what it is given.
///
/// Mirrors backend `models.DefaultCurrencyCode` ("MRU") -- used ONLY as a
/// last-resort fallback when an API object legitimately has no
/// currency_code at all (very old cached/legacy responses), never as a
/// general-purpose default for new data.

class MoneyFormatter {
  MoneyFormatter._();

  /// Fallback currency when an object has no currency_code at all (legacy
  /// API responses predating migration 000046). Matches backend
  /// DefaultCurrencyCode in backend/internal/models/user.go.
  static const String fallbackCurrencyCode = 'MRU';

  /// Conventional minor-unit digit counts per ISO-4217 code, mirroring
  /// backend `currencies.minor_units` (migration 000046). Used only for
  /// display rounding -- never for conversion.
  static const Map<String, int> _minorUnits = {
    'MRU': 0,
    'TND': 3,
    'MAD': 2,
    'XOF': 0,
    'EUR': 2,
    'USD': 2,
  };

  static int minorUnitsFor(String? currencyCode) {
    final code = (currencyCode == null || currencyCode.isEmpty)
        ? fallbackCurrencyCode
        : currencyCode;
    return _minorUnits[code] ?? 2;
  }

  /// Formats [amount] with its [currencyCode] as "amount CODE", e.g.
  /// "100 MRU" / "1,250 TND" / "45.5 MAD". Falls back to [fallbackCurrencyCode]
  /// only when [currencyCode] is null/empty (legacy object with no
  /// currency_code at all) -- never silently substitutes a different
  /// currency's code.
  static String format(num amount, String? currencyCode, {bool grouped = true}) {
    final code = (currencyCode == null || currencyCode.isEmpty)
        ? fallbackCurrencyCode
        : currencyCode;
    final digits = _minorUnits[code] ?? 2;

    final fixed = amount.toStringAsFixed(digits);
    final formattedNumber = grouped ? _groupThousands(fixed) : fixed;

    return '$formattedNumber $code';
  }

  /// Same as [format] but returns just the numeric part (no currency code),
  /// useful when the code is displayed separately in the UI.
  static String formatAmountOnly(num amount, String? currencyCode, {bool grouped = true}) {
    final code = (currencyCode == null || currencyCode.isEmpty)
        ? fallbackCurrencyCode
        : currencyCode;
    final digits = _minorUnits[code] ?? 2;
    final fixed = amount.toStringAsFixed(digits);
    return grouped ? _groupThousands(fixed) : fixed;
  }

  static String _groupThousands(String fixed) {
    final parts = fixed.split('.');
    final intPart = parts[0];
    final isNegative = intPart.startsWith('-');
    final digitsOnly = isNegative ? intPart.substring(1) : intPart;

    final buffer = StringBuffer();
    for (int i = 0; i < digitsOnly.length; i++) {
      if (i > 0 && (digitsOnly.length - i) % 3 == 0) {
        buffer.write(',');
      }
      buffer.write(digitsOnly[i]);
    }

    final groupedInt = (isNegative ? '-' : '') + buffer.toString();
    if (parts.length > 1) {
      return '$groupedInt.${parts[1]}';
    }
    return groupedInt;
  }
}
