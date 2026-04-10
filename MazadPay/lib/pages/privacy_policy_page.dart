import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class PrivacyPolicyPage extends StatelessWidget {
  const PrivacyPolicyPage({super.key});

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          title: Text(
            'سياسة الخصوصية',
            style: GoogleFonts.plusJakartaSans(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: isDarkMode ? Colors.white : Colors.black,
            ),
          ),
          leading: IconButton(
            icon: Icon(Icons.arrow_back_ios, color: isDarkMode ? Colors.white : Colors.black, size: 20),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ),
        body: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
               _buildSection('1. المقدمة', 'نحن في مزاد باي (MazadPay) نلتزم بحماية خصوصيتك ومعلوماتك الشخصية. توضح هذه السياسة كيفية جمع واستخدام وحماية بياناتك عند استخدام تطبيقنا خدماتنا.'),
               const SizedBox(height: 24),
               _buildSection('2. المعلومات التي نجمعها', 'نقوم بجمع المعلومات التي تقدمها لنا مباشرة عند إنشاء حساب، مثل الاسم، رقم الهاتف، والبريد الإلكتروني. كما نسجل بيانات المعاملات المالية والمزايدات التي تشارك بها.'),
               const SizedBox(height: 24),
               _buildSection('3. كيفية استخدام البيانات', 'نستخدم معلوماتك لمعالجة المزايدات، إدارة حسابك، إرسال تنبيهات السعر، وضمان أمان المعاملات المالية داخل التطبيق.'),
               const SizedBox(height: 24),
               _buildSection('4. حماية المعلومات', 'نحن نستخدم تقنيات تشفير متطورة لحماية بياناتك من الوصول غير المصرح به. معلوماتك المالية يتم معالجتها من خلال بوابات دفع آمنة ومعتمدة.'),
               const SizedBox(height: 24),
               _buildSection('5. التغييرات على هذه السياسة', 'نحتفظ بالحق في تحديث سياسة الخصوصية هذه من وقت لآخر. سيتم إخطارك بأي تغييرات جوهرية عبر البريد الإلكتروني أو تنبيه داخل التطبيق.'),
               const SizedBox(height: 40),
               Center(
                 child: Text(
                   'آخر تحديث: مارس 2026',
                   style: GoogleFonts.plusJakartaSans(color: Colors.grey, fontSize: 12),
                 ),
               ),
               const SizedBox(height: 24),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSection(String title, String content) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
         Text(
           title,
           style: GoogleFonts.plusJakartaSans(fontSize: 18, fontWeight: FontWeight.bold, color: const Color(0xFF0081FF)),
         ),
         const SizedBox(height: 12),
         Text(
           content,
           style: GoogleFonts.plusJakartaSans(fontSize: 14, height: 1.6, color: Colors.grey[800]),
           textAlign: TextAlign.justify,
         ),
      ],
    );
  }
}
