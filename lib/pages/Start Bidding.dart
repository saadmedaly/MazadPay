import 'package:flutter/material.dart';
import 'package:mezadpay/pages/terms_page.dart';
import 'package:mezadpay/pages/language_page.dart';
import 'package:mezadpay/core/theme.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:mezadpay/pages/MazadIntro.dart';

class StartBiddingPage extends StatefulWidget {
  const StartBiddingPage({super.key});

  @override
  State<StartBiddingPage> createState() => _StartBiddingPageState();
}

class _StartBiddingPageState extends State<StartBiddingPage> {
  final PageController _pageController = PageController();
  int _currentPage = 0;

  final List<Map<String, String>> onboardingData = [
    {
      'image': 'assets/Start Bidding.png',
      'title': 'زايد الآن لتكون الفائز الأقرب بالصفقة!',
      'description':
          'انضم إلى آلاف المزايدين واقتنص الفرص الحصرية بلمسة واحدة.',
    },
  ];

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final Color primaryColor = AppTheme.primaryColor;

    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: Theme.of(context).scaffoldBackgroundColor,
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          leading: Container(
            margin: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: isDarkMode
                  ? const Color(0xFF1D1D1D)
                  : const Color(0xFFF2F4F7),
              borderRadius: BorderRadius.circular(20),
            ),
            child: IconButton(
              icon: Icon(
                Icons.language,
                color: isDarkMode ? Colors.white : Colors.black,
                size: 20,
              ),
              onPressed: () {
                // Navigate to Language Selection or show picker
                Navigator.of(
                  context,
                ).push(MaterialPageRoute(builder: (_) => const LanguagePage()));
              },
            ),
          ),
          actions: [
            if (_currentPage > 0)
              Container(
                margin: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: isDarkMode
                      ? const Color(0xFF1D1D1D)
                      : const Color(0xFFF2F4F7),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: IconButton(
                  icon: Icon(
                    Icons.arrow_forward_ios,
                    color: isDarkMode ? Colors.white : Colors.black,
                    size: 18,
                  ),
                  onPressed: () {
                    _pageController.previousPage(
                      duration: const Duration(milliseconds: 300),
                      curve: Curves.easeInOut,
                    );
                  },
                ),
              ),
          ],
        ),
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
                          // Image Card
                          Expanded(
                            flex: 6,
                            child: Container(
                              width: double.infinity,
                              decoration: BoxDecoration(
                                color: isDarkMode
                                    ? const Color(0xFF1D1D1D)
                                    : Colors.white,
                                borderRadius: BorderRadius.circular(32),
                              ),
                              child: Stack(
                                alignment: Alignment.center,
                                children: [
                                  // Background soft circle
                                  Positioned(
                                    top: 40,
                                    child: Container(
                                      width: 320,
                                      height: 320,
                                      decoration: BoxDecoration(
                                        color: primaryColor.withOpacity(0.08),
                                        shape: BoxShape.circle,
                                      ),
                                    ),
                                  ),
                                  Padding(
                                    padding: const EdgeInsets.all(0.0),
                                    child: Transform.scale(
                                      scale: 1.3,
                                      child: Image.asset(
                                        onboardingData[index]['image']!,
                                        fit: BoxFit.contain,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                          const SizedBox(height: 48),
                          // Content
                          Text(
                            onboardingData[index]['title']!,
                            textAlign: TextAlign.center,
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 28,
                              fontWeight: FontWeight.w800,
                              height: 1.2,
                              color: isDarkMode
                                  ? Colors.white
                                  : const Color(0xFF101828),
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
                              color: isDarkMode
                                  ? Colors.white70
                                  : const Color(0xFF667085),
                            ),
                          ),
                          const Spacer(flex: 1),
                        ],
                      ),
                    );
                  },
                ),
              ),

              // Bottom Section
              Padding(
                padding: const EdgeInsets.fromLTRB(24, 8, 24, 32),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    // Page Indicator (Active Bar)
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      textDirection: TextDirection.ltr,
                      children: List.generate(
                        onboardingData.length,
                        (index) => AnimatedContainer(
                          duration: const Duration(milliseconds: 300),
                          margin: const EdgeInsets.symmetric(horizontal: 4),
                          width: _currentPage == index ? 32 : 8,
                          height: 8,
                          decoration: BoxDecoration(
                            color: _currentPage == index
                                ? primaryColor
                                : primaryColor.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(4),
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 32),

                    // Primary Button
                    SizedBox(
                      width: double.infinity,
                      height: 56,
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
                                builder: (_) => const MazadIntroPage(),
                              ),
                            );
                          }
                        },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: primaryColor,
                          foregroundColor: Colors.white,
                          elevation: isDarkMode ? 0 : 4,
                          shadowColor: primaryColor.withOpacity(0.4),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16),
                          ),
                        ),
                        child: Text(
                          _currentPage == onboardingData.length - 1
                              ? 'ابدأ المزايدة الآن'
                              : 'متابعة',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(height: 12),

                    // Skip Text Button
                    TextButton(
                      onPressed: () {
                        Navigator.of(context).pushReplacement(
                          MaterialPageRoute(builder: (_) => const MazadIntroPage()),
                        );
                      },
                      style: TextButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 12),
                      ),
                      child: Text(
                        'تخطي ',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                          color: isDarkMode
                              ? Colors.white60
                              : const Color(0xFF667085),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
