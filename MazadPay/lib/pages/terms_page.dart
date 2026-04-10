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
    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: Theme.of(context).scaffoldBackgroundColor,
        appBar: AppBar(
          centerTitle: true,
          elevation: 0,
          backgroundColor: Colors.transparent,
          title: Text(
            'شروط الاستخدام',
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
                          child: const Text(
                            'السابق',
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
                          currentPage == 1 ? 'متابعة' : 'أوافق على الشروط',
                          style: const TextStyle(fontWeight: FontWeight.bold,fontSize: 12),
                        ),
                      ),
                    ),
                  ],
                ),
                if (currentPage == 2) ...[
                  const SizedBox(height: 16),
                  const Text(
                    'آخر تحديث: يونيو 2024',
                    style: TextStyle(color: Color(0xFF667085), fontSize: 12),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }

  List<Widget> _buildPage1() {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return [
      Text(
        'شروط استخدام تطبيق "مزاد موريتانيا"',
        style: TextStyle(
          fontWeight: FontWeight.bold,
          fontSize: 18,
          color: isDarkMode ? Colors.white : Colors.black,
        ),
      ),
      const SizedBox(height: 16),
      _sectionHeader('الموافقة على الشروط'),
      _sectionBody(
        'باستخدامك التطبيق، فإنك توافق على الالتزام بهذه الشروط والقوانين المعمول بها في موريتانيا.',
      ),
      const SizedBox(height: 16),
      _sectionHeader('الأهلية'),
      _sectionBody(
        'يحق فقط للأشخاص الذين بلغوا 18 سنة أو أكثر المشاركة في المزادات.',
      ),
      const SizedBox(height: 16),
      _sectionHeader('الحساب والأمان'),
      _sectionBody('يجب تسجيل الحساب باستخدام رقم هاتف صحيح وفعال.'),
      _sectionBody(
        'أنت مسؤول عن سرية كلمة المرور وجميع الأنشطة التي تتم بحسابك.',
      ),
    ];
  }

  List<Widget> _buildPage2() {
    return [
      _sectionHeader('التعديلات والإشعارات:'),
      _bulletPoint('يحق لنا تعديل الشروط أو إضافة مزايا جديدة في أي وقت.'),
      _bulletPoint('أي تغييرات مهمة سنرسل إشعاراً للمستخدمين.'),
      const SizedBox(height: 16),
      _sectionHeader('1. إنهاء الحساب:'),
      _bulletPoint(
        'نحتفظ بحق تعليق أو حذف أي حساب ينتهك الشروط أو يضر بمستخدمي التطبيق.',
      ),
      const SizedBox(height: 16),
      _sectionHeader('2. القانون الواجب التطبيق:'),
      _bulletPoint(
        'تخضع هذه الشروط لقوانين الجمهورية الإسلامية الموريتانية، وأي نزاع يتم حله وفقاً لها.',
      ),
      const SizedBox(height: 16),
      _sectionHeader('شروط المشاركة في المزادات:'),
      _bulletPoint('يتطلب بعض المزادات دفع مبلغ تأمين لضمان جدية المزايدة.'),
      _bulletPoint(
        'يُسترجع مبلغ التأمين تلقائيًا في حال عدم الفوز بالمزاد خلال ساعة من انتهاء المزاد.',
      ),
      _bulletPoint(
        'لا يُسترجع مبلغ التأمين في حال الفوز بالمزاد وعدم إتمام عملية الدفع أو عند مخالفة شروط التطبيق.',
      ),
      _bulletPoint(
        'يحق لإدارة التطبيق إلغاء أي مزاد في حال وجود خطأ تقني أو شبهة احتيال، ويتم في هذه الحالة إعادة مبلغ التأمين للمشاركين.',
      ),
      _bulletPoint(
        'يشترط للاشتراك في التطبيق دفع رسوم اشتراك سنوية قدرها 100 أوقية جديدة للاستفادة من خدمات المزاد والمشاركة فيه',
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
