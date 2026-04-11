import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:mezadpay/pages/terms_page.dart';

class MazadIntroPage extends StatefulWidget {
  const MazadIntroPage({super.key});

  @override
  State<MazadIntroPage> createState() => _MazadIntroPageState();
}

class _MazadIntroPageState extends State<MazadIntroPage> {
  final PageController _pageController = PageController();
  int _currentPage = 0;

  @override
  Widget build(BuildContext context) {
    final List<Map<String, String>> onboardingData = [
      {
        'image': 'assets/onboarding2.png',
        'title': AppLocalizations.of(context)!.text_226,
        'description': AppLocalizations.of(context)!.text_227,
      },
      {
        'image': 'assets/MazadIntro.png',
        'title': AppLocalizations.of(context)!.text_228,
        'description': AppLocalizations.of(context)!.text_229,
      },
      {
        'image': 'assets/onboarding1.png',
        'title': AppLocalizations.of(context)!.text_230,
        'description': AppLocalizations.of(context)!.text_231,
      },
    ];

    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    // Pixel-perfect primary blue from the design
    const Color primaryBlue = Color(0xFF0084FF);

    return Scaffold(
        backgroundColor: Theme.of(context).scaffoldBackgroundColor,
        body: SafeArea(
          child: Column(
            children: [
              Expanded(
                child: PageView.builder(
                  controller: _pageController,
                  onPageChanged: (index) {
                    setState(() {
                      _currentPage = index;
                    });
                  },
                  itemCount: onboardingData.length,
                  itemBuilder: (context, index) {
                    return Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 24.0),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Expanded(
                            flex: 5,
                            child: Container(
                              width: double.infinity,
                              decoration: BoxDecoration(
                                color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                                borderRadius: BorderRadius.circular(32),
                                boxShadow: [
                                  if (!isDarkMode)
                                    BoxShadow(
                                      color: Colors.black.withOpacity(0.04),
                                      blurRadius: 20,
                                      offset: const Offset(0, 8),
                                    ),
                                ],
                              ),
                              child: Center(
                                child: Padding(
                                  padding: const EdgeInsets.all(24.0),
                                  child: Image.asset(
                                    onboardingData[index]['image']!,
                                    fit: BoxFit.contain,
                                  ),
                                ),
                              ),
                            ),
                          ),
                          const SizedBox(height: 40),
                          Text(
                            onboardingData[index]['title']!,
                            textAlign: TextAlign.center,
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 26,
                              fontWeight: FontWeight.w800,
                              color: isDarkMode ? Colors.white : const Color(0xFF101828),
                            ),
                          ),
                          const SizedBox(height: 16),
                          Text(
                            onboardingData[index]['description']!,
                            textAlign: TextAlign.center,
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 16,
                              fontWeight: FontWeight.w500,
                              height: 1.5,
                              color: isDarkMode ? Colors.white70 : const Color(0xFF667085),
                            ),
                          ),
                        ],
                      ),
                    );
                  },
                ),
              ),

              // Bottom Section: Dots & Buttons
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 32,
                  vertical: 8,
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    // Page Indicators
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: List.generate(
                        onboardingData.length,
                        (index) => AnimatedContainer(
                          duration: const Duration(milliseconds: 300),
                          margin: const EdgeInsets.symmetric(horizontal: 4),
                          width: _currentPage == index ? 24 : 8,
                          height: 8,
                          decoration: BoxDecoration(
                            color: _currentPage == index ? primaryBlue : primaryBlue.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 20),

                    // Continue Button
                    SizedBox(
                      width: double.infinity,
                      height: 48,
                      child: ElevatedButton(
                        onPressed: () {
                          if (_currentPage < onboardingData.length - 1) {
                            _pageController.nextPage(
                              duration: const Duration(milliseconds: 300),
                              curve: Curves.easeInOut,
                            );
                          } else {
                            Navigator.of(context).pushReplacement(
                              MaterialPageRoute(
                                builder: (_) => const TermsPage(),
                              ),
                            );
                          }
                        },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: primaryBlue,
                          foregroundColor: Colors.white,
                          elevation: 0,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16),
                          ),
                        ),
                        child: Text(
                          _currentPage == onboardingData.length - 1
                              ? AppLocalizations.of(context)!.text_232
                              : AppLocalizations.of(context)!.text_144,
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 16,
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 12),

                    // Skip Button
                    TextButton(
                      onPressed: () {
                        Navigator.of(context).pushReplacement(
                          MaterialPageRoute(builder: (_) => const TermsPage()),
                        );
                      },
                      child: Text(
                        AppLocalizations.of(context)!.text_233,
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                          color: isDarkMode ? Colors.white60 : const Color(0xFF667085),
                        ),
                      ),
                    ),
                    const SizedBox(height: 8), // Bottom safe space
                  ],
                ),
              ),
            ],
          ),
        ),
      );
  }
}
