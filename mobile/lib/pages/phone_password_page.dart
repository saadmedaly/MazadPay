import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import '../services/auth_api.dart';
import '../models/api_response.dart';
import 'login_page.dart';
import '../services/auth_service.dart';
import '../services/category_api.dart';
import '../utils/error_mapper.dart';
import '../utils/phone_validation.dart';
import '../widgets/country_picker_sheet.dart';

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
    _confirmPasswordController.addListener(() => setState(() {}));
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

  String? _validateFields() {
    final l10n = AppLocalizations.of(context)!;
    final name = _nameController.text.trim();
    final phone = _phoneController.text.trim();
    final password = _passwordController.text;
    final confirmPassword = _confirmPasswordController.text;

    if (name.isEmpty || phone.isEmpty || password.isEmpty) {
      return l10n.error_fill_required_fields;
    }

    if (_selectedCountry == null) {
      return l10n.error_country_required;
    }

    if (!isPhoneLengthValid(phone, _selectedCountry)) {
      return l10n.error_phone_length;
    }

    if (password.length < 8) {
      return l10n.error_password_too_short;
    }

    if (password != confirmPassword) {
      return l10n.error_password_mismatch;
    }

    return null;
  }

  bool get _isFormValid {
    final name = _nameController.text.trim();
    final phone = _phoneController.text.trim();
    final password = _passwordController.text;
    final confirmPassword = _confirmPasswordController.text;
    if ((widget.fullName == null && name.isEmpty) ||
        phone.isEmpty ||
        password.isEmpty ||
        _selectedCountry == null) {
      return false;
    }
    if (!isPhoneLengthValid(phone, _selectedCountry)) return false;
    if (password.length < 8) return false;
    if (password != confirmPassword) return false;
    return true;
  }

  Future<void> _register() async {
    if (_isLoading) return;
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
    final countryIso = _selectedCountry!['code'].toString();
    final formattedPhone = '$countryCode$phone';

    setState(() => _isLoading = true);

    try {
      final ApiResponse<Map<String, dynamic>> response = await _authApi
          .register(
            phone: formattedPhone,
            pin: password,
            fullName: name,
            countryIso: countryIso,
            countryCode: countryCode,
          );

      if (response.success && mounted) {
        await AuthService().saveHasRegistered(true);
        Navigator.pushReplacement(
          context,
          MaterialPageRoute(
            // Transmet le numéro local et le country_code choisis au register
            // (Mobile Auth Phase 2) — jamais le numéro déjà préfixé, pour que
            // LoginPage recompose exactement le même format que celui envoyé
            // à /auth/register.
            builder: (context) =>
                LoginPage(initialPhone: phone, initialCountryCode: countryCode),
          ),
        );
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(AppLocalizations.of(context)!.profile_created),
            backgroundColor: Colors.green,
          ),
        );
      } else if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              mapError(
                context,
                response.error?.code,
                response.error?.message ?? '',
              ),
            ),
            backgroundColor: Colors.red,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(AppLocalizations.of(context)!.error_connection),
            backgroundColor: Colors.red,
          ),
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
                      Text(
                        'M',
                        style: const TextStyle(
                          fontSize: 60,
                          fontWeight: FontWeight.w900,
                          fontStyle: FontStyle.italic,
                          color: Colors.white,
                        ),
                      ),
                      Text(
                        'azad Pay',
                        style: const TextStyle(
                          fontSize: 40,
                          fontWeight: FontWeight.w900,
                          color: Colors.white,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    widget.fullName ?? '',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w500,
                      color: Colors.white.withOpacity(0.9),
                    ),
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
                    Text(
                      l10n.text_146,
                      style: const TextStyle(
                        fontSize: 22,
                        fontWeight: FontWeight.bold,
                        color: Colors.black,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      l10n.text_147,
                      style: const TextStyle(
                        fontSize: 14,
                        color: Color(0xFF9E9E9E),
                      ),
                    ),
                    const SizedBox(height: 32),

                    if (widget.fullName == null)
                      _buildField(
                        controller: _nameController,
                        hint: l10n.text_148,
                        icon: Icons.person_outline,
                      ),
                    if (widget.fullName == null) const SizedBox(height: 16),

                    _buildPhoneField(),

                    if (_phoneController.text.isNotEmpty &&
                        !isPhoneLengthValid(
                          _phoneController.text,
                          _selectedCountry,
                        ))
                      Padding(
                        padding: const EdgeInsets.only(top: 4, right: 8),
                        child: Align(
                          alignment: AlignmentDirectional.centerEnd,
                          child: Text(
                            l10n.error_phone_length,
                            style: const TextStyle(
                              color: Colors.red,
                              fontSize: 11,
                            ),
                          ),
                        ),
                      ),

                    const SizedBox(height: 20),

                    _buildField(
                      controller: _passwordController,
                      hint: 'أدخل كلمة السر',
                      icon: Icons.lock_outline,
                      isPassword: true,
                      obscureText: _obscurePassword,
                      onToggle: () =>
                          setState(() => _obscurePassword = !_obscurePassword),
                    ),
                    Padding(
                      padding: const EdgeInsets.only(top: 4, right: 8),
                      child: Align(
                        alignment: AlignmentDirectional.centerEnd,
                        child: Text(
                          _passwordController.text.isNotEmpty &&
                                  _passwordController.text.length < 8
                              ? l10n.error_password_too_short
                              : l10n.hint_password_min_length,
                          style: TextStyle(
                            color: _passwordController.text.isNotEmpty &&
                                    _passwordController.text.length < 8
                                ? Colors.red
                                : const Color(0xFF9E9E9E),
                            fontSize: 11,
                          ),
                        ),
                      ),
                    ),

                    const SizedBox(height: 20),

                    _buildField(
                      controller: _confirmPasswordController,
                      hint: 'أكد كلمة السر',
                      icon: Icons.lock_outline,
                      isPassword: true,
                      obscureText: _obscureConfirmPassword,
                      onToggle: () => setState(
                        () =>
                            _obscureConfirmPassword = !_obscureConfirmPassword,
                      ),
                    ),

                    if (_confirmPasswordController.text.isNotEmpty &&
                        _confirmPasswordController.text !=
                            _passwordController.text)
                      Padding(
                        padding: const EdgeInsets.only(top: 4, right: 8),
                        child: Align(
                          alignment: AlignmentDirectional.centerEnd,
                          child: Text(
                            l10n.error_password_mismatch,
                            style: const TextStyle(
                              color: Colors.red,
                              fontSize: 11,
                            ),
                          ),
                        ),
                      ),

                    const SizedBox(height: 32),

                    SizedBox(
                      width: double.infinity,
                      height: 50,
                      child: ElevatedButton(
                        onPressed: (_isLoading || !_isFormValid)
                            ? null
                            : _register,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: const Color(0xFF135BEC),
                          elevation: 0,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(10),
                          ),
                        ),
                        child: _isLoading
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(
                                  color: Colors.white,
                                  strokeWidth: 2,
                                ),
                              )
                            : Text(
                                l10n.text_217,
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 16,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                      ),
                    ),

                    const SizedBox(height: 20),

                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Flexible(
                          child: TextButton(
                            onPressed: () => Navigator.pushReplacement(
                              context,
                              MaterialPageRoute(
                                builder: (context) => const LoginPage(),
                              ),
                            ),
                            child: const Text(
                              'احجز مزادك بسرعة',
                              textAlign: TextAlign.center,
                              style: TextStyle(
                                color: Color(0xFF135BEC),
                                fontWeight: FontWeight.bold,
                                fontSize: 14,
                              ),
                            ),
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
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF135BEC).withOpacity(0.15),
            blurRadius: 10,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: TextField(
        controller: controller,
        obscureText: obscureText,
        textAlign: TextAlign.right,
        style: const TextStyle(
          fontWeight: FontWeight.w600,
          color: Colors.black,
          fontSize: 15,
        ),
        keyboardType:
            isPassword ? TextInputType.visiblePassword : TextInputType.text,
        maxLength: isPassword ? 72 : null,
        decoration: InputDecoration(
          counterText: "",
          hintText: hint,
          hintStyle: const TextStyle(color: Color(0xFF9E9E9E), fontSize: 14),
          prefixIcon: Icon(icon, color: const Color(0xFF135BEC)),
          suffixIcon: isPassword
              ? IconButton(
                  icon: Icon(
                    obscureText ? Icons.visibility_off : Icons.visibility,
                    color: const Color(0xFF135BEC),
                  ),
                  onPressed: onToggle,
                )
              : null,
          contentPadding: const EdgeInsets.symmetric(
            horizontal: 16,
            vertical: 18,
          ),
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
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF135BEC).withOpacity(0.15),
            blurRadius: 10,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: Row(
        children: [
          InkWell(
            onTap: () => _showCountryPicker(context),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 16),
              decoration: BoxDecoration(
                border: Border(
                  right: BorderSide(
                    color: const Color(0xFF135BEC).withOpacity(0.3),
                    width: 1,
                  ),
                ),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  ClipRRect(
                    borderRadius: BorderRadius.circular(2),
                    child: _selectedCountry?['code'] != null
                        ? Image.network(
                            'https://flagcdn.com/w80/${_selectedCountry!['code'].toString().toLowerCase()}.png',
                            width: 24,
                            height: 16,
                            fit: BoxFit.cover,
                            errorBuilder: (context, error, stack) =>
                                Image.asset(
                                  'assets/mr.png',
                                  width: 24,
                                  height: 16,
                                  fit: BoxFit.cover,
                                ),
                          )
                        : Image.asset(
                            'assets/mr.png',
                            width: 24,
                            height: 16,
                            fit: BoxFit.cover,
                          ),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    _selectedCountry?['country_code'] ?? '+222',
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 13,
                      color: Color(0xFF135BEC),
                    ),
                    textDirection: TextDirection.ltr,
                  ),
                  const Icon(
                    Icons.keyboard_arrow_down,
                    color: Color(0xFF135BEC),
                    size: 18,
                  ),
                ],
              ),
            ),
          ),
          Expanded(
            child: TextField(
              controller: _phoneController,
              keyboardType: TextInputType.phone,
              textAlign: TextAlign.right,
              maxLength: phoneMaxLengthFor(_selectedCountry),
              style: const TextStyle(
                fontWeight: FontWeight.w600,
                color: Colors.black,
                fontSize: 15,
              ),
              decoration: InputDecoration(
                hintText: AppLocalizations.of(context)!.text_213,
                hintStyle: const TextStyle(
                  color: Color(0xFF9E9E9E),
                  fontSize: 14,
                ),
                prefixIcon: const Icon(
                  Icons.phone_outlined,
                  color: Color(0xFF135BEC),
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 18,
                ),
                border: InputBorder.none,
                counterText: '',
              ),
            ),
          ),
        ],
      ),
    );
  }

  void _showCountryPicker(BuildContext context) {
    showCountryPickerSheet(
      context,
      countries: _countries,
      onSelected: (country) => setState(() => _selectedCountry = country),
    );
  }
}
