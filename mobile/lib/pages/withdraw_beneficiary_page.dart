import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import '../services/wallet_api.dart';
import '../utils/money_formatter.dart';

/// Customer #34: step 2 of the insurance-refund/withdrawal flow -- "خدمة
/// بنكية" (client reference screenshot "تعديل رقم 7"). Reached from
/// WithdrawPage after the user picks the mobile/Bankily receiving method and
/// confirms an amount; this screen collects the beneficiary phone/account
/// number and submits the SAME withdrawal request WithdrawPage already
/// builds (gateway + amount), just with beneficiaryAccount added -- no
/// second wallet system, no client-side transfer.
class WithdrawBeneficiaryPage extends StatefulWidget {
  final double amount;
  final String gateway;
  final String? currencyCode;

  const WithdrawBeneficiaryPage({
    super.key,
    required this.amount,
    required this.gateway,
    this.currencyCode,
  });

  @override
  State<WithdrawBeneficiaryPage> createState() => _WithdrawBeneficiaryPageState();
}

class _WithdrawBeneficiaryPageState extends State<WithdrawBeneficiaryPage> {
  final TextEditingController _beneficiaryController = TextEditingController();
  final WalletApi _walletApi = WalletApi();
  bool _isLoading = false;

  @override
  void dispose() {
    _beneficiaryController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!mounted) return;
    final messenger = ScaffoldMessenger.of(context);
    final localizations = AppLocalizations.of(context)!;

    final beneficiary = _beneficiaryController.text.trim();
    if (beneficiary.isEmpty) {
      messenger.showSnackBar(SnackBar(content: Text(localizations.text_422)));
      return;
    }

    setState(() => _isLoading = true);

    try {
      // Server re-derives/validates amount and user_id independently
      // (wallet_handler.go: Withdraw, RequestWithdraw) -- this call is a
      // request, never a direct transfer.
      final response = await _walletApi.withdraw(
        amount: widget.amount,
        gateway: widget.gateway,
        beneficiaryAccount: beneficiary,
      );

      if (!mounted) return;
      setState(() => _isLoading = false);

      if (response.success) {
        _showSuccessDialog(context);
      } else {
        messenger.showSnackBar(
          SnackBar(content: Text(response.message ?? localizations.error_withdraw_failed)),
        );
      }
    } catch (e) {
      if (!mounted) return;
      setState(() => _isLoading = false);
      messenger.showSnackBar(SnackBar(content: Text(localizations.error_connection)));
    }
  }

  void _showSuccessDialog(BuildContext context) {
    final localizations = AppLocalizations.of(context)!;
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.check_circle, color: Color(0xFF00C58D), size: 64),
            const SizedBox(height: 24),
            Text(localizations.text_350, style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold)),
            const SizedBox(height: 12),
            Text(
              localizations.text_351,
              textAlign: TextAlign.center,
              style: TextStyle(color: Colors.grey[600]),
            ),
            const SizedBox(height: 32),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: () {
                  Navigator.of(context).pop();
                  Navigator.of(context).pop();
                  Navigator.of(context).pop();
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF0081FF),
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                ),
                child: Text(localizations.text_303),
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final localizations = AppLocalizations.of(context)!;
    const blue = Color(0xFF0084FF);

    return Scaffold(
      backgroundColor: isDark ? const Color(0xFF121212) : const Color(0xFFF5F5F5),
      appBar: AppBar(
        backgroundColor: blue,
        elevation: 0,
        centerTitle: true,
        leading: IconButton(
          icon: const Icon(Icons.arrow_forward_ios, color: Colors.white, size: 18),
          onPressed: () => Navigator.pop(context),
        ),
        title: Text(
          localizations.text_412,
          style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 18),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: const Color(0xFFE8F4FF),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: const Color(0xFFB3D9FF)),
              ),
              child: Row(
                children: [
                  const Icon(Icons.account_balance_outlined, color: blue),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      localizations.text_413,
                      textAlign: TextAlign.right,
                      style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 24),

            Text(localizations.text_414, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
            const SizedBox(height: 8),
            TextField(
              controller: _beneficiaryController,
              keyboardType: TextInputType.phone,
              textAlign: TextAlign.right,
              decoration: InputDecoration(
                hintText: localizations.text_415,
                hintStyle: const TextStyle(color: Colors.grey),
                filled: true,
                fillColor: isDark ? const Color(0xFF1D1D1D) : Colors.white,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
              ),
            ),
            const SizedBox(height: 24),

            Text(localizations.text_416, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
            const SizedBox(height: 12),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: isDark ? const Color(0xFF1D1D1D) : Colors.white,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: Colors.grey.shade300),
              ),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        // Authoritative amount (Customer #34 requirement #7):
                        // this is the same amount WithdrawPage already
                        // validated against the real balance -- never
                        // user-editable on this screen.
                        MoneyFormatter.format(widget.amount, widget.currencyCode),
                        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15, color: blue),
                      ),
                      Text(localizations.text_417, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(localizations.text_419, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15, color: Color(0xFF00C58D))),
                      Text(localizations.text_418, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13)),
                    ],
                  ),
                  const Divider(height: 24),
                  Row(
                    children: [
                      Icon(Icons.info_outline, size: 16, color: Colors.grey[600]),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          localizations.text_420,
                          textAlign: TextAlign.right,
                          style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const SizedBox(height: 40),

            SizedBox(
              width: double.infinity,
              height: 56,
              child: ElevatedButton(
                onPressed: _isLoading ? null : _submit,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF00C58D),
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                ),
                child: _isLoading
                    ? const CircularProgressIndicator(color: Colors.white)
                    : Text(
                        localizations.text_421,
                        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                      ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
