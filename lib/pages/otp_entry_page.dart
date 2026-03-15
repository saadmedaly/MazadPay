import 'package:flutter/material.dart';
import 'set_password_page.dart';
import 'package:mezadpay/l10n/app_localizations.dart';

class OtpEntryPage extends StatefulWidget {
  final String phoneNumber;

  const OtpEntryPage({super.key, required this.phoneNumber});

  @override
  State<OtpEntryPage> createState() => _OtpEntryPageState();
}

class _OtpEntryPageState extends State<OtpEntryPage> {
  final List<TextEditingController> _controllers = List.generate(
    6,
    (_) => TextEditingController(),
  );
  final List<FocusNode> _focusNodes = List.generate(6, (_) => FocusNode());

  @override
  void dispose() {
    for (var controller in _controllers) {
      controller.dispose();
    }
    for (var node in _focusNodes) {
      node.dispose();
    }
    super.dispose();
  }

  void _onChanged(String value, int index) {
    if (value.length == 1 && index < 5) {
      _focusNodes[index + 1].requestFocus();
    }
    if (value.isEmpty && index > 0) {
      _focusNodes[index - 1].requestFocus();
    }
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
            child: Row(
              children: [
                InkWell(
                  onTap: () {
                    // Handle dynamic theme change if needed
                  },
                  borderRadius: BorderRadius.circular(20),
                  child: Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: isDarkMode ? Colors.grey[800] : Colors.grey[100],
                    ),
                    child: Icon(
                      isDarkMode ? Icons.light_mode : Icons.dark_mode,
                      color: isDarkMode ? Colors.grey[400] : Colors.grey[600],
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                InkWell(
                  onTap: () {
                    Navigator.pop(context);
                  },
                  borderRadius: BorderRadius.circular(20),
                  child: Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: isDarkMode ? Colors.grey[800] : Colors.grey[100],
                    ),
                    child: Icon(
                      isRtl ? Icons.arrow_forward_ios : Icons.arrow_back_ios,
                      size: 20,
                      color: isDarkMode ? Colors.grey[400] : Colors.grey[600],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24.0),
          child: Column(
            crossAxisAlignment: isRtl ? CrossAxisAlignment.end : CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 8),
              Text(
                l10n.enterVerificationCode,
                style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 12),
              RichText(
                textAlign: isRtl ? TextAlign.right : TextAlign.left,
                text: TextSpan(
                  style: TextStyle(
                    fontSize: 14,
                    height: 1.5,
                    color: isDarkMode ? Colors.grey[400] : Colors.grey[500],
                  ),
                  children: [
                    TextSpan(
                      text: '${l10n.otpUnderPhone}\n',
                    ),
                    TextSpan(
                      text: widget.phoneNumber,
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: isDarkMode ? Colors.grey[200] : Colors.grey[800],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),
              Directionality(
                textDirection: TextDirection.ltr,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: List.generate(6, (index) {
                    return SizedBox(
                      width: 48,
                      height: 56,
                      child: TextField(
                        controller: _controllers[index],
                        focusNode: _focusNodes[index],
                        keyboardType: TextInputType.number,
                        textAlign: TextAlign.center,
                        maxLength: 1,
                        style: const TextStyle(
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                        ),
                        decoration: InputDecoration(
                          counterText: '',
                          contentPadding: EdgeInsets.zero,
                          enabledBorder: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(16),
                            borderSide: BorderSide(
                              color: isDarkMode
                                  ? Colors.grey[700]!
                                  : Colors.grey[200]!,
                              width: 2,
                            ),
                          ),
                          focusedBorder: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(16),
                            borderSide: BorderSide(
                              color: Theme.of(context).primaryColor,
                              width: 2,
                            ),
                          ),
                        ),
                        onChanged: (value) => _onChanged(value, index),
                      ),
                    );
                  }),
                ),
              ),
              const SizedBox(height: 24),
              Center(
                child: RichText(
                  text: TextSpan(
                    style: TextStyle(
                      fontSize: 14,
                      color: isDarkMode ? Colors.grey[500] : Colors.grey[400],
                    ),
                    children: [
                      TextSpan(text: '${l10n.resendCodeIn} '),
                      TextSpan(
                        text: '4:24',
                        style: TextStyle(
                          fontWeight: FontWeight.w500,
                          color: isDarkMode
                              ? Colors.grey[300]
                              : Colors.grey[600],
                        ),
                      ),
                    ],
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
                        builder: (context) => const SetPasswordPage(),
                      ),
                    );
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Theme.of(context).primaryColor,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    elevation: isDarkMode ? 0 : 8,
                    shadowColor: Theme.of(
                      context,
                    ).primaryColor.withValues(alpha: 0.3),
                  ),
                  child: Text(
                    l10n.verify,
                    style: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
              const Spacer(),
              Center(
                child: TextButton(
                  onPressed: () {},
                  style: TextButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 4),
                    minimumSize: Size.zero,
                    tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  ),
                  child: Text(
                    l10n.changeVerificationMethod,
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                      color: isDarkMode ? Colors.grey[400] : Colors.grey[500],
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 10),
            ],
          ),
        ),
      ),
    );
  }
}
