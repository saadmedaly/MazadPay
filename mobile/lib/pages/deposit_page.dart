import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:mezadpay/pages/payment_details_page.dart';
import '../services/wallet_api.dart';
import '../services/payment_methods_service.dart';
import '../models/payment_method.dart';

// Helper to get the name according to the app's current locale
String _localizedMethodName(PaymentMethod method, BuildContext context) {
  final locale = Localizations.localeOf(context);
  switch (locale.languageCode) {
    case 'ar':
      return method.nameAr;
    case 'fr':
      return method.nameFr;
    case 'en':
      return method.nameEn ?? method.nameFr;
    default:
      // Fallback to French if unknown locale
      return method.nameFr;
  }
}


class DepositPage extends StatefulWidget {
  const DepositPage({super.key});

  @override
  State<DepositPage> createState() => _DepositPageState();
}

class _DepositPageState extends State<DepositPage> {
  String? _selectedMethodId;
  bool _hasShownModal = false;
  final WalletApi _walletApi = WalletApi();
  bool _isLoading = false;

  List<PaymentMethod> _methods = [];

  @override
  void initState() {
    super.initState();
    _loadMethods();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!_hasShownModal) {
        _showTermsBottomSheet();
        _hasShownModal = true;
      }
    });
  }

  Future<void> _loadMethods() async {
    try {
      final service = PaymentMethodsService();
      final fetched = await service.getPaymentMethods();
      setState(() {
        _methods = fetched;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _methods = [];
        _isLoading = false;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Failed to load payment methods')),
      );
    }
  }

  Future<void> _makeDeposit() async {
    if (_selectedMethodId == null) return;

    setState(() => _isLoading = true);

    try {
      final response = await _walletApi.deposit(
        method: _selectedMethodId!,
        amount: 0, // Amount will be set in PaymentDetailsPage
      );

      setState(() => _isLoading = false);

      if (response.success) {
        final selectedMethod = _methods.firstWhere((m) => m.code == _selectedMethodId);
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (context) => PaymentDetailsPage(
              methodName: _localizedMethodName(selectedMethod, context),
            ),
          ),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(response.message ?? AppLocalizations.of(context)!.error_deposit_failed)),
        );
      }
    } catch (e) {
      setState(() => _isLoading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context)!.error_connection)),
      );
    }
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
                    _isLoading
            ? const Center(child: CircularProgressIndicator())
            : Column(
                children: _methods
                    .map((method) => _buildPaymentMethodTile(method, isDarkMode))
                    .toList(),
              ),
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
                      onPressed: _isLoading ? null : _makeDeposit,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF0084FF),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                      child: _isLoading
                          ? const CircularProgressIndicator(color: Colors.white)
                          : Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                const Icon(Icons.arrow_back, color: Colors.white, size: 20),
                                const SizedBox(width: 12),
                                Text(
                                  'ادفع عبر ${_localizedMethodName(_methods.firstWhere((m) => m.code == _selectedMethodId), context)}',
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
    bool isSelected = _selectedMethodId == method.code;
    return Card(
      elevation: 0,
      margin: const EdgeInsets.only(bottom: 12),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: isSelected ? const Color(0xFF0084FF) : Colors.grey.withOpacity(0.2),
          width: isSelected ? 2 : 1,
        ),
      ),
      color: isDarkMode ? const Color(0xFF1D1D1D) : Colors.white,
      child: ListTile(
        onTap: () => setState(() => _selectedMethodId = method.code),
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        leading: method.logoUrl != null && method.logoUrl!.isNotEmpty
            ? Image.network(method.logoUrl!, height: 40, width: 40, fit: BoxFit.contain)
            : const Icon(Icons.payment, size: 40, color: Colors.grey),
        title: Text(
          _localizedMethodName(method, context),
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Text(
          '${AppLocalizations.of(context)!.text_173} ${_localizedMethodName(method, context)}',
          style: const TextStyle(fontSize: 12, color: Colors.grey),
        ),
        trailing: isSelected
            ? const Icon(Icons.radio_button_checked, color: Color(0xFF0084FF))
            : const Icon(Icons.radio_button_unchecked, color: Colors.grey),
      ),
    );
  }
}