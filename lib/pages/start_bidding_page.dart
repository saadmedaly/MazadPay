import 'package:flutter/material.dart';
import 'package:mezadpay/pages/terms_page.dart';
import 'package:mezadpay/pages/language_page.dart';
import 'package:mezadpay/core/theme.dart';
import 'package:google_fonts/google_fonts.dart';

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
      'description': 'انضم إلى آلاف المزايدين واقتنص الفرص الحصرية بلمسة واحدة.',
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
                Navigator.of(context).push(
                  MaterialPageRoute(builder: (_) => const LanguagePage()),
                );
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
                    return SingleChildScrollView(
                      child: Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 24.0),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.center,
                          children: [
                            // margin-top: 60px
                            const SizedBox(height: 60),

                            // Hero Illustration داخل دائرة خفيفة
                            Container(
                              width: 220,
                              height: 220,
                              decoration: BoxDecoration(
                                color: const Color(0xFFEAF4FF),
                                shape: BoxShape.circle,
                              ),
                              padding: const EdgeInsets.all(20),
                              child: Image.asset(
                                onboardingData[index]['image']!,
                                fit: BoxFit.contain,
                              ),
                            ),

                            // بين الأيقونة والعنوان: 24px
                            const SizedBox(height: 24),

                            // العنوان الرئيسي
                            Text(
                              onboardingData[index]['title']!,
                              textAlign: TextAlign.center,
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 22,
                                fontWeight: FontWeight.bold,
                                height: 1.5,
                                color: isDarkMode ? Colors.white : const Color(0xFF101828),
                              ),
                            ),

                            // بين العنوان والوصف: 12px
                            const SizedBox(height: 12),

                            // الوصف
                            Text(
                              onboardingData[index]['description']!,
                              textAlign: TextAlign.center,
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 14,
                                fontWeight: FontWeight.w400,
                                height: 1.6,
                                color: isDarkMode ? Colors.white60 : const Color(0xFF6B7280),
                              ),
                            ),

                            // بين الوصف والزر: 32px
                            const SizedBox(height: 32),

                            // Page Indicator
                            Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              textDirection: TextDirection.ltr,
                              children: List.generate(
                                onboardingData.length,
                                (i) => AnimatedContainer(
                                  duration: const Duration(milliseconds: 300),
                                  margin: const EdgeInsets.symmetric(horizontal: 4),
                                  width: _currentPage == i ? 32 : 8,
                                  height: 8,
                                  decoration: BoxDecoration(
                                    color: _currentPage == i
                                        ? primaryColor
                                        : primaryColor.withOpacity(0.2),
                                    borderRadius: BorderRadius.circular(4),
                                  ),
                                ),
                              ),
                            ),

                            const SizedBox(height: 32),

                            // زر CTA: width 90%، height 50px، border-radius 12px
                            SizedBox(
                              width: double.infinity,
                              height: 50,
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
                                  backgroundColor: const Color(0xFF1E88E5),
                                  foregroundColor: Colors.white,
                                  elevation: isDarkMode ? 0 : 3,
                                  shadowColor: const Color(0xFF1E88E5).withOpacity(0.4),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                ),
                                child: Text(
                                  'متابعة',
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 16,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.white,
                                  ),
                                ),
                              ),
                            ),

                            const SizedBox(height: 12),

                            // Skip
                            TextButton(
                              onPressed: () {
                                Navigator.of(context).pushReplacement(
                                  MaterialPageRoute(
                                    builder: (_) => const TermsPage(),
                                  ),
                                );
                              },
                              child: Text(
                                'تخطي للجولة',
                                style: GoogleFonts.plusJakartaSans(
                                  fontSize: 14,
                                  fontWeight: FontWeight.w600,
                                  color: isDarkMode ? Colors.white60 : const Color(0xFF6B7280),
                                ),
                              ),
                            ),

                            const SizedBox(height: 16),
                          ],
                        ),
                      ),
                    );
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
