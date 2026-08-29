import 'dart:async';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'login_page.dart';
import '../services/auth_api.dart';
import '../services/category_api.dart';
import '../utils/error_mapper.dart';
import '../utils/phone_validation.dart';
import '../widgets/country_picker_sheet.dart';

class NewPasswordPage extends StatefulWidget {
  final String? initialPhone;
  final String? initialCountryCode;
  final String? initialCountryIso;

  const NewPasswordPage({
    super.key,
    this.initialPhone,
    this.initialCountryCode,
    this.initialCountryIso,
  });

  @override
  State<NewPasswordPage> createState() => _NewPasswordPageState();
}

class _NewPasswordPageState extends State<NewPasswordPage> {
  final AuthApi _authApi = AuthApi();
  final CategoryApi _categoryApi = CategoryApi();
  final _phoneController = TextEditingController();
  final List<TextEditingController> _otpControllers = List.generate(4, (_) => TextEditingController());
  final List<FocusNode> _otpFocusNodes = List.generate(4, (_) => FocusNode());
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  bool _obscurePassword = true;
  bool _obscureConfirmPassword = true;

  bool _isLoading = false;
  bool _otpSent = false;
  int _countdown = 0;
  Timer? _timer;
  List<dynamic> _countries = [];
  Map<String, dynamic>? _selectedCountry;

  String get _fullPhone =>
      '${_selectedCountry?['country_code'] ?? '+222'}${_phoneController.text.trim()}';

