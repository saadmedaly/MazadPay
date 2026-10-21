import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:mezadpay/core/theme.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/models/bankily_payment.dart';
import 'package:mezadpay/pages/bankily_payment_status_page.dart';
import 'package:mezadpay/services/api_service.dart';
import 'package:mezadpay/services/bankily_api.dart';
import 'package:uuid/uuid.dart';

/// Bankily B-PAY automatic wallet top-up (shared Android + iOS flow).
///
/// The customer generates a one-time payment code in the Bankily app (B-PAY,
/// MazadPay merchant code, amount, PIN) and enters it here with their Bankily
/// phone number. Only the MazadPay backend talks to Bankily. The payment code
/// is kept in memory for this screen only -- never persisted -- and the
/// Bankily PIN is never asked for.
class BankilyBpayPage extends StatefulWidget {
  final BankilyConfig config;
  final BankilyApi? api;

  const BankilyBpayPage({super.key, required this.config, this.api});

  @override
  State<BankilyBpayPage> createState() => _BankilyBpayPageState();
}

class _BankilyBpayPageState extends State<BankilyBpayPage> {
  late final BankilyApi _api = widget.api ?? BankilyApi();
  final _amountController = TextEditingController();
  final _phoneController = TextEditingController();
  final _passcodeController = TextEditingController();
  bool _submitting = false;
  List<BankilyPayment> _recent = const [];

  // One key per submission: reused for every retry of the same submission
  // (double tap, network timeout) so the backend never starts a second
  // payment; reset when the user changes the amount or phone.
  String? _idempotencyKey;
  String? _keyFor;

  @override
  void initState() {
    super.initState();
    _loadRecent();
  }

  @override
  void dispose() {
    _passcodeController.clear();
    _amountController.dispose();
    _phoneController.dispose();
    _passcodeController.dispose();
    super.dispose();
  }

  Future<void> _loadRecent() async {
    try {
      final items = await _api.list();
      if (mounted) setState(() => _recent = items.take(5).toList());
    } catch (_) {}
  }

  double? get _amount => double.tryParse(_amountController.text.trim());

  String _keyForCurrentInput() {
    final signature = '${_amountController.text.trim()}|${BankilyInputRules.normalizePhone(_phoneController.text)}';
    if (_idempotencyKey == null || _keyFor != signature) {
      _idempotencyKey = const Uuid().v4();
      _keyFor = signature;
    }
    return _idempotencyKey!;
  }

