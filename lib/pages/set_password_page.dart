import 'package:flutter/material.dart';
import 'login_page.dart';
import 'package:mezadpay/l10n/app_localizations.dart';

class SetPasswordPage extends StatefulWidget {
  const SetPasswordPage({super.key});

  @override
  State<SetPasswordPage> createState() => _SetPasswordPageState();
}

class _SetPasswordPageState extends State<SetPasswordPage> {
  final List<TextEditingController> _passControllers = List.generate(
    4,
    (_) => TextEditingController(),
  );
  final List<FocusNode> _passFocusNodes = List.generate(4, (_) => FocusNode());

  final List<TextEditingController> _confirmControllers = List.generate(
    4,
    (_) => TextEditingController(),
  );
  final List<FocusNode> _confirmFocusNodes = List.generate(
    4,
    (_) => FocusNode(),
  );

  @override
  void dispose() {
    for (var controller in _passControllers) {
      controller.dispose();
    }
    for (var node in _passFocusNodes) {
      node.dispose();
    }
    for (var controller in _confirmControllers) {
      controller.dispose();
    }
    for (var node in _confirmFocusNodes) {
      node.dispose();
    }
    super.dispose();
  }

  void _onPassChanged(String value, int index) {
    if (value.length == 1) {
      if (index < 3) {
        _passFocusNodes[index + 1].requestFocus();
      } else {
        _confirmFocusNodes[0].requestFocus();
      }
    } else if (value.isEmpty && index > 0) {
      _passFocusNodes[index - 1].requestFocus();
    }
  }

  void _onConfirmChanged(String value, int index) {
    if (value.length == 1 && index < 3) {
      _confirmFocusNodes[index + 1].requestFocus();
    } else if (value.isEmpty && index > 0) {
      _confirmFocusNodes[index - 1].requestFocus();
    }
  }

  void _showSuccessDialog() {
    final l10n = AppLocalizations.of(context)!;
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) {
        bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
        return Dialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
          ),
          backgroundColor: isDarkMode ? Colors.grey[900] : Colors.white,
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: 80,
                  height: 80,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: Theme.of(context).primaryColor.withValues(alpha: 0.2),
                  ),
                  child: Center(
                    child: Container(
                      width: 60,
                      height: 60,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: Theme.of(context).primaryColor,
                      ),
                      child: const Icon(
                        Icons.check,
                        color: Colors.white,
                        size: 40,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 24),
                Text(
                  l10n.accountCreatedSuccess,
                  style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 24),
                SizedBox(
                  width: double.infinity,
                  height: 48,
                  child: ElevatedButton(
                    onPressed: () {
                      Navigator.of(context).pushAndRemoveUntil(
                          MaterialPageRoute(
                              builder: (context) => const LoginPage()),
                          (route) => false);
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Theme.of(context).primaryColor,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      elevation: 0,
                    ),
                    child: Text(
                      l10n.back,
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildPinRow(
    List<TextEditingController> controllers,
    List<FocusNode> focusNodes,
    Function(String, int) onChanged,
    bool isDarkMode,
  ) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: List.generate(4, (index) {
        return Container(
          width: 45,
          height: 60,
          margin: const EdgeInsets.symmetric(horizontal: 6),
          child: TextField(
            controller: controllers[index],
            focusNode: focusNodes[index],
            keyboardType: TextInputType.number,
            textAlign: TextAlign.center,
            maxLength: 1,
            style: const TextStyle(fontSize: 28, fontWeight: FontWeight.bold),
            decoration: InputDecoration(
              counterText: '',
              contentPadding: EdgeInsets.zero,
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(
                  color: isDarkMode
                      ? const Color(0xFF333333)
                      : Theme.of(context).primaryColor.withValues(alpha: 0.5),
                  width: 2,
                ),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(
                  color: Theme.of(context).primaryColor,
                  width: 2,
                ),
              ),
            ),
            onChanged: (value) => onChanged(value, index),
          ),
        );
      }),
    );
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: Theme.of(context).scaffoldBackgroundColor,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: Container(),
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 24.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                const SizedBox(height: 12),
                // Illustration
                SizedBox(
                  height: 100,
                  width: 100,
                  child: Stack(
                    alignment: Alignment.center,
                    children: [
                      Icon(
                        Icons.sync,
                        size: 90,
                        color: Theme.of(context).primaryColor.withValues(alpha: 0.2),
                      ),
                      const Positioned(
                        top: 20,
                        right: 15,
                        child: Icon(
                          Icons.vpn_key,
                          size: 25,
                          color: Color(0xFFA08480),
                        ),
                      ),
                      const Positioned(
                        bottom: 25,
                        child: Icon(
                          Icons.lock_open,
                          size: 40,
                          color: Color(0xFFFFD54F),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 20),
                Text(
                  l10n.newPassword,
                  style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 16),
                Directionality(
                  textDirection: TextDirection.ltr,
                  child: _buildPinRow(
                    _passControllers,
                    _passFocusNodes,
                    _onPassChanged,
                    isDarkMode,
                  ),
                ),
                const SizedBox(height: 20),
                Text(
                  l10n.confirmNewPassword,
                  style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 16),
                Directionality(
                  textDirection: TextDirection.ltr,
                  child: _buildPinRow(
                    _confirmControllers,
                    _confirmFocusNodes,
                    _onConfirmChanged,
                    isDarkMode,
                  ),
                ),
                const SizedBox(height: 20),
                SizedBox(
                  width: double.infinity,
                  height: 48,
                  child: ElevatedButton(
                    onPressed: _showSuccessDialog,
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
                      l10n.next,
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 20),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