  @override
  void initState() {
    super.initState();
    if (widget.initialPhone != null) {
      _phoneController.text = widget.initialPhone!;
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
            if (widget.initialCountryIso != null) {
              _selectedCountry = _countries.firstWhere(
                (c) => c['code'] == widget.initialCountryIso,
              );
            } else if (widget.initialCountryCode != null) {
              _selectedCountry = _countries.firstWhere(
                (c) => c['country_code'] == widget.initialCountryCode,
              );
            } else {
              _selectedCountry = _countries.firstWhere(
                (c) => c['country_code'] == '+222' || c['code'] == 'MR',
              );
            }
          } catch (e) {
            if (_countries.isNotEmpty) {
              _selectedCountry = _countries.first;
            }
          }
        });
      }
    }
  }

  bool get _isPhoneValid =>
      isPhoneLengthValid(_phoneController.text, _selectedCountry);

  @override
  void dispose() {
    _timer?.cancel();
    _phoneController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    for (var c in _otpControllers) { c.dispose(); }
    for (var n in _otpFocusNodes) { n.dispose(); }
    super.dispose();
  }

  String get _formattedTime {
    final min = (_countdown ~/ 60).toString().padLeft(2, '0');
    final sec = (_countdown % 60).toString().padLeft(2, '0');
    return '$min:$sec';
  }

  void _startCountdown() {
    _timer?.cancel();
    setState(() => _countdown = 60);
    _timer = Timer.periodic(const Duration(seconds: 1), (t) {
      if (_countdown <= 1) { t.cancel(); setState(() => _countdown = 0); }
      else { setState(() => _countdown--); }
    });
  }

  Future<void> _sendOTP() async {
    final phone = _phoneController.text.trim();
    if (_selectedCountry == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_country_required)),
      );
      return;
    }
    if (!isPhoneLengthValid(phone, _selectedCountry)) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_phone_length)),
      );
      return;
    }

    setState(() => _isLoading = true);
    try {
      final response = await _authApi.sendOTP(phone: _fullPhone, purpose: 'reset_password');
      setState(() => _isLoading = false);
      if (response.success) {
        setState(() => _otpSent = true);
        _startCountdown();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.text_261)),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(mapError(context, response.error?.code, response.error?.message ?? ''))),
        );
      }
    } catch (e) {
      setState(() => _isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_connection)),
      );
    }
  }

  Future<void> _resetPassword() async {
    final otp = _otpControllers.map((c) => c.text).join();
    final pin = _passwordController.text;
    final confirm = _confirmPasswordController.text;

    if (otp.length != 4) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_otp_required)),
      );
      return;
    }
    if (pin.length < 8) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_password_too_short)),
      );
      return;
    }
    if (pin != confirm) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_password_mismatch)),
      );
      return;
    }

    setState(() => _isLoading = true);
    try {
      final response = await _authApi.resetPassword(
        phone: _fullPhone,
        newPin: pin,
        otpCode: otp,
      );
      setState(() => _isLoading = false);
      if (response.success) {
        Navigator.pushAndRemoveUntil(
          context,
          MaterialPageRoute(builder: (context) => const LoginPage(showSuccessDialog: true)),
          (route) => false,
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(mapError(context, response.error?.code, response.error?.message ?? ''))),
        );
      }
    } catch (e) {
      setState(() => _isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_connection)),
      );
    }
  }

  void _onOtpChanged(String value, int index) {
    if (value.length == 1 && index < 3) _otpFocusNodes[index + 1].requestFocus();
    if (value.isEmpty && index > 0) _otpFocusNodes[index - 1].requestFocus();
  }

  void _showCountryPicker() {
    showCountryPickerSheet(
      context,
      countries: _countries,
      onSelected: (country) => setState(() => _selectedCountry = country),
    );
  }

  Widget _buildPinRow(List<TextEditingController> controllers, List<FocusNode> nodes,
      void Function(String, int) onChanged) {
    return Directionality(
      textDirection: TextDirection.ltr,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: List.generate(4, (index) {
          return Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4),
            child: SizedBox(
              width: 50, height: 60,
              child: TextField(
                controller: controllers[index],
                focusNode: nodes[index],
                keyboardType: TextInputType.number,
                textAlign: TextAlign.center,
                maxLength: 1,
                style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                decoration: InputDecoration(
                  counterText: '',
                  contentPadding: EdgeInsets.zero,
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(color: Colors.grey.shade400, width: 1.5),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(color: Theme.of(context).primaryColor, width: 1.5),
                  ),
                ),
                onChanged: (value) => onChanged(value, index),
              ),
            ),
          );
        }),
      ),
    );
  }

  Widget _buildPasswordField({
    required TextEditingController controller,
    required bool obscure,
    required VoidCallback onToggle,
    required bool isDarkMode,
  }) {
    return TextField(
      controller: controller,
      obscureText: obscure,
      textAlign: TextAlign.right,
      maxLength: 72,
      style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: isDarkMode ? Colors.white : Colors.black),
      decoration: InputDecoration(
        counterText: '',
        filled: true,
        fillColor: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
        suffixIcon: IconButton(
          icon: Icon(obscure ? Icons.visibility_off_outlined : Icons.visibility_outlined, color: const Color(0xFF98A2B3)),
          onPressed: onToggle,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: isDarkMode ? const Color(0xFF333333) : const Color(0xFFF2F4F7), width: 1.5),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Theme.of(context).primaryColor, width: 1.5),
        ),
      ),
    );
  }

  bool get _isResetFormValid {
    final otp = _otpControllers.map((c) => c.text).join();
    final pin = _passwordController.text;
    final confirm = _confirmPasswordController.text;
    return otp.length == 4 && pin.length >= 8 && pin == confirm;
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;

    return Scaffold(
      backgroundColor: Theme.of(context).scaffoldBackgroundColor,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: true,
        leading: Padding(
          padding: const EdgeInsetsDirectional.only(end: 16.0),
          child: Center(
            child: InkWell(
              onTap: () => Navigator.pop(context),
              borderRadius: BorderRadius.circular(20),
              child: Container(
                width: 40, height: 40,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: isDarkMode ? Colors.grey[800] : Colors.grey[100],
                ),
                child: Icon(Icons.chevron_right, color: isDarkMode ? Colors.grey[300] : Colors.grey[600]),
              ),
            ),
          ),
        ),
        title: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(AppLocalizations.of(context)!.text_241, style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: isDarkMode ? Colors.grey[400] : Colors.grey[500])),
            Text(AppLocalizations.of(context)!.text_242, style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: isDarkMode ? Colors.grey[200] : Colors.grey[800])),
          ],
        ),
        actions: [
          Padding(
            padding: const EdgeInsetsDirectional.only(start: 16.0),
            child: Center(
              child: Container(
                width: 40, height: 40,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: isDarkMode ? Colors.grey[800] : Colors.grey[100],
                ),
                child: Icon(Icons.smart_toy_outlined, size: 20, color: isDarkMode ? Colors.grey[300] : Colors.grey[600]),
              ),
            ),
          ),
        ],
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              _buildIconIllustration(context),
              const SizedBox(height: 24),
              Text(
                AppLocalizations.of(context)!.text_215,
                style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: isDarkMode ? Colors.white : Colors.black),
              ),
              const SizedBox(height: 16),
              // Phone input with country code
              Row(
                children: [
                  InkWell(
                    onTap: _showCountryPicker,
                    borderRadius: BorderRadius.circular(12),
                    child: Container(
                      height: 56,
                      padding: const EdgeInsets.symmetric(horizontal: 12),
                      decoration: BoxDecoration(
                        color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                        border: Border.all(color: isDarkMode ? const Color(0xFF333333) : const Color(0xFFF2F4F7), width: 1.5),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            _selectedCountry?['country_code'] ?? '+222',
                            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                            textDirection: TextDirection.ltr,
                          ),
                          const SizedBox(width: 4),
                          Icon(Icons.keyboard_arrow_down, size: 16, color: Colors.grey[400]),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: TextField(
                      controller: _phoneController,
                      keyboardType: TextInputType.phone,
                      textAlign: TextAlign.center,
                      maxLength: phoneMaxLengthFor(_selectedCountry),
                      style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: isDarkMode ? Colors.white : Colors.black),
                      decoration: InputDecoration(
                        hintText: AppLocalizations.of(context)!.text_213,
                        counterText: '',
                        filled: true,
                        fillColor: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                        enabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                          borderSide: BorderSide(color: isDarkMode ? const Color(0xFF333333) : const Color(0xFFF2F4F7), width: 1.5),
                        ),
                        focusedBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                          borderSide: BorderSide(color: Theme.of(context).primaryColor, width: 1.5),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
              if (_phoneController.text.isNotEmpty && !_isPhoneValid)
                Padding(
                  padding: const EdgeInsets.only(top: 4),
                  child: Align(
                    alignment: Alignment.centerRight,
                    child: Text(
                      AppLocalizations.of(context)!.error_phone_length,
                      style: const TextStyle(color: Colors.red, fontSize: 11),
                    ),
                  ),
                ),
              const SizedBox(height: 16),
              if (!_otpSent)
                SizedBox(
                  width: double.infinity,
                  height: 48,
                  child: ElevatedButton(
                    onPressed: (_isLoading || _selectedCountry == null || !_isPhoneValid) ? null : _sendOTP,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF0081FF),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                      elevation: 0,
                    ),
                    child: _isLoading
                      ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
                      : Text(AppLocalizations.of(context)!.text_211, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white)),
                  ),
                ),
              if (_otpSent) ...[
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.chat_outlined, size: 18, color: const Color(0xFF25D366)),
                    const SizedBox(width: 6),
                    Flexible(
                      child: Text(AppLocalizations.of(context)!.text_261, style: TextStyle(fontSize: 13, color: isDarkMode ? Colors.grey[400] : Colors.grey[500])),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                _buildPinRow(_otpControllers, _otpFocusNodes, _onOtpChanged),
                const SizedBox(height: 12),
                if (_countdown > 0)
                  Text('${AppLocalizations.of(context)!.text_262} $_formattedTime', style: TextStyle(fontSize: 13, color: isDarkMode ? Colors.grey[400] : Colors.grey[500]))
                else
                  TextButton(onPressed: _sendOTP, child: Text(AppLocalizations.of(context)!.text_264, style: const TextStyle(color: Color(0xFF0081FF)))),
                const SizedBox(height: 24),
                Text(AppLocalizations.of(context)!.text_243, style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: isDarkMode ? Colors.white : Colors.black)),
                const SizedBox(height: 12),
                _buildPasswordField(
                  controller: _passwordController,
                  obscure: _obscurePassword,
                  onToggle: () => setState(() => _obscurePassword = !_obscurePassword),
                  isDarkMode: isDarkMode,
                ),
                Padding(
                  padding: const EdgeInsets.only(top: 4),
                  child: Align(
                    alignment: Alignment.centerRight,
                    child: Text(
                      _passwordController.text.isNotEmpty && _passwordController.text.length < 8
                          ? AppLocalizations.of(context)!.error_password_too_short
                          : AppLocalizations.of(context)!.hint_password_min_length,
                      style: TextStyle(
                        color: _passwordController.text.isNotEmpty && _passwordController.text.length < 8
                            ? Colors.red
                            : (isDarkMode ? Colors.grey[400] : Colors.grey[500]),
                        fontSize: 11,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 24),
                Text(AppLocalizations.of(context)!.text_244, style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: isDarkMode ? Colors.white : Colors.black)),
                const SizedBox(height: 12),
                _buildPasswordField(
                  controller: _confirmPasswordController,
                  obscure: _obscureConfirmPassword,
                  onToggle: () => setState(() => _obscureConfirmPassword = !_obscureConfirmPassword),
                  isDarkMode: isDarkMode,
                ),
                if (_confirmPasswordController.text.isNotEmpty &&
                    _confirmPasswordController.text != _passwordController.text)
                  Padding(
                    padding: const EdgeInsets.only(top: 4),
                    child: Align(
                      alignment: Alignment.centerRight,
                      child: Text(
                        AppLocalizations.of(context)!.error_password_mismatch,
                        style: const TextStyle(color: Colors.red, fontSize: 11),
                      ),
                    ),
                  ),
                const SizedBox(height: 32),
                SizedBox(
                  width: double.infinity,
                  height: 48,
                  child: ElevatedButton(
                    onPressed: (_isLoading || !_isResetFormValid) ? null : _resetPassword,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF0081FF),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                      elevation: 0,
                    ),
                    child: _isLoading
                      ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
                      : Text(AppLocalizations.of(context)!.text_211, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.white)),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildIconIllustration(BuildContext context) {
    return SizedBox(
      height: 120,
      width: double.infinity,
      child: Stack(
        alignment: Alignment.center,
        children: [
          Icon(Icons.refresh, size: 100, color: Colors.blue.withOpacity(0.1)),
          const Icon(Icons.lock_reset_rounded, size: 50, color: Color(0xFF0081FF)),
        ],
      ),
    );
  }
}
