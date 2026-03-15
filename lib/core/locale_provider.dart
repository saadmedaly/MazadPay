import 'package:flutter/material.dart';

class LocaleProvider extends ChangeNotifier {
  Locale _locale = const Locale('ar');

  Locale get locale => _locale;

  void setLocale(Locale locale) {
    if (!['en', 'fr', 'ar'].contains(locale.languageCode)) return;
    _locale = locale;
    notifyListeners();
  }

  void clearLocale() {
    _locale = const Locale('ar');
    notifyListeners();
  }
}
