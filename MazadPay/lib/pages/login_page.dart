import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'login_controller.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'phone_registration_page.dart';
import 'language_page.dart';
import '../widgets/mazad_pay_logo.dart';
import 'new_password_page.dart';
import 'home_page.dart';
import 'otp_entry_page.dart';
import '../widgets/success_dialog.dart';

class LoginPage extends ConsumerStatefulWidget {
  final bool showSuccessDialog;
  
  const LoginPage({super.key, this.showSuccessDialog = false});

  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage> {
  final _phoneController = TextEditingController();
  final _passwordController = TextEditingController();
  bool _obscurePassword = true;

  @override
  void initState() {
    super.initState();
    _phoneController.addListener(() {
      setState(() {});
    });
    _passwordController.addListener(() {
      setState(() {});
    });

    if (widget.showSuccessDialog) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        showSuccessDialog(
          context,
          onBack: () {
            Navigator.pushReplacement(
              context,
              MaterialPageRoute(builder: (context) => const HomePage()),
            );
          },
        );
      });
    }
  }

  @override
  void dispose() {
    _phoneController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
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
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (context) => const LanguagePage()),
                );
              },
            ),
          ),
          actions: [
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
                onPressed: () => Navigator.of(context).pop(),
              ),
            ),
          ],
        ),
        body: SingleChildScrollView(
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24),
            child: Column(
              children: [
                // Logo & Branding
                const MazadPayLogo(fontSize: 42, arabicFontSize: 24),

                const SizedBox(height: 20),

                // Form Fields
                _buildInputField(
                  controller: _phoneController,
                  label: 'رقم الهاتف',
                  hint: 'أدخل رقم هاتفك',
                  keyboardType: TextInputType.phone,
                  counter: '${_phoneController.text.length}/8',
                  maxLength: 8,
                ),
                const SizedBox(height: 8),
                _buildInputField(
                  controller: _passwordController,
                  label: 'كلمة السر',
                  hint: '• • • •',
                  isPassword: true,
                  obscureText: _obscurePassword,
                  onToggleVisibility: () {
                    setState(() {
                      _obscurePassword = !_obscurePassword;
                    });
                  },
                  counter: '${_passwordController.text.length}/4',
                  maxLength: 4,
                  keyboardType: TextInputType.number,
                  letterSpacing: 24, // High spacing for digits as in screenshot
                ),

                Align(
                  alignment: Alignment.centerRight,
                  child: TextButton(
                    onPressed: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => const NewPasswordPage(),
                        ),
                      );
                    },
                    child: const Text(
                      'هل نسيت كلمة المرور؟',
                      style: TextStyle(color: Color(0xFF667085)),
                    ),
                  ),
                ),
                const SizedBox(height: 10),
                GestureDetector(
                  onTap: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => const PhoneRegistrationPage(),
                      ),
                    );
                  },
                  child: const Text(
                    'مستخدم جديد؟ سجل الآن!',
                    style: TextStyle(
                      color: Color(0xFF667085),
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
                const SizedBox(height: 10), // Reduced space for bottom button
              ],
            ),
          ),
        ),
        bottomNavigationBar: SafeArea(
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                SizedBox(
                  width: double.infinity,
                  height: 48,
                  child: ElevatedButton(
                    onPressed: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => OtpEntryPage(
                            phoneNumber: _phoneController.text,
                          ),
                        ),
                      );
                    },
                    child: ref.watch(loginControllerProvider).isLoading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2),
                        )
                      : const Text('تسجيل الدخول'),
                  ),
                ),
                TextButton(
                  onPressed: () => _showContactInfo(context),
                  style: TextButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 4),
                    minimumSize: Size.zero,
                    tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  ),
                  child: const Text(
                    'اتصل بنا',
                    style: TextStyle(
                      color: Color(0xFF135BEC),
                      fontWeight: FontWeight.bold,
                      decoration: TextDecoration.underline,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildInputField({
    required TextEditingController controller,
    required String label,
    required String hint,
    String? counter,
    bool isPassword = false,
    bool obscureText = false,
    VoidCallback? onToggleVisibility,
    TextInputType keyboardType = TextInputType.text,
    int? maxLength,
    double? letterSpacing,
  }) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    bool isPhone = keyboardType == TextInputType.phone;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.end,
      children: [
        Text(
          label,
          style: TextStyle(
            fontWeight: FontWeight.bold,
            fontSize: 16,
            color: isDarkMode ? Colors.white : Colors.black,
          ),
        ),
        const SizedBox(height: 8),
        TextField(
          controller: controller,
          obscureText: obscureText,
          textAlign: TextAlign.right,
          keyboardType: keyboardType,
          style: TextStyle(
            fontWeight: FontWeight.w500,
            color: isDarkMode ? Colors.white : Colors.black,
            letterSpacing: letterSpacing,
          ),
          maxLength: maxLength,
          decoration: InputDecoration(
            hintText: hint,
            hintStyle: const TextStyle(color: Color(0xFF98A2B3), fontSize: 14, letterSpacing: 0),
            counterText: "",
            suffixIcon: isPassword
                ? IconButton(
                    icon: Icon(
                      obscureText
                          ? Icons.visibility_off_outlined
                          : Icons.visibility_outlined,
                      color: const Color(0xFF98A2B3),
                    ),
                    onPressed: onToggleVisibility,
                  )
                : null,
            prefixIcon: isPhone
                ? InkWell(
                    onTap: () => _showCountryPicker(context),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12),
                      decoration: BoxDecoration(
                        border: Border(
                          right: BorderSide( // Swapping border to right for prefix
                            color: isDarkMode
                                ? const Color(0xFF333333)
                                : const Color(0xFFF2F4F7),
                            width: 1.5,
                          ),
                        ),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          ClipRRect(
                            borderRadius: BorderRadius.circular(2),
                            child: Image.asset(
                              'assets/mr.png',
                              width: 24,
                              height: 16,
                              fit: BoxFit.cover,
                            ),
                          ),
                          const SizedBox(width: 8),
                          const Text(
                            '+222',
                            style: TextStyle(
                              fontWeight: FontWeight.bold,
                              fontSize: 14,
                              color: Color(0xFF475467),
                            ),
                            textDirection: TextDirection.ltr,
                          ),
                          const SizedBox(width: 4),
                          Icon(
                            Icons.keyboard_arrow_down,
                            color: Colors.grey[400],
                            size: 18,
                          ),
                        ],
                      ),
                    ),
                  )
                : null,
            filled: true,
            fillColor: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 16,
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(
                color: isDarkMode
                    ? const Color(0xFF333333)
                    : const Color(0xFFF2F4F7),
                width: 1.5,
              ),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(
                color: Theme.of(context).primaryColor,
                width: 1.5,
              ),
            ),
          ),
        ),
        const SizedBox(height: 4),
        Text(
          counter ?? '',
          style: const TextStyle(color: Color(0xFF98A2B3), fontSize: 12),
        ),
      ],
    );
  }

  void _showCountryPicker(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) {
        return Container(
          decoration: BoxDecoration(
            color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
            borderRadius: const BorderRadius.only(
              topLeft: Radius.circular(24),
              topRight: Radius.circular(24),
            ),
          ),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: isDarkMode ? Colors.grey[700] : Colors.grey[300],
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              const SizedBox(height: 24),
              const Text(
                'اختر الدولة',
                style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 24),
              _buildCountryItem(
                context,
                name: 'موريتانيا',
                code: '+222',
                flagUrl: 'assets/mr.png',
                isAvailable: true,
              ),
              const Divider(height: 32),
              const Align(
                alignment: Alignment.centerRight,
                child: Text(
                  'قريباً',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: Colors.grey,
                  ),
                ),
              ),
              const SizedBox(height: 16),
              _buildCountryItem(
                context,
                name: 'السنغال',
                code: '+221',
                flagUrl: 'https://flagcdn.com/w80/sn.png',
                isAvailable: false,
              ),
              const SizedBox(height: 12),
              _buildCountryItem(
                context,
                name: 'المغرب',
                code: '+212',
                flagUrl: 'https://flagcdn.com/w80/ma.png',
                isAvailable: false,
              ),
              const SizedBox(height: 12),
              _buildCountryItem(
                context,
                name: 'تونس',
                code: '+216',
                flagUrl: 'https://flagcdn.com/w80/tn.png',
                isAvailable: false,
              ),
              const SizedBox(height: 40),
            ],
          ),
        );
      },
    );
  }

  Widget _buildCountryItem(
    BuildContext context, {
    required String name,
    required String code,
    required String flagUrl,
    required bool isAvailable,
  }) {
    return Opacity(
      opacity: isAvailable ? 1.0 : 0.5,
      child: InkWell(
        onTap: isAvailable ? () => Navigator.pop(context) : null,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: isAvailable
                    ? const Color(0xFF135BEC).withOpacity(0.3)
                    : Colors.grey.withOpacity(0.2),
              ),
          ),
          child: Row(
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(2),
                child: flagUrl.startsWith('http')
                    ? Image.network(
                        flagUrl,
                        width: 32,
                        height: 20,
                        fit: BoxFit.cover,
                        errorBuilder: (context, error, stackTrace) => Container(
                          width: 32,
                          height: 20,
                          color: Colors.grey[300],
                          child: const Icon(Icons.flag, size: 12),
                        ),
                      )
                    : Image.asset(
                        flagUrl,
                        width: 32,
                        height: 20,
                        fit: BoxFit.cover,
                      ),
              ),
              const SizedBox(width: 16),
              Text(
                name,
                style: const TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const Spacer(),
              Text(
                code,
                style: const TextStyle(
                  fontSize: 16,
                  color: Colors.grey,
                  fontWeight: FontWeight.w500,
                ),
                textDirection: TextDirection.ltr,
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showContactInfo(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      builder: (context) {
        return Container(
          decoration: BoxDecoration(
            color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
            borderRadius: const BorderRadius.only(
              topLeft: Radius.circular(24),
              topRight: Radius.circular(24),
            ),
          ),
          padding: const EdgeInsets.only(
            left: 24,
            right: 24,
            top: 16,
            bottom: 24,
          ),
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: isDarkMode ? Colors.grey[700] : Colors.grey[300],
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
                const SizedBox(height: 24),
                const Text(
                  'تواصل معنا',
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    fontFamily: 'Plus Jakarta Sans',
                  ),
                ),
                const SizedBox(height: 24),
                _buildContactItem(
                  icon: FontAwesomeIcons.whatsapp,
                  label: 'واتس اب',
                  value: '47601175',
                  color: const Color(0xFF25D366),
                ),
                const SizedBox(height: 16),
                _buildContactItem(
                  icon: Icons.email_outlined,
                  label: 'البريد الإلكتروني',
                  value: 'mazadpay@gmail.com',
                  color: const Color(0xFF135BEC),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildContactItem({
    required IconData icon,
    required String label,
    required String value,
    required Color color,
  }) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isDarkMode ? const Color(0xFF262626) : const Color(0xFFF9FAFB),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: isDarkMode ? const Color(0xFF333333) : const Color(0xFFEAECF0),
        ),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: color.withOpacity(0.1),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Icon(icon, color: color, size: 24),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: TextStyle(
                    fontSize: 12,
                    color: isDarkMode
                        ? Colors.grey[400]
                        : const Color(0xFF667085),
                  ),
                ),
                Text(
                  value,
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: isDarkMode ? Colors.white : const Color(0xFF1D2939),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
