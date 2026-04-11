import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:mezadpay/pages/payment_details_page.dart';

class PaymentMethod {
  final String id;
  final String title;
  final String subtitle;
  final String logoUrl;
  final Color highlightColor;

  PaymentMethod(this.id, this.title, this.subtitle, this.logoUrl, this.highlightColor);
}

class DepositPage extends StatefulWidget {
  const DepositPage({super.key});

  @override
  State<DepositPage> createState() => _DepositPageState();
}

class _DepositPageState extends State<DepositPage> {
  String? _selectedMethodId;
  bool _hasShownModal = false;

  late List<PaymentMethod> _methods;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final l10n = AppLocalizations.of(context)!;
    _methods = [
      PaymentMethod(
        'masrvi',
        l10n.text_170,
        l10n.text_171,
        'assets/Masrivi.png',
        const Color(0xFF00A99D),
      ),
      PaymentMethod(
        'bankily',
        l10n.text_172,
        l10n.text_173,
        'assets/Bankily.png',
        const Color(0xFF0084FF),
      ),
      PaymentMethod(
        'sedad',
        l10n.text_174,
        l10n.text_175,
        'assets/Sedad.png',
        const Color(0xFF33CC33),
      ),
      PaymentMethod(
        'click',
        l10n.text_176,
        l10n.text_177,
        'assets/Click.png',
        Colors.black,
      ),
    ];
  }

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!_hasShownModal) {
        _showTermsBottomSheet();
        _hasShownModal = true;
      }
    });
  }

  void _showTermsBottomSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (BuildContext context) {
        bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
        return Container(
          decoration: BoxDecoration(
            color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
          ),
          constraints: BoxConstraints(
            maxHeight: MediaQuery.of(context).size.height * 0.85,
          ),
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
              Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.grey[300],
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              const SizedBox(height: 24),
              // Icon Placeholder
              Container(
                height: 80,
                width: 80,
                decoration: BoxDecoration(
                  color: Colors.blue[50],
                  shape: BoxShape.circle,
                ),
                child: Center(
                  child: Icon(Icons.document_scanner, color: const Color(0xFF0084FF), size: 40),
                ),
              ),
              const SizedBox(height: 20),
              Text(
                AppLocalizations.of(context)!.text_178,
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              Text(
                AppLocalizations.of(context)!.text_179,
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 14, color: Colors.grey, height: 1.5),
              ),
              const SizedBox(height: 24),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  Text(AppLocalizations.of(context)!.text_180, style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
                  SizedBox(width: 8),
                  Icon(Icons.article, color: Colors.brown, size: 18),
                ],
              ),
              const SizedBox(height: 16),
              // Terms list
              _buildTermItem(AppLocalizations.of(context)!.text_181),
              _buildTermItem(AppLocalizations.of(context)!.text_182),
              _buildTermItem(AppLocalizations.of(context)!.text_183),
              _buildTermItem(AppLocalizations.of(context)!.text_184),
              _buildTermItem(AppLocalizations.of(context)!.text_185),
              const SizedBox(height: 24),
              Text(
                AppLocalizations.of(context)!.text_186,
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12),
              ),
              const SizedBox(height: 24),
              Row(
                children: [
                  Expanded(
                    child: ElevatedButton(
                      onPressed: () => Navigator.pop(context),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Colors.grey[200],
                        foregroundColor: Colors.black,
                        elevation: 0,
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                      child: Text(AppLocalizations.of(context)!.text_187, style: TextStyle(fontWeight: FontWeight.bold)),
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    flex: 2,
                    child: ElevatedButton(
                      onPressed: () => Navigator.pop(context),
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF0084FF),
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                      child: Text(AppLocalizations.of(context)!.text_188, style: TextStyle(fontWeight: FontWeight.bold, color: Colors.white)),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16), // Bottom safe area
            ],
          ),
          ),
        );
      },
    );
  }

  Widget _buildTermItem(String text) {
    return Padding(
      padding: const EdgeInsetsDirectional.only(bottom: 8.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Text(
              text,
              textAlign: TextAlign.right,
              style: const TextStyle(fontSize: 12, height: 1.5, fontWeight: FontWeight.w500),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    bool isDarkMode = Theme.of(context).brightness == Brightness.dark;
    
    return Scaffold(
        backgroundColor: isDarkMode ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: PreferredSize(
          preferredSize: const Size.fromHeight(70),
          child: SafeArea(
            child: SizedBox(
              height: 70,
              child: Row(
                children: [
                  IconButton(
                    icon: Icon(Icons.arrow_forward_ios, color: isDarkMode ? Colors.white : Colors.black, size: 20),
                    onPressed: () => Navigator.pop(context),
                  ),
                  Text(
                    AppLocalizations.of(context)!.text_189,
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18, color: isDarkMode ? Colors.white : Colors.black),
                  ),
                ],
              ),
            ),
          ),
        ),
        body: Column(
          children: [
            Expanded(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Padding(
                      padding: EdgeInsets.symmetric(vertical: 8.0),
                      child: Text(
                        AppLocalizations.of(context)!.text_190,
                        style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                      ),
                    ),
                    const SizedBox(height: 16),
                    ..._methods.map((method) => _buildPaymentMethodTile(method, isDarkMode)),
                  ],
                ),
              ),
            ),
            
            // Bottom Action Button
            if (_selectedMethodId != null)
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.05),
                      offset: const Offset(0, -4),
                      blurRadius: 10,
                    ),
                  ],
                ),
                child: SafeArea(
                  child: SizedBox(
                    width: double.infinity,
                    height: 54,
                    child: ElevatedButton(
                      onPressed: () {
                        final selectedMethod = _methods.firstWhere((m) => m.id == _selectedMethodId);
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (context) => PaymentDetailsPage(
                              methodName: selectedMethod.title,
                            ),
                          ),
                        );
                      },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF0084FF),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(Icons.arrow_back, color: Colors.white, size: 20),
                          const SizedBox(width: 12),
                          Text(
                            'ادفع عبر ${_methods.firstWhere((m) => m.id == _selectedMethodId).title}',
                            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16, color: Colors.white),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
          ],
        ),
      );
  }

  Widget _buildPaymentMethodTile(PaymentMethod method, bool isDarkMode) {
    bool isSelected = _selectedMethodId == method.id;
    
    return GestureDetector(
      onTap: () {
        setState(() {
          _selectedMethodId = method.id;
        });
      },
      child: Container(
        margin: const EdgeInsetsDirectional.only(bottom: 16),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: isSelected ? Colors.green : Colors.transparent,
            width: 2,
          ),
          boxShadow: [
            if (!isDarkMode && !isSelected)
              BoxShadow(
                color: Colors.black.withOpacity(0.02),
                blurRadius: 8,
                offset: const Offset(0, 2),
              ),
          ],
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
             Image.asset(
              method.logoUrl,
              height: 40,
              width: 80,
              fit: BoxFit.contain,
              errorBuilder: (context, error, stackTrace) => Text(
                method.title,
                style: TextStyle(color: method.highlightColor, fontWeight: FontWeight.bold),
              ),
            ),
            const SizedBox(width: 16),
            
            // Text Content in the Middle
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    method.title,
                    style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    method.subtitle,
                    style: TextStyle(color: Colors.grey[600], fontSize: 12, height: 1.3),
                  ),
                ],
              ),
            ),

            // Checkmark on the Left (End in RTL)
            if (isSelected)
              const Icon(Icons.check_circle, color: Colors.green, size: 24),
          ],
        ),
      ),
    );
  }
}