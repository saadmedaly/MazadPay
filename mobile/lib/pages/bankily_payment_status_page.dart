import 'dart:async';

import 'package:flutter/material.dart';
import 'package:mezadpay/core/theme.dart';
import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:mezadpay/models/bankily_payment.dart';
import 'package:mezadpay/services/bankily_api.dart';

/// Shows the outcome of one Bankily B-PAY attempt. While the backend still
/// verifies it (timeout / Bankily "TA"), the page polls the backend and
/// offers "Check again" -- which only asks the backend to run Bankily's
/// checkTransaction; it can never credit the wallet by itself.
class BankilyPaymentStatusPage extends StatefulWidget {
  final BankilyPayment payment;
  final BankilyApi? api;
  final Duration pollInterval;
  final int maxPolls;

  const BankilyPaymentStatusPage({
    super.key,
    required this.payment,
    this.api,
    this.pollInterval = const Duration(seconds: 5),
    this.maxPolls = 24,
  });

  @override
  State<BankilyPaymentStatusPage> createState() => _BankilyPaymentStatusPageState();
}

class _BankilyPaymentStatusPageState extends State<BankilyPaymentStatusPage> {
  late final BankilyApi _api = widget.api ?? BankilyApi();
  late BankilyPayment _payment = widget.payment;
  Timer? _timer;
  int _polls = 0;
  bool _checking = false;

  @override
  void initState() {
    super.initState();
    _schedulePoll();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  void _schedulePoll() {
    _timer?.cancel();
    if (_payment.status != BankilyPaymentStatus.pending || _polls >= widget.maxPolls) return;
    _timer = Timer(widget.pollInterval, () async {
      _polls++;
      try {
        final p = await _api.getPayment(_payment.id);
        if (mounted) setState(() => _payment = p);
      } catch (_) {
        // Keep showing the last known state; the next poll or a manual
        // check will refresh it.
      }
      if (mounted) _schedulePoll();
    });
  }

  Future<void> _checkNow() async {
    setState(() => _checking = true);
    try {
      final p = await _api.check(_payment.id);
      if (mounted) setState(() => _payment = p);
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(AppLocalizations.of(context)!.bpay_err_generic)),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _checking = false);
        _polls = 0;
        _schedulePoll();
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    const blue = Color(0xFF0084FF);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    late final IconData icon;
    late final Color color;
    late final String title;
    late final String body;
    switch (_payment.status) {
      case BankilyPaymentStatus.success:
        icon = Icons.check_circle;
        color = const Color(0xFF16A34A);
        title = l10n.bpay_success_title;
        body = l10n.bpay_success_body(_payment.amount.toStringAsFixed(0));
      case BankilyPaymentStatus.failed:
        icon = Icons.cancel;
        color = const Color(0xFFDC2626);
        title = l10n.bpay_failed_title;
        body = l10n.bpay_failed_body;
      case BankilyPaymentStatus.pending:
        icon = Icons.hourglass_top;
        color = const Color(0xFFF59E0B);
        title = l10n.bpay_pending_title;
        body = l10n.bpay_pending_body;
    }

    return PopScope(
      canPop: true,
      child: Scaffold(
        backgroundColor: isDark ? const Color(0xFF121212) : const Color(0xFFF5F5F5),
        appBar: AppBar(
          backgroundColor: blue,
          systemOverlayStyle: AppTheme.lightStatusIcons,
          elevation: 0,
          centerTitle: true,
          iconTheme: const IconThemeData(color: Colors.white),
          title: Text(l10n.bpay_title, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 18)),
        ),
        body: SafeArea(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: Column(
              children: [
                const SizedBox(height: 24),
                Icon(icon, key: ValueKey('bpay-status-${_payment.status.name}'), color: color, size: 88),
                const SizedBox(height: 16),
                Text(title, textAlign: TextAlign.center, style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold)),
                const SizedBox(height: 12),
                Text(body, textAlign: TextAlign.center, style: const TextStyle(fontSize: 15, height: 1.5)),
                if (_payment.status == BankilyPaymentStatus.pending) ...[
                  const SizedBox(height: 20),
                  const LinearProgressIndicator(minHeight: 3),
                ],
                const SizedBox(height: 28),
                _detailRow(l10n.bpay_amount_label, _payment.amount.toStringAsFixed(0), isDark),
                _detailRow(l10n.bpay_phone_label, _payment.clientPhoneMasked, isDark),
                _detailRow(l10n.bpay_operation_id, _payment.operationId, isDark),
                if (_payment.bankilyTransactionId != null && _payment.bankilyTransactionId!.isNotEmpty)
                  _detailRow(l10n.bpay_transaction_id, _payment.bankilyTransactionId!, isDark),
                const SizedBox(height: 28),
                if (_payment.status == BankilyPaymentStatus.pending)
                  _button(
                    label: l10n.bpay_check_again,
                    onPressed: _checking ? null : _checkNow,
                    busy: _checking,
                    color: blue,
                  ),
                if (_payment.status == BankilyPaymentStatus.failed) ...[
                  _button(label: l10n.bpay_try_again, onPressed: () => Navigator.of(context).pop('retry'), color: blue),
                  const SizedBox(height: 12),
                  _button(
                    label: l10n.bpay_use_manual,
                    onPressed: () => Navigator.of(context).pop('manual'),
                    color: Colors.grey.shade700,
                  ),
                ],
                if (_payment.status == BankilyPaymentStatus.success)
                  _button(
                    label: l10n.bpay_done,
                    onPressed: () => Navigator.of(context).popUntil((r) => r.isFirst),
                    color: blue,
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _detailRow(String label, String value, bool isDark) {
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: isDark ? const Color(0xFF1D1D1D) : Colors.white,
        borderRadius: BorderRadius.circular(10),
      ),
      child: Row(
        children: [
          Expanded(child: Text(label, style: TextStyle(color: Colors.grey.shade600, fontSize: 13))),
          Flexible(child: SelectableText(value, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14))),
        ],
      ),
    );
  }

  Widget _button({required String label, required VoidCallback? onPressed, required Color color, bool busy = false}) {
    return SizedBox(
      width: double.infinity,
      height: 52,
      child: ElevatedButton(
        onPressed: onPressed,
        style: ElevatedButton.styleFrom(
          backgroundColor: color,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          elevation: 0,
        ),
        child: busy
            ? const SizedBox(width: 22, height: 22, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
            : Text(label, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16)),
      ),
    );
  }
}
