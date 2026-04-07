import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class AboutMazadPayPage extends StatelessWidget {
  const AboutMazadPayPage({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    
    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : Colors.white,
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          leading: IconButton(
            icon: Icon(
              Icons.arrow_back_ios,
              color: isDarkMode ? Colors.white : Colors.black,
            ),
            onPressed: () => Navigator.of(context).pop(),
          ),
          title: Text(
            'حول مزاد باي (Mazad Pay)',
            style: GoogleFonts.plusJakartaSans(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: isDarkMode ? Colors.white : Colors.black,
            ),
          ),
          centerTitle: true,
        ),
        body: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildSectionTitle('حول مزاد باي (Mazad Pay)', isDarkMode),
              const SizedBox(height: 12),
              _buildParagraph(
                'مزاد باي هو أول تطبيق للمزادات الرقمية في موريتانيا، والمنصة الرائدة التي تنقل مفهوم المزايدة التقليدية إلى تجربة تقنية حديثة، آمنة، وشفافة. نحن فخورون لكوننا أول من أطلق هذا النظام المتكامل في السوق الموريتاني، لنجمع لك بين إثارة المزايدة وتلبية احتياجاتك اليومية في تطبيق واحد وبسواعد وطنية.',
                isDarkMode,
              ),
              const SizedBox(height: 24),
              
              _buildSectionTitle('خدماتنا الرائدة:', isDarkMode),
              const SizedBox(height: 12),
              _buildParagraph(
                'من خلال تطبيقنا، نضع بين يديك باقة متنوعة من الخدمات المصممة خصيصاً لمجتمعنا:',
                isDarkMode,
              ),
              const SizedBox(height: 12),
              _buildBulletPoint(
                'المزادات الرقمية (الأولى في موريتانيا): نظام مزايدة علني وشفاف على السيارات، العقارات، والإلكترونيات، يضمن حقوق الجميع بكل موثوقية.',
                isDarkMode,
              ),
              _buildBulletPoint(
                'التجارة الإلكترونية العالمية: منصة تسوق ذكية تربطك بأفضل المواقع العالمية مثل (Amazon, AliExpress, Shein) لتصلك منتجاتك المفضلة أينما كنت.',
                isDarkMode,
              ),
              _buildBulletPoint(
                'خدمات التوصيل والنقل الشاملة: حلول متكاملة تشمل نقل الأشخاص، شحن البضائع، وخدمات سحب ونقل السيارات (الحاملات).',
                isDarkMode,
              ),
              _buildBulletPoint(
                'حجز الأنشطة الرياضية: إمكانية حجز مختلف أنواع الرياضات؛ بما في ذلك صالات اللياقة البدنية، الملاعب، وحصص السباحة.',
                isDarkMode,
              ),
              _buildBulletPoint(
                'طلب الطعام: استمتع بطلب وجباتك من أشهر المطاعم ومقاهي بضغطة زر.',
                isDarkMode,
              ),
              _buildBulletPoint(
                'حجز الفنادق والإقامات: خطط لرحلاتك داخل وخارج موريتانيا بأسعار تنافسية.',
                isDarkMode,
              ),
              _buildBulletPoint(
                'المواعيد الطبية: تنظيم وتسهيل حجز المواعيد في كبرى المصحات والعيادات لضمان راحتك وصحتك.',
                isDarkMode,
              ),
              
              const SizedBox(height: 32),
              _buildSectionTitle('رؤيتنا', isDarkMode),
              const SizedBox(height: 12),
              _buildParagraph(
                'الريادة المطلقة كأول وجهة رقمية للمزادات في موريتانيا، وتقديم حلول شاملة تدمج بين التجارة الإلكترونية، مختلف أنواع توصيل، وخدمات، لتعزيز نمط حياة المواطن الموريتاني وتسهيل معاملاته.',
                isDarkMode,
              ),
              
              const SizedBox(height: 32),
              _buildSectionTitle('التزامنا', isDarkMode),
              const SizedBox(height: 12),
              _buildParagraph(
                'بصفتنا أول مزاد رقمي في البلاد، يلتزم فريق مزاد باي بتوفير بيئة رقمية آمنة كلياً، تلتزم بأعلى معايير المصداقية والشفافية، وتدعم التحول الرقمي في موريتانيا.',
                isDarkMode,
              ),
              const SizedBox(height: 40),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSectionTitle(String title, bool isDarkMode) {
    return Text(
      title,
      style: GoogleFonts.plusJakartaSans(
        fontSize: 20,
        fontWeight: FontWeight.w800,
        color: isDarkMode ? Colors.white : Colors.black,
      ),
    );
  }

  Widget _buildParagraph(String text, bool isDarkMode) {
    return Text(
      text,
      textAlign: TextAlign.justify,
      style: GoogleFonts.plusJakartaSans(
        fontSize: 15,
        height: 1.6,
        color: isDarkMode ? Colors.white70 : Colors.black87,
      ),
    );
  }

  Widget _buildBulletPoint(String text, bool isDarkMode) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            margin: const EdgeInsets.only(top: 8, left: 12),
            width: 6,
            height: 6,
            decoration: BoxDecoration(
              color: isDarkMode ? Colors.white38 : Colors.black38,
              shape: BoxShape.circle,
            ),
          ),
          Expanded(
            child: Text(
              text,
              style: GoogleFonts.plusJakartaSans(
                fontSize: 15,
                height: 1.5,
                color: isDarkMode ? Colors.white70 : Colors.black87,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
