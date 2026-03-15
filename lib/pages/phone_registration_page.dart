import 'package:flutter/material.dart';
import 'otp_entry_page.dart';
import 'package:mezadpay/l10n/app_localizations.dart';

class PhoneRegistrationPage extends StatefulWidget {
  const PhoneRegistrationPage({super.key});

  @override
  State<PhoneRegistrationPage> createState() => _PhoneRegistrationPageState();
}

class _PhoneRegistrationPageState extends State<PhoneRegistrationPage> {
  final _phoneController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _phoneController.addListener(() {
      setState(() {});
    });
  }

  @override
  void dispose() {
    _phoneController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final l10n = AppLocalizations.of(context)!;
    bool isRtl = Directionality.of(context) == TextDirection.rtl;

    return Scaffold(
      backgroundColor: Theme.of(context).scaffoldBackgroundColor,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: Container(),
        actions: [
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24.0),
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
                  isRtl ? Icons.chevron_right : Icons.chevron_left,
                  color: isDarkMode ? Colors.grey[300] : Colors.grey[600],
                ),
              ),
            ),
          ),
        ],
        title: Row(
          mainAxisAlignment: isRtl ? MainAxisAlignment.end : MainAxisAlignment.start,
          children: [
            Column(
              crossAxisAlignment: isRtl ? CrossAxisAlignment.end : CrossAxisAlignment.start,
              children: [
                Text(
                  l10n.needHelp,
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: isDarkMode ? Colors.grey[400] : Colors.grey[500],
                  ),
                ),
                Text(
                  l10n.contactSupport,
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    color: isDarkMode ? Colors.grey[200] : Colors.grey[800],
                  ),
                ),
              ],
            ),
            const SizedBox(width: 12),
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: isDarkMode ? Colors.grey[800] : Colors.grey[100],
              ),
              child: Icon(
                Icons.headset_mic,
                size: 20,
                color: isDarkMode ? Colors.grey[300] : Colors.grey[600],
              ),
            ),
          ],
        ),
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24.0),
          child: Column(
            children: [
              const SizedBox(height: 8),
              Text(
                l10n.registrationPrompt,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  height: 1.5,
                ),
              ),
              const SizedBox(height: 20),
              Row(
                textDirection: TextDirection.ltr,
                children: [
                  InkWell(
                    onTap: () => _showCountryPicker(context),
                    borderRadius: BorderRadius.circular(12),
                    child: Container(
                      height: 56,
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      decoration: BoxDecoration(
                        color: isDarkMode ? Colors.grey[800]!.withValues(alpha: 0.5) : Colors.grey[50],
                        border: Border.all(
                          color: isDarkMode ? Colors.grey[700]! : Colors.grey[200]!,
                        ),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Text(
                            '+222',
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.bold,
                              color: Colors.grey,
                            ),
                          ),
                          const SizedBox(width: 8),
                          ClipRRect(
                            borderRadius: BorderRadius.circular(2),
                            child: Image.network(
                              'https://flagcdn.com/w80/mr.png',
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
                        color: isDarkMode ? Colors.grey[800]!.withValues(alpha: 0.5) : Colors.grey[50],
                        border: Border.all(
                          color: isDarkMode ? Colors.grey[700]! : Colors.grey[200]!,
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
                            color: isDarkMode ? Colors.grey[600] : Colors.grey[400],
                          ),
                          counterText: "",
                        ),
                        maxLength: 8,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Align(
                alignment: isRtl ? Alignment.centerRight : Alignment.centerLeft,
                child: Text(
                  '${_phoneController.text.length}/8',
                  style: TextStyle(
                    fontSize: 12,
                    color: isDarkMode ? Colors.grey[400] : Colors.grey[500],
                  ),
                ),
              ),
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton(
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => OtpEntryPage(phoneNumber: _phoneController.text),
                      ),
                    );
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Theme.of(context).primaryColor,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    elevation: isDarkMode ? 0 : 8,
                    shadowColor: Theme.of(context).primaryColor.withValues(alpha: 0.3),
                  ),
                  child: Text(
                    l10n.next,
                    style: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                ),
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
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final l10n = AppLocalizations.of(context)!;
    bool isRtl = Directionality.of(context) == TextDirection.rtl;

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
              Text(
                l10n.selectCountry,
                style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 24),
              _buildCountryItem(
                context,
                name: l10n.mauritania,
                code: '+222',
                flagUrl: 'https://flagcdn.com/w80/mr.png',
                isAvailable: true,
              ),
              const Divider(height: 32),
              Align(
                alignment: isRtl ? Alignment.centerRight : Alignment.centerLeft,
                child: Text(
                  l10n.comingSoon,
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    color: Colors.grey,
                  ),
                ),
              ),
              const SizedBox(height: 16),
              _buildCountryItem(
                context,
                name: l10n.senegal,
                code: '+221',
                flagUrl: 'https://flagcdn.com/w80/sn.png',
                isAvailable: false,
              ),
              const SizedBox(height: 12),
              _buildCountryItem(
                context,
                name: l10n.morocco,
                code: '+212',
                flagUrl: 'https://flagcdn.com/w80/ma.png',
                isAvailable: false,
              ),
              const SizedBox(height: 12),
              _buildCountryItem(
                context,
                name: l10n.tunisia,
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
                  ? const Color(0xFF135BEC).withValues(alpha: 0.3)
                  : Colors.grey.withValues(alpha: 0.2),
            ),
          ),
          child: Row(
            children: [
              ClipRRect(
                borderRadius: BorderRadius.circular(2),
                child: Image.network(
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
}
