import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import '../utils/phone_validation.dart';

/// Shared searchable bottom-sheet country picker.
///
/// Used by the phone auth screens (registration, login, phone+password) so
/// the ~195-country list returned by `GET /countries` stays usable — a plain
/// unfiltered list would be unwieldy at that size.
Future<void> showCountryPickerSheet(
  BuildContext context, {
  required List<dynamic> countries,
  required void Function(Map<String, dynamic> country) onSelected,
}) {
  final usable = filterUsableCountries(countries);
  bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

  return showModalBottomSheet(
    context: context,
    backgroundColor: Colors.transparent,
    isScrollControlled: true,
    builder: (sheetContext) {
      return DraggableScrollableSheet(
        initialChildSize: 0.75,
        minChildSize: 0.4,
        maxChildSize: 0.92,
        expand: false,
        builder: (context, scrollController) {
          return _CountryPickerBody(
            isDarkMode: isDarkMode,
            scrollController: scrollController,
            countries: usable,
            onSelected: (c) {
              onSelected(c);
              Navigator.pop(context);
            },
          );
        },
      );
    },
  );
}

class _CountryPickerBody extends StatefulWidget {
  final bool isDarkMode;
  final ScrollController scrollController;
  final List<dynamic> countries;
  final void Function(Map<String, dynamic> country) onSelected;

  const _CountryPickerBody({
    required this.isDarkMode,
    required this.scrollController,
    required this.countries,
    required this.onSelected,
  });

  @override
  State<_CountryPickerBody> createState() => _CountryPickerBodyState();
}

class _CountryPickerBodyState extends State<_CountryPickerBody> {
  String _query = '';

  List<dynamic> get _filtered {
    if (_query.trim().isEmpty) return widget.countries;
    return widget.countries
        .where((c) => countryMatchesQuery(c as Map<String, dynamic>, _query))
        .toList();
  }

  @override
  Widget build(BuildContext context) {
    final isDarkMode = widget.isDarkMode;
    final l10n = AppLocalizations.of(context)!;
    final filtered = _filtered;

    return Container(
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: const BorderRadiusDirectional.only(
          topStart: Radius.circular(24),
          topEnd: Radius.circular(24),
        ),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 20),
      child: Column(
        children: [
          Container(
            width: 40,
            height: 4,
            decoration: BoxDecoration(
              color: isDarkMode ? Colors.grey[700] : Colors.grey[300],
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          const SizedBox(height: 16),
          Text(
            l10n.text_219,
            style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 16),
          TextField(
            autofocus: false,
            textAlign: TextAlign.right,
            onChanged: (v) => setState(() => _query = v),
            decoration: InputDecoration(
              hintText: l10n.search_country_hint,
              prefixIcon: const Icon(Icons.search, color: Colors.grey),
              filled: true,
              fillColor: isDarkMode ? Colors.black26 : Colors.grey[100],
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(16),
                borderSide: BorderSide.none,
              ),
            ),
          ),
          const SizedBox(height: 16),
          Expanded(
            child: filtered.isEmpty
                ? Center(
                    child: Text(
                      l10n.error_no_data,
                      style: const TextStyle(color: Colors.grey),
                    ),
                  )
                : ListView.separated(
                    controller: widget.scrollController,
                    itemCount: filtered.length,
                    separatorBuilder: (context, index) =>
                        const SizedBox(height: 12),
                    itemBuilder: (context, index) {
                      final country = filtered[index] as Map<String, dynamic>;
                      final isArabic =
                          Localizations.localeOf(context).languageCode == 'ar';
                      final isFrench =
                          Localizations.localeOf(context).languageCode == 'fr';
                      final name = isArabic
                          ? country['name_ar']
                          : (isFrench
                                ? country['name_fr']
                                : country['name_en']);
                      final isAvailable = country['is_active'] ?? true;

                      return Opacity(
                        opacity: isAvailable ? 1.0 : 0.5,
                        child: InkWell(
                          onTap: isAvailable
                              ? () => widget.onSelected(country)
                              : null,
                          borderRadius: BorderRadius.circular(12),
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 12,
                            ),
                            decoration: BoxDecoration(
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(
                                color: isAvailable
                                    ? const Color(0xFF135BEC).withOpacity(0.3)
                                    : Colors.grey.withOpacity(0.2),
                              ),
                            ),
                            child: Row(
                              children: [
                                ClipRRect(
                                  borderRadius: BorderRadius.circular(2),
                                  child: Image.network(
                                    'https://flagcdn.com/w80/${country['code'].toString().toLowerCase()}.png',
                                    width: 32,
                                    height: 20,
                                    fit: BoxFit.cover,
                                    errorBuilder:
                                        (context, error, stackTrace) =>
                                            Container(
                                              width: 32,
                                              height: 20,
                                              color: Colors.grey[300],
                                              child: const Icon(
                                                Icons.flag,
                                                size: 12,
                                              ),
                                            ),
                                  ),
                                ),
                                const SizedBox(width: 16),
                                Expanded(
                                  child: Text(
                                    name?.toString() ?? 'Unknown',
                                    style: const TextStyle(
                                      fontSize: 16,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                                Text(
                                  country['country_code']?.toString() ?? '',
                                  style: const TextStyle(
                                    fontSize: 16,
                                    color: Colors.grey,
                                    fontWeight: FontWeight.w500,
                                  ),
                                  textDirection: TextDirection.ltr,
                                ),
                              ],
                            ),
                          ),
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}
