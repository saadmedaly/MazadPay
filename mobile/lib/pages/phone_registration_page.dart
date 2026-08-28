import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'login_page.dart';
import 'otp_entry_page.dart';
import '../services/auth_api.dart';
import '../services/category_api.dart';
import '../utils/error_mapper.dart';
import '../utils/phone_validation.dart';
import '../widgets/country_picker_sheet.dart';

class PhoneRegistrationPage extends ConsumerStatefulWidget {
  const PhoneRegistrationPage({super.key});

  @override
  ConsumerState<PhoneRegistrationPage> createState() =>
      _PhoneRegistrationPageState();
}

class _PhoneRegistrationPageState extends ConsumerState<PhoneRegistrationPage> {
  final _phoneController = TextEditingController();
  final AuthApi _authApi = AuthApi();
  final CategoryApi _categoryApi = CategoryApi();
  bool _isLoading = false;
  List<dynamic> _countries = [];
  Map<String, dynamic>? _selectedCountry;

  @override
  void initState() {
    super.initState();
    _phoneController.addListener(() {
      setState(() {});
    });
    _loadCountries();
  }

  Future<void> _loadCountries() async {
    final response = await _categoryApi.getCountries();
    if (response.success && response.data != null) {
      if (mounted) {
        setState(() {
          _countries = response.data!;
          // Select default country (+222 Mauritania) or first available
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
    _phoneController.dispose();
    super.dispose();
  }

  Future<void> _sendOTP() async {
    if (_isLoading) return;
    final l10n = AppLocalizations.of(context)!;
    final phone = _phoneController.text.trim();

    if (_selectedCountry == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(l10n.error_country_required),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    if (!isPhoneLengthValid(phone, _selectedCountry)) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(l10n.error_phone_length),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      final countryCode = _selectedCountry?['country_code'] ?? '+222';
      final response = await _authApi.sendOTP(
        phone: '$countryCode${_phoneController.text}',
        purpose: 'register',
      );

      setState(() => _isLoading = false);

      if (response.success) {
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (context) => OtpEntryPage(
              phoneNumber: '$countryCode${_phoneController.text}',
            ),
          ),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              mapError(
                context,
                response.error?.code,
                response.error?.message ?? '',
              ),
            ),
          ),
        );
      }
    } catch (e) {
      setState(() => _isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_connection)),
      );
    }
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
        // BACK BUTTON (Right side in RTL)
        leading: Padding(
          padding: const EdgeInsetsDirectional.only(end: 16.0),
          child: Center(
            child: InkWell(
              onTap: () => Navigator.pop(context),
              borderRadius: BorderRadius.circular(20),
              child: Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: isDarkMode ? Colors.grey[800] : Colors.grey[100],
                ),
                child: Icon(
                  Icons.chevron_right,
                  color: isDarkMode ? Colors.grey[300] : Colors.grey[600],
                ),
              ),
            ),
          ),
        ),
        // HELP TEXT (Center)
        title: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              AppLocalizations.of(context)!.text_241,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: isDarkMode ? Colors.grey[400] : Colors.grey[500],
              ),
            ),
            Text(
              AppLocalizations.of(context)!.text_242,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: isDarkMode ? Colors.grey[200] : Colors.grey[800],
              ),
            ),
          ],
        ),
        // ROBOT ICON (Left side in RTL)
        actions: [
          Padding(
            padding: const EdgeInsetsDirectional.only(start: 16.0),
            child: Center(
              child: Container(
                width: 40,
                height: 40,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: isDarkMode ? Colors.grey[800] : Colors.grey[100],
                ),
                child: Icon(
                  Icons.smart_toy_outlined, // Robot head icon
                  size: 20,
                  color: isDarkMode ? Colors.grey[300] : Colors.grey[600],
                ),
              ),
            ),
          ),
        ],
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24.0),
          child: Column(
            children: [
              const SizedBox(height: 8),
              Text(
                AppLocalizations.of(context)!.text_283,
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  height: 1.5,
                ),
              ),
              const SizedBox(height: 20),
              Row(
                textDirection:
                    TextDirection.rtl, // Match Stitch: Flag on the right
                children: [
                  InkWell(
                    onTap: () => _showCountryPicker(context),
                    borderRadius: BorderRadius.circular(12),
                    child: Container(
                      height: 56,
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      decoration: BoxDecoration(
                        color: Theme.of(context).brightness == Brightness.light
                            ? Colors.grey[50]
                            : Colors.grey[800]!.withOpacity(0.5),
                        border: Border.all(
                          color:
                              Theme.of(context).brightness == Brightness.light
                              ? Colors.grey[200]!
                              : Colors.grey[700]!,
                        ),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            _selectedCountry?['country_code'] ?? '+222',
                            style: const TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.bold,
                              color: Colors.grey,
                            ),
                            textDirection: TextDirection.ltr,
                          ),
                          const SizedBox(width: 8),
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
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Container(
                      height: 56,
                      decoration: BoxDecoration(
                        color: Theme.of(context).brightness == Brightness.light
                            ? Colors.grey[50]
                            : Colors.grey[800]!.withOpacity(0.5),
                        border: Border.all(
                          color:
                              Theme.of(context).brightness == Brightness.light
                              ? Colors.grey[200]!
                              : Colors.grey[700]!,
                        ),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: TextField(
                        controller: _phoneController,
                        textAlign: TextAlign.center,
                        keyboardType: TextInputType.phone,
                        style: const TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                          letterSpacing: 2,
                        ),
                        decoration: InputDecoration(
                          hintText: '00000000',
                          border: InputBorder.none,
                          contentPadding: EdgeInsets.zero,
                          hintStyle: TextStyle(
                            color:
                                Theme.of(context).brightness == Brightness.light
                                ? Colors.grey[400]
                                : Colors.grey[600],
                          ),
                          counterText: "", // Hide default counter
                        ),
                        maxLength: phoneMaxLengthFor(_selectedCountry),
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Align(
                alignment: AlignmentDirectional.centerEnd,
                child: Text(
                  '${_phoneController.text.length}/${phoneMaxLengthFor(_selectedCountry)}',
                  style: TextStyle(
                    fontSize: 12,
                    color: Theme.of(context).brightness == Brightness.light
                        ? Colors.grey[500]
                        : Colors.grey[400],
                  ),
                ),
              ),
              const SizedBox(height: 8),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.chat_outlined,
                    size: 16,
                    color: const Color(0xFF25D366),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    AppLocalizations.of(context)!.text_261,
                    style: TextStyle(
                      fontSize: 12,
                      color: isDarkMode ? Colors.grey[400] : Colors.grey[500],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton(
                  onPressed: _isLoading ? null : _sendOTP,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF0081FF),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    elevation: 0,
                  ),
                  child: _isLoading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(
                            color: Colors.white,
                            strokeWidth: 2,
                          ),
                        )
                      : Text(
                          AppLocalizations.of(context)!.text_211,
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                          ),
                        ),
                ),
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    AppLocalizations.of(context)!.text_391,
                    style: TextStyle(
                      fontSize: 14,
                      color: isDarkMode ? Colors.grey[400] : Colors.grey[600],
                    ),
                  ),
                  GestureDetector(
                    onTap: () {
                      Navigator.pushReplacement(
                        context,
                        MaterialPageRoute(
                          builder: (context) => const LoginPage(),
                        ),
                      );
                    },
                    child: Text(
                      AppLocalizations.of(context)!.text_217,
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: const Color(0xFF0081FF),
                      ),
                    ),
                  ),
                ],
              ),
              const Spacer(),
              Opacity(
                opacity: 0.2,
                child: Column(
                  children: [
                    Container(
                      width: 24,
                      height: 24,
                      decoration: BoxDecoration(
                        color: const Color(0xFF38bdf8),
                        borderRadius: BorderRadius.circular(6),
                      ),
                    ),
                    const SizedBox(height: 4),
                    const Text(
                      'MAZADPAY',
                      style: TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                        letterSpacing: 2,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 10),
            ],
          ),
        ),
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
