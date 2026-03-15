import 'package:flutter/material.dart';
import 'package:mezadpay/pages/login_page.dart';
import 'package:mezadpay/l10n/app_localizations.dart';

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
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
        backgroundColor: Theme.of(context).scaffoldBackgroundColor,
        appBar: AppBar(
          centerTitle: true,
          elevation: 0,
          backgroundColor: Colors.transparent,
          title: Text(
            l10n.termsOfUse,
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
                      borderRadius: BorderRadius.only(
                        topRight: Radius.circular(4),
                        bottomRight: Radius.circular(4),
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
            padding: const EdgeInsets.all(24),
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
                            l10n.previous,
                            style: const TextStyle(
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
                                builder: (_) => const LoginPage(),
                              ),
                            );
                          }
                        },
                        child: Text(
                          currentPage == 1 ? l10n.continueText : l10n.iAgreeToTerms,
                          style: const TextStyle(fontWeight: FontWeight.bold),
                        ),
                      ),
                    ),
                  ],
                ),
                if (currentPage == 2) ...[
                  const SizedBox(height: 16),
                  Text(
                    l10n.lastUpdate,
                    style: const TextStyle(color: Color(0xFF667085), fontSize: 12),
                  ),
                ],
              ],
            ),
          ),
        ),
      );
  }

  List<Widget> _buildPage1() {
    final l10n = AppLocalizations.of(context)!;
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return [
      Text(
        l10n.termsOfUseTitle,
        style: TextStyle(
          fontWeight: FontWeight.bold,
          fontSize: 18,
          color: isDarkMode ? Colors.white : Colors.black,
        ),
      ),
      const SizedBox(height: 16),
      _sectionHeader(l10n.agreeToTermsHeader),
      _sectionBody(
        l10n.agreeToTermsBody,
      ),
      const SizedBox(height: 16),
      _sectionHeader(l10n.eligibilityHeader),
      _sectionBody(
        l10n.eligibilityBody,
      ),
      const SizedBox(height: 16),
      _sectionHeader(l10n.accountSecurityHeader),
      _sectionBody(l10n.accountSecurityBody1),
      _sectionBody(
        l10n.accountSecurityBody2,
      ),
    ];
  }

  List<Widget> _buildPage2() {
    final l10n = AppLocalizations.of(context)!;
    return [
      _sectionHeader(l10n.modificationsHeader),
      _bulletPoint(l10n.modificationsBullet1),
      _bulletPoint(l10n.modificationsBullet2),
      const SizedBox(height: 16),
      _sectionHeader(l10n.terminationHeader),
      _bulletPoint(
        l10n.terminationBullet1,
      ),
      const SizedBox(height: 16),
      _sectionHeader(l10n.lawHeader),
      _bulletPoint(
        l10n.lawBullet1,
      ),
      const SizedBox(height: 16),
      _sectionHeader(l10n.participationTermsHeader),
      _bulletPoint(l10n.participationBullet1),
      _bulletPoint(
        l10n.participationBullet2,
      ),
      _bulletPoint(
        l10n.participationBullet3,
      ),
      _bulletPoint(
        l10n.participationBullet4,
      ),
      _bulletPoint(
        l10n.participationBullet5,
      ),
    ];
  }

  Widget _sectionHeader(String title) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
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
      padding: const EdgeInsets.only(bottom: 8),
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
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.only(top: 6),
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
