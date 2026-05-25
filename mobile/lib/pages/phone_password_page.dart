import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import '../services/auth_api.dart';
import '../models/api_response.dart';
import 'package:flutter/services.dart';
import 'login_page.dart';
import '../services/auth_service.dart';
import '../services/category_api.dart';
import '../utils/error_mapper.dart';

class PhonePasswordPage extends StatefulWidget {
  final String? fullName;

  const PhonePasswordPage({super.key, this.fullName});

  @override
  State<PhonePasswordPage> createState() => _PhonePasswordPageState();
}

class _PhonePasswordPageState extends State<PhonePasswordPage> {
  final _nameController = TextEditingController();
  final _phoneController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  final AuthApi _authApi = AuthApi();
  final CategoryApi _categoryApi = CategoryApi();
  bool _isLoading = false;
  bool _obscurePassword = true;
  bool _obscureConfirmPassword = true;
  List<dynamic> _countries = [];
  Map<String, dynamic>? _selectedCountry;

  @override
  void initState() {
    super.initState();
    if (widget.fullName != null) {
      _nameController.text = widget.fullName!;
    }
    _phoneController.addListener(() => setState(() {}));
    _passwordController.addListener(() => setState(() {}));
    _loadCountries();
  }

  Future<void> _loadCountries() async {
    final response = await _categoryApi.getCountries();
    if (response.success && response.data != null) {
      if (mounted) {
        setState(() {
          _countries = response.data!;
          try {
            _selectedCountry = _countries.firstWhere(
              (c) => c['country_code'] == '+222' || c['code'] == 'MR',
            );
          } catch (e) {
            if (_countries.isNotEmpty) {
              _selectedCountry = _countries.first;
            }
          }
        });
      }
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    _phoneController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    super.dispose();
  }

  bool get _isMauritania => _selectedCountry?['country_code'] == '+222';

  String? _validateFields() {
    final l10n = AppLocalizations.of(context)!;
    final name = _nameController.text.trim();
    final phone = _phoneController.text.trim();
    final password = _passwordController.text;
    final confirmPassword = _confirmPasswordController.text;

    if (name.isEmpty || phone.isEmpty || password.isEmpty) {
      return l10n.error_fill_required_fields;
    }

    if (_isMauritania) {
      if (phone.length != 8) return l10n.error_phone_length;
      if (!RegExp(r'^[234]').hasMatch(phone)) return l10n.error_phone_start_digit;
    }

    if (password.length < 4) {
      return l10n.error_password_too_short;
    }

    if (password != confirmPassword) {
      return l10n.error_password_mismatch;
    }

    return null;
  }

  Future<void> _register() async {
    final error = _validateFields();
    if (error != null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(error), backgroundColor: Colors.red),
      );
      return;
    }

    final name = _nameController.text.trim();
    final phone = _phoneController.text.trim();
    final password = _passwordController.text;
    final countryCode = _selectedCountry?['country_code'] ?? '+222';
    final formattedPhone = '$countryCode$phone';

    setState(() => _isLoading = true);

