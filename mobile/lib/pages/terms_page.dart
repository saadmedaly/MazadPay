import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:mezadpay/pages/create_profile_page.dart';

class TermsPage extends StatefulWidget {
  const TermsPage({super.key});

  @override
  State<TermsPage> createState() => _TermsPageState();
}

class _TermsPageState extends State<TermsPage> {
  int currentPage = 1;

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Scaffold(
        backgroundColor: Theme.of(context).scaffoldBackgroundColor,
        appBar: AppBar(
          centerTitle: true,
          elevation: 0,
          backgroundColor: Colors.transparent,
          title: Text(
            AppLocalizations.of(context)!.text_318,
            style: TextStyle(
              fontWeight: FontWeight.bold,
              color: isDarkMode ? Colors.white : Colors.black,
              fontSize: 18,
            ),
          ),
          leading: IconButton(
            icon: Icon(
              Icons.arrow_back_ios_new,
              color: isDarkMode ? Colors.white : Colors.black,
              size: 20,
            ),
            onPressed: () {
              if (currentPage == 2) {
                setState(() => currentPage = 1);
              } else {
                Navigator.of(context).pop();
              }
            },
          ),
          actions: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8),
              child: IconButton(
                icon: Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: isDarkMode
                        ? const Color(0xFF1D1D1D)
                        : const Color(0xFFF2F4F7),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    Icons.close,
                    color: isDarkMode ? Colors.white : Colors.black,
                    size: 16,
                  ),
                ),
                onPressed: () => Navigator.of(context).pop(),
              ),
            ),
          ],
        ),
        body: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Container(
            decoration: BoxDecoration(
              color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: isDarkMode
                    ? const Color(0xFF333333)
                    : const Color(0xFFF2F4F7),
                width: 1.5,
              ),
            ),
            child: IntrinsicHeight(
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  // Text Content
                  Expanded(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: currentPage == 1
                            ? _buildPage1()
                            : _buildPage2(),
                      ),
                    ),
                  ),
                  // Blue Indicator
                  Container(
                    width: 4,
                    decoration: const BoxDecoration(
                      color: Color(0xFF135BEC),
                      borderRadius: BorderRadiusDirectional.only(
                        topEnd: Radius.circular(4),
                        bottomEnd: Radius.circular(4),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
        bottomNavigationBar: SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(24, 8, 24, 16),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  children: [
                    if (currentPage == 2) ...[
                      Expanded(
                        child: OutlinedButton(
                          onPressed: () {
                            setState(() => currentPage = 1);
                          },
                          style: OutlinedButton.styleFrom(
                            side: const BorderSide(color: Color(0xFFD0D5DD)),
                            padding: const EdgeInsets.symmetric(vertical: 16),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(8),
                            ),
                          ),
                          child: Text(
                            AppLocalizations.of(context)!.text_319,
                            style: TextStyle(
                              color: Color(0xFF344054),
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                    ],
                    Expanded(
                      child: ElevatedButton(
                        onPressed: () {
                          if (currentPage == 1) {
                            setState(() => currentPage = 2);
                          } else {
                            Navigator.of(context).pushReplacement(
                              MaterialPageRoute(
                                builder: (_) => const CreateProfilePage(),
                              ),
                            );
                          }
                        },
                        child: Text(
                          currentPage == 1 ? AppLocalizations.of(context)!.text_144 : AppLocalizations.of(context)!.text_320,
                          style: const TextStyle(fontWeight: FontWeight.bold,fontSize: 12),
                        ),
                      ),
                    ),
                  ],
                ),
                if (currentPage == 2) ...[
                  const SizedBox(height: 16),
                  Text(
                    AppLocalizations.of(context)!.text_321,
                    style: TextStyle(color: Color(0xFF667085), fontSize: 12),
                  ),
                ],
              ],
            ),
          ),
        ),
      );
  }

  List<Widget> _buildPage1() {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return [
      Text(
        AppLocalizations.of(context)!.text_322,
        style: TextStyle(
          fontWeight: FontWeight.bold,
          fontSize: 18,
          color: isDarkMode ? Colors.white : Colors.black,
        ),
      ),
      const SizedBox(height: 16),
      _sectionHeader(AppLocalizations.of(context)!.text_323),
      _sectionBody(
        AppLocalizations.of(context)!.text_324,
      ),
      const SizedBox(height: 16),
      _sectionHeader(AppLocalizations.of(context)!.text_325),
      _sectionBody(
        AppLocalizations.of(context)!.text_326,
      ),
      const SizedBox(height: 16),
      _sectionHeader(AppLocalizations.of(context)!.text_327),
      _sectionBody(AppLocalizations.of(context)!.text_328),
      _sectionBody(
        AppLocalizations.of(context)!.text_329,
      ),
    ];
  }

  List<Widget> _buildPage2() {
    return [
      _sectionHeader(AppLocalizations.of(context)!.text_330),
      _bulletPoint(AppLocalizations.of(context)!.text_331),
      _bulletPoint(AppLocalizations.of(context)!.text_332),
      const SizedBox(height: 16),
      _sectionHeader(AppLocalizations.of(context)!.text_333),
      _bulletPoint(
        AppLocalizations.of(context)!.text_334,
      ),
      const SizedBox(height: 16),
      _sectionHeader(AppLocalizations.of(context)!.text_335),
      _bulletPoint(
        AppLocalizations.of(context)!.text_336,
      ),
      const SizedBox(height: 16),
      _sectionHeader(AppLocalizations.of(context)!.text_337),
      _bulletPoint(AppLocalizations.of(context)!.text_338),
      _bulletPoint(
        AppLocalizations.of(context)!.text_339,
      ),
      _bulletPoint(
        AppLocalizations.of(context)!.text_340,
      ),
      _bulletPoint(
        AppLocalizations.of(context)!.text_341,
      ),
      _bulletPoint(
        AppLocalizations.of(context)!.text_342,
      ),
    ];
  }

  Widget _sectionHeader(String title) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Padding(
      padding: const EdgeInsetsDirectional.only(bottom: 8),
      child: Text(
        title,
        style: TextStyle(
          fontWeight: FontWeight.bold,
          fontSize: 16,
          color: isDarkMode ? Colors.white : Colors.black,
        ),
      ),
    );
  }

  Widget _sectionBody(String body) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Padding(
      padding: const EdgeInsetsDirectional.only(bottom: 8),
      child: Text(
        body,
        style: TextStyle(
          color: isDarkMode ? const Color(0xFFD0D5DD) : const Color(0xFF475467),
          fontSize: 14,
          height: 1.5,
        ),
      ),
    );
  }

  Widget _bulletPoint(String text) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Padding(
      padding: const EdgeInsetsDirectional.only(bottom: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsetsDirectional.only(top: 6),
            child: Icon(
              Icons.circle,
              size: 6,
              color: isDarkMode
                  ? const Color(0xFFD0D5DD)
                  : const Color(0xFF475467),
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              text,
              style: TextStyle(
                color: isDarkMode
                    ? const Color(0xFFD0D5DD)
                    : const Color(0xFF475467),
                fontSize: 14,
                height: 1.5,
              ),
            ),
          ),
        ],
      ),
    );
  }
}