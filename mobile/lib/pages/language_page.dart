import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mezadpay/providers/locale_provider.dart';

import 'package:mezadpay/pages/onboarding_page.dart';

class LanguagePage extends ConsumerStatefulWidget {
  const LanguagePage({super.key});

  @override
  ConsumerState<LanguagePage> createState() => _LanguagePageState();
}

class _LanguagePageState extends ConsumerState<LanguagePage> {
  String selectedLanguage = 'Arabic';

  late List<Map<String, String>> languages;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    
    // Sync selectedLanguage with the current locale from Riverpod
    final currentLocale = ref.watch(localeNotifierProvider);
    if (currentLocale.languageCode == 'ar') {
      selectedLanguage = 'Arabic';
    } else if (currentLocale.languageCode == 'en') {
      selectedLanguage = 'English';
    } else if (currentLocale.languageCode == 'fr') {
      selectedLanguage = 'French';
    }

    languages = [
      {'name': 'العربية', 'key': 'Arabic', 'code': 'ar'},
      {'name': 'English', 'key': 'English', 'code': 'en'},
      {'name': 'Français', 'key': 'French', 'code': 'fr'}, 
    ];
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    
    return Scaffold(
        backgroundColor: Theme.of(context).scaffoldBackgroundColor,
        body: SingleChildScrollView(
          child: Column(
            children: [
              // Top Illustration Background
              Stack(
                children: [
                  Container(
                    height: MediaQuery.of(context).size.height * 0.3,
                    width: double.infinity,
                    decoration: const BoxDecoration(
                      image: DecorationImage(
                        image: AssetImage('assets/lang_bg.png'),
                        fit: BoxFit.cover,
                      ),
                    ),
                  ),
                  Positioned(
                    top: 60,
                    right: 20,
                    child: Text(
                      AppLocalizations.of(context)!.text_210,
                      style: const TextStyle(
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                        fontSize: 28,
                      ),
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 20),

              // Language List
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: languages.map((lang) {
                    bool isSelected = selectedLanguage == lang['key'];
                    return Padding(
                      padding: const EdgeInsetsDirectional.only(bottom: 12),
                      child: InkWell(
                        onTap: () {
                          setState(() {
                            selectedLanguage = lang['key']!;
                          });
                          // Change the app locale via Riverpod
                          ref.read(localeNotifierProvider.notifier).setLocale(Locale(lang['code']!));
                        },
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 20,
                            vertical: 18,
                          ),
                          decoration: BoxDecoration(
                            color: isSelected
                                ? (isDarkMode
                                      ? const Color(0xFF135BEC).withOpacity(0.15)
                                      : const Color(0xFFE8F0FE))
                                : (isDarkMode
                                      ? const Color(0xFF1D1D1D)
                                      : Colors.white),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: isSelected
                                  ? const Color(0xFF135BEC)
                                  : (isDarkMode
                                        ? const Color(0xFF333333)
                                        : const Color(0xFFEEEEEE)),
                              width: 1.5,
                            ),
                          ),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(
                                lang['name']!,
                                style: TextStyle(
                                  fontSize: 18,
                                  fontWeight: isSelected
                                      ? FontWeight.bold
                                      : FontWeight.w500,
                                  color: isSelected
                                      ? const Color(0xFF135BEC)
                                      : (isDarkMode
                                            ? Colors.white
                                            : Colors.black),
                                ),
                              ),
                              if (isSelected)
                                const Icon(
                                  Icons.check_circle,
                                  color: Color(0xFF135BEC),
                                ),
                            ],
                          ),
                        ),
                      ),
                    );
                  }).toList(),
                ),
              ),
              const SizedBox(height: 20),
            ],
          ),
        ),
        bottomNavigationBar: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: ElevatedButton(
              onPressed: () {
                Navigator.of(context).pushReplacement(
                  MaterialPageRoute(builder: (_) => const OnboardingPage()),
                );
              },
              child: Text(AppLocalizations.of(context)!.text_211),
            ),
          ),
        ),
      );
  }
}