    try {
      final ApiResponse<Map<String, dynamic>> response = await _authApi.register(
        phone: formattedPhone,
        pin: password,
        fullName: name,
        countryCode: countryCode,
      );

      if (response.success && mounted) {
        await AuthService().saveHasRegistered(true);
        Navigator.pushReplacement(
          context,
          MaterialPageRoute(builder: (context) => LoginPage()),
        );
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.profile_created), backgroundColor: Colors.green),
        );
      } else if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(mapError(context, response.error?.code, response.error?.message ?? '')),
            backgroundColor: Colors.red,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.error_connection), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: const Color(0xFF135BEC),
      body: Column(
        children: [
          Expanded(
            flex: 3,
            child: Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    textDirection: TextDirection.ltr,
                    children: [
                      Text('M', style: const TextStyle(fontSize: 60, fontWeight: FontWeight.w900, fontStyle: FontStyle.italic, color: Colors.white)),
                      Text('azad Pay', style: const TextStyle(fontSize: 40, fontWeight: FontWeight.w900, color: Colors.white)),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    widget.fullName ?? '',
                    style: TextStyle(fontSize: 18, fontWeight: FontWeight.w500, color: Colors.white.withOpacity(0.9)),
                  ),
                ],
              ),
            ),
          ),

          Expanded(
            flex: 7,
            child: Container(
              width: double.infinity,
              decoration: const BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadiusDirectional.only(
                  topStart: Radius.circular(30),
                  topEnd: Radius.circular(30),
                ),
              ),
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
              child: SingleChildScrollView(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: [
                    const SizedBox(height: 16),
                    Text(l10n.text_146, style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold, color: Colors.black)),
                    const SizedBox(height: 8),
                    Text(l10n.text_147, style: const TextStyle(fontSize: 14, color: Color(0xFF9E9E9E))),
                    const SizedBox(height: 32),

                    if (widget.fullName == null)
                      _buildField(
                        controller: _nameController,
                        hint: l10n.text_148,
                        icon: Icons.person_outline,
                      ),
                    if (widget.fullName == null) const SizedBox(height: 16),

                    _buildPhoneField(),

                    if (_isMauritania && _phoneController.text.isNotEmpty && _phoneController.text.trim().length == 8 && !RegExp(r'^[234]').hasMatch(_phoneController.text.trim()))
                      Padding(
                        padding: const EdgeInsets.only(top: 4, right: 8),
                        child: Align(
                          alignment: AlignmentDirectional.centerEnd,
                          child: Text(l10n.error_phone_start_digit, style: const TextStyle(color: Colors.red, fontSize: 11)),
                        ),
                      ),

                    const SizedBox(height: 20),

                    _buildField(
                      controller: _passwordController,
                      hint: l10n.text_5,
                      icon: Icons.lock_outline,
                      isPassword: true,
                      obscureText: _obscurePassword,
                      onToggle: () => setState(() => _obscurePassword = !_obscurePassword),
                    ),

                    const SizedBox(height: 20),

                    _buildField(
                      controller: _confirmPasswordController,
                      hint: l10n.text_7,
                      icon: Icons.lock_outline,
                      isPassword: true,
                      obscureText: _obscureConfirmPassword,
                      onToggle: () => setState(() => _obscureConfirmPassword = !_obscureConfirmPassword),
                    ),

                    const SizedBox(height: 32),

                    SizedBox(
                      width: double.infinity,
                      height: 50,
                      child: ElevatedButton(
                        onPressed: _isLoading ? null : _register,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: const Color(0xFF135BEC),
                          elevation: 0,
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
                        ),
                        child: _isLoading
                          ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
                          : Text(l10n.text_217, style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold)),
                      ),
                    ),

                    const SizedBox(height: 20),

                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Flexible(
                          child: TextButton(
                            onPressed: () => Navigator.pushReplacement(context, MaterialPageRoute(builder: (context) => const LoginPage())),
                            child: Text(l10n.text_11, textAlign: TextAlign.center, overflow: TextOverflow.ellipsis,
                              style: const TextStyle(color: Color(0xFF135BEC), fontWeight: FontWeight.bold, fontSize: 14)),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildField({
    required TextEditingController controller,
    required String hint,
    required IconData icon,
    bool isPassword = false,
    bool obscureText = false,
    VoidCallback? onToggle,
  }) {
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFEEF4FF),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFF135BEC), width: 1.8),
        boxShadow: [BoxShadow(color: const Color(0xFF135BEC).withOpacity(0.15), blurRadius: 10, offset: const Offset(0, 3))],
      ),
      child: TextField(
        controller: controller,
        obscureText: obscureText,
        textAlign: TextAlign.right,
        style: const TextStyle(fontWeight: FontWeight.w600, color: Colors.black, fontSize: 15),
        keyboardType: isPassword ? TextInputType.number : TextInputType.text,
        maxLength: isPassword ? 4 : null,
        inputFormatters: isPassword ? [FilteringTextInputFormatter.digitsOnly] : null,
        decoration: InputDecoration(
          counterText: "",
          hintText: hint,
          hintStyle: const TextStyle(color: Color(0xFF9E9E9E), fontSize: 14),
          prefixIcon: Icon(icon, color: const Color(0xFF135BEC)),
          suffixIcon: isPassword
            ? IconButton(
                icon: Icon(obscureText ? Icons.visibility_off : Icons.visibility, color: const Color(0xFF135BEC)),
                onPressed: onToggle,
              )
            : null,
          contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 18),
          border: InputBorder.none,
        ),
      ),
    );
  }

  Widget _buildPhoneField() {
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFEEF4FF),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFF135BEC), width: 1.8),
        boxShadow: [BoxShadow(color: const Color(0xFF135BEC).withOpacity(0.15), blurRadius: 10, offset: const Offset(0, 3))],
      ),
      child: Row(
        children: [
          InkWell(
            onTap: () => _showCountryPicker(context),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 16),
              decoration: BoxDecoration(
                border: Border(right: BorderSide(color: const Color(0xFF135BEC).withOpacity(0.3), width: 1)),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  ClipRRect(
                    borderRadius: BorderRadius.circular(2),
                    child: _selectedCountry?['code'] != null
                      ? Image.network(
                          'https://flagcdn.com/w80/${_selectedCountry!['code'].toString().toLowerCase()}.png',
                          width: 24, height: 16, fit: BoxFit.cover,
                          errorBuilder: (context, error, stack) => Image.asset('assets/mr.png', width: 24, height: 16, fit: BoxFit.cover),
                        )
                      : Image.asset('assets/mr.png', width: 24, height: 16, fit: BoxFit.cover),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    _selectedCountry?['country_code'] ?? '+222',
                    style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Color(0xFF135BEC)),
                    textDirection: TextDirection.ltr,
                  ),
                  const Icon(Icons.keyboard_arrow_down, color: Color(0xFF135BEC), size: 18),
                ],
              ),
            ),
          ),
          Expanded(
            child: TextField(
              controller: _phoneController,
              keyboardType: TextInputType.phone,
              textAlign: TextAlign.right,
              style: const TextStyle(fontWeight: FontWeight.w600, color: Colors.black, fontSize: 15),
              decoration: InputDecoration(
                hintText: AppLocalizations.of(context)!.text_213,
                hintStyle: const TextStyle(color: Color(0xFF9E9E9E), fontSize: 14),
                prefixIcon: const Icon(Icons.phone_outlined, color: Color(0xFF135BEC)),
                contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 18),
                border: InputBorder.none,
              ),
            ),
          ),
        ],
      ),
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
            borderRadius: const BorderRadiusDirectional.only(topStart: Radius.circular(24), topEnd: Radius.circular(24)),
          ),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(width: 40, height: 4, decoration: BoxDecoration(color: isDarkMode ? Colors.grey[700] : Colors.grey[300], borderRadius: BorderRadius.circular(2))),
              const SizedBox(height: 24),
              Text(AppLocalizations.of(context)!.text_219, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
              const SizedBox(height: 24),
              if (_countries.isEmpty)
                const CircularProgressIndicator()
              else
                ..._countries.map((country) {
                  final isArabic = Localizations.localeOf(context).languageCode == 'ar';
                  final isFrench = Localizations.localeOf(context).languageCode == 'fr';
                  final name = isArabic ? country['name_ar'] : (isFrench ? country['name_fr'] : country['name_en']);
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 12.0),
                    child: InkWell(
                      onTap: () {
                        if (country['is_active'] ?? true) {
                          setState(() => _selectedCountry = country);
                          Navigator.pop(context);
                        }
                      },
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(color: (country['is_active'] ?? true) ? const Color(0xFF135BEC).withOpacity(0.3) : Colors.grey.withOpacity(0.2)),
                        ),
                        child: Row(
                          children: [
                            ClipRRect(
                              borderRadius: BorderRadius.circular(2),
                              child: Image.network('https://flagcdn.com/w80/${country['code'].toString().toLowerCase()}.png',
                                width: 32, height: 20, fit: BoxFit.cover,
                                errorBuilder: (context, error, stackTrace) => Container(width: 32, height: 20, color: Colors.grey[300], child: const Icon(Icons.flag, size: 12)),
                              ),
                            ),
                            const SizedBox(width: 16),
                            Text(name ?? 'Unknown', style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                            const Spacer(),
                            Text(country['country_code'] ?? '', style: const TextStyle(fontSize: 16, color: Colors.grey, fontWeight: FontWeight.w500), textDirection: TextDirection.ltr),
                          ],
                        ),
                      ),
                    ),
                  );
                }),
              const SizedBox(height: 40),
            ],
          ),
        );
      },
    );
  }
}