  void _snack(String msg) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
  }

  String _errorMessage(Object e, AppLocalizations l10n) {
    if (e is ApiException) {
      switch (e.code) {
        case 'bankily_invalid_phone':
          return l10n.bpay_err_phone;
        case 'bankily_invalid_passcode':
          return l10n.bpay_err_passcode;
        case 'bankily_invalid_amount':
          return l10n.bpay_err_amount;
        case 'bankily_disabled':
        case 'bankily_currency_unsupported':
          return l10n.bpay_err_unavailable;
      }
      if (e.message.isNotEmpty && e.httpStatus != null && e.httpStatus! < 500) return e.message;
    }
    return l10n.bpay_err_generic;
  }

  Future<void> _submit() async {
    final l10n = AppLocalizations.of(context)!;
    if (_submitting) return;
    if (!BankilyInputRules.isValidAmount(_amount, widget.config)) {
      _snack(l10n.bpay_err_amount);
      return;
    }
    if (!BankilyInputRules.isValidPhone(_phoneController.text)) {
      _snack(l10n.bpay_err_phone);
      return;
    }
    if (!BankilyInputRules.isValidPasscode(_passcodeController.text)) {
      _snack(l10n.bpay_err_passcode);
      return;
    }
    FocusScope.of(context).unfocus();
    setState(() => _submitting = true);

    final key = _keyForCurrentInput();
    BankilyPayment? payment;
    Object? lastError;
    // A network failure does not mean the payment failed: retry with the SAME
    // idempotency key -- the backend hands back the existing attempt.
    for (var attempt = 0; attempt < 3 && payment == null; attempt++) {
      try {
        payment = await _api.pay(
          amount: _amountController.text.trim(),
          clientPhone: BankilyInputRules.normalizePhone(_phoneController.text),
          passcode: _passcodeController.text.trim(),
          idempotencyKey: key,
        );
      } catch (e) {
        lastError = e;
        if (!BankilyApi.isUncertain(e)) break;
        await Future<void>.delayed(Duration(seconds: 2 * (attempt + 1)));
      }
    }
    if (!mounted) return;
    setState(() => _submitting = false);

    if (payment == null) {
      _snack(_errorMessage(lastError ?? Exception(), l10n));
      return;
    }
    _passcodeController.clear(); // single-use code: never kept after submission
    await _openStatus(payment);
  }

  Future<void> _openStatus(BankilyPayment payment) async {
    final result = await Navigator.of(context).push<String>(
      MaterialPageRoute(builder: (_) => BankilyPaymentStatusPage(payment: payment, api: _api)),
    );
    if (!mounted) return;
    if (result == 'manual') {
      Navigator.of(context).pop();
      return;
    }
    if (result == 'retry') {
      _idempotencyKey = null; // a new attempt needs a new key and a new code
    }
    _loadRecent();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final isDark = Theme.of(context).brightness == Brightness.dark;
    const blue = Color(0xFF0084FF);
    final cfg = widget.config;
    final pending = _recent.where((p) => p.status == BankilyPaymentStatus.pending).toList();

    return Scaffold(
      backgroundColor: isDark ? const Color(0xFF121212) : const Color(0xFFF5F5F5),
      appBar: AppBar(
        backgroundColor: blue,
        systemOverlayStyle: AppTheme.lightStatusIcons,
        elevation: 0,
        centerTitle: true,
        iconTheme: const IconThemeData(color: Colors.white),
        title: Text(l10n.bpay_title, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 18)),
      ),
      body: AbsorbPointer(
        absorbing: _submitting,
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (pending.isNotEmpty) ...[
                _pendingBanner(pending, l10n),
                const SizedBox(height: 16),
              ],
              _instructions(l10n, cfg, blue),
              const SizedBox(height: 20),
              _label(l10n.bpay_amount_label),
              _field(
                key: const Key('bpay-amount'),
                controller: _amountController,
                hint: l10n.bpay_amount_hint,
                keyboard: TextInputType.numberWithOptions(decimal: cfg.allowDecimals),
                formatters: [FilteringTextInputFormatter.allow(RegExp(cfg.allowDecimals ? r'[0-9.]' : r'[0-9]'))],
                isDark: isDark,
                helper: l10n.bpay_amount_range(cfg.minAmount.toStringAsFixed(0), cfg.maxAmount.toStringAsFixed(0)),
              ),
              const SizedBox(height: 16),
              _label(l10n.bpay_phone_label),
              _field(
                key: const Key('bpay-phone'),
                controller: _phoneController,
                hint: l10n.bpay_phone_hint,
                keyboard: TextInputType.phone,
                formatters: [FilteringTextInputFormatter.allow(RegExp(r'[0-9+]')), LengthLimitingTextInputFormatter(13)],
                isDark: isDark,
                autofill: const [AutofillHints.telephoneNumber],
              ),
              const SizedBox(height: 16),
              _label(l10n.bpay_passcode_label),
              _field(
                key: const Key('bpay-passcode'),
                controller: _passcodeController,
                hint: l10n.bpay_passcode_hint,
                keyboard: TextInputType.visiblePassword,
                formatters: [FilteringTextInputFormatter.allow(RegExp(r'[A-Za-z0-9]')), LengthLimitingTextInputFormatter(32)],
                isDark: isDark,
                obscure: true,
              ),
              const SizedBox(height: 8),
              Row(
                children: [
                  const Icon(Icons.lock_outline, size: 16, color: Colors.grey),
                  const SizedBox(width: 6),
                  Expanded(child: Text(l10n.bpay_passcode_note, style: TextStyle(fontSize: 12, color: Colors.grey.shade600))),
                ],
              ),
              const SizedBox(height: 28),
              SizedBox(
                height: 54,
                child: ElevatedButton(
                  key: const Key('bpay-pay'),
                  onPressed: _submitting ? null : _submit,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: blue,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    elevation: 0,
                  ),
                  child: _submitting
                      ? Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2)),
                            const SizedBox(width: 12),
                            Text(l10n.bpay_processing_title, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
                          ],
                        )
                      : Text(l10n.bpay_pay_button, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16)),
                ),
              ),
              if (_submitting) ...[
                const SizedBox(height: 8),
                Text(l10n.bpay_processing_body, textAlign: TextAlign.center, style: TextStyle(fontSize: 12, color: Colors.grey.shade600)),
              ],
              const SizedBox(height: 24),
            ],
          ),
        ),
      ),
    );
  }

  Widget _pendingBanner(List<BankilyPayment> pending, AppLocalizations l10n) {
    // Material (not a decorated Container) so each ListTile has a Material
    // ancestor to paint its background and ink ripple on.
    return Material(
      color: const Color(0xFFFFF7E6),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: Color(0xFFF59E0B)),
      ),
      clipBehavior: Clip.antiAlias,
      child: Column(
        children: pending
            .map((p) => ListTile(
                  key: ValueKey('bpay-pending-${p.id}'),
                  leading: const Icon(Icons.hourglass_top, color: Color(0xFFF59E0B)),
                  title: Text('${l10n.bpay_status_pending} · ${p.amount.toStringAsFixed(0)} MRU'),
                  subtitle: Text(p.operationId),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => _openStatus(p),
                ))
            .toList(),
      ),
    );
  }

  Widget _instructions(AppLocalizations l10n, BankilyConfig cfg, Color blue) {
    final steps = [l10n.bpay_step1, l10n.bpay_step2, l10n.bpay_step3, l10n.bpay_step4];
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFE8F4FF),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFB3D9FF)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              Image.asset('assets/logo-bankily.png', width: 28, height: 28, errorBuilder: (_, _, _) => const SizedBox.shrink()),
              const SizedBox(width: 8),
              Expanded(
                child: Text(l10n.bpay_instructions_title, style: TextStyle(color: blue, fontWeight: FontWeight.bold, fontSize: 14)),
              ),
            ],
          ),
          const SizedBox(height: 10),
          for (var i = 0; i < steps.length; i++)
            Padding(
              padding: const EdgeInsets.only(bottom: 6),
              child: Text('${i + 1}. ${steps[i]}', style: const TextStyle(fontSize: 13, height: 1.5)),
            ),
          if (cfg.merchantCode.isNotEmpty) ...[
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: const Color(0xFFB3D9FF)),
              ),
              child: Row(
                children: [
                  Text('${l10n.bpay_merchant_code}: ', style: const TextStyle(fontSize: 13, color: Colors.black87)),
                  Expanded(
                    child: Text(
                      cfg.merchantCode,
                      key: const Key('bpay-merchant-code'),
                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18, letterSpacing: 1, color: Colors.black),
                    ),
                  ),
                  IconButton(
                    icon: Icon(Icons.copy_outlined, color: blue, size: 20),
                    onPressed: () {
                      Clipboard.setData(ClipboardData(text: cfg.merchantCode));
                      _snack(l10n.bpay_copied);
                    },
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _label(String text) => Padding(
        padding: const EdgeInsets.only(bottom: 8),
        child: Text(text, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),
      );

  Widget _field({
    required Key key,
    required TextEditingController controller,
    required String hint,
    required TextInputType keyboard,
    required bool isDark,
    List<TextInputFormatter>? formatters,
    String? helper,
    bool obscure = false,
    Iterable<String>? autofill,
  }) {
    const blue = Color(0xFF0084FF);
    final border = OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: Colors.grey.shade300));
    return TextField(
      key: key,
      controller: controller,
      keyboardType: keyboard,
      inputFormatters: formatters,
      obscureText: obscure,
      enableSuggestions: !obscure,
      autocorrect: false,
      autofillHints: obscure ? null : autofill,
      enableIMEPersonalizedLearning: !obscure,
      decoration: InputDecoration(
        hintText: hint,
        helperText: helper,
        hintStyle: const TextStyle(color: Colors.grey),
        filled: true,
        fillColor: isDark ? const Color(0xFF1D1D1D) : Colors.white,
        border: border,
        enabledBorder: border,
        focusedBorder: border.copyWith(borderSide: const BorderSide(color: blue)),
      ),
    );
  }
}
