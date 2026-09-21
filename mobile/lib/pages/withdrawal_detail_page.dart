import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:intl/intl.dart' hide TextDirection;
import 'package:mezadpay/models/wallet.dart';
import 'package:mezadpay/services/wallet_api.dart';
import 'package:mezadpay/utils/money_formatter.dart';

/// MAZADPAY — withdrawal notification routing bug: a withdrawal_processed
/// push previously routed to DepositPage (wrong -- that's a deposit screen).
/// This page is the correct destination: it shows the withdrawal's own
/// status, amount, beneficiary account, admin notes, and the admin-uploaded
/// transfer receipt (admin_attachment_url), fetched by the real transaction
/// id carried in the push payload's data.transaction_id.
class WithdrawalDetailPage extends StatefulWidget {
  final String transactionId;

  const WithdrawalDetailPage({super.key, required this.transactionId});

  @override
  State<WithdrawalDetailPage> createState() => _WithdrawalDetailPageState();
}

class _WithdrawalDetailPageState extends State<WithdrawalDetailPage> {
  Transaction? _transaction;
  bool _loading = true;
  bool _error = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final response = await WalletApi().getTransactionDetails(widget.transactionId);
      if (!mounted) return;
      if (!response.success || response.data == null) {
        setState(() {
          _loading = false;
          _error = true;
        });
        return;
      }
      setState(() {
        _transaction = Transaction.fromJson(response.data!);
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = true;
      });
    }
  }

  String _statusLabel(String status) {
    switch (status) {
      case 'completed':
        return 'مكتمل';
      case 'rejected':
      case 'failed':
        return 'مرفوض';
      case 'cancelled':
        return 'ملغى';
      case 'under_review':
        return 'قيد المراجعة';
      case 'pending':
      default:
        return 'قيد الانتظار';
    }
  }

  void _openFullscreenImage(String url) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => Scaffold(
          backgroundColor: Colors.black,
          appBar: AppBar(
            backgroundColor: Colors.black,
            iconTheme: const IconThemeData(color: Colors.white),
          ),
          body: Center(
            child: InteractiveViewer(
              child: CachedNetworkImage(
                imageUrl: url,
                fit: BoxFit.contain,
                errorWidget: (context, url, error) => const Icon(
                  Icons.image_not_supported,
                  color: Colors.white54,
                  size: 60,
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Directionality(
      textDirection: TextDirection.rtl,
      child: Scaffold(
        backgroundColor: isDark ? const Color(0xFF121212) : const Color(0xFFFBFBFB),
        appBar: AppBar(
          backgroundColor: Colors.transparent,
          elevation: 0,
          centerTitle: true,
          leading: IconButton(
            icon: Icon(Icons.arrow_forward_ios, color: isDark ? Colors.white : Colors.black, size: 20),
            onPressed: () => Navigator.of(context).pop(),
          ),
          title: Text(
            'تفاصيل طلب السحب',
            style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18, color: isDark ? Colors.white : Colors.black),
          ),
        ),
        body: _loading
            ? const Center(child: CircularProgressIndicator())
            : _error || _transaction == null
                ? Center(
                    child: Text(
                      'تعذر تحميل تفاصيل طلب السحب',
                      style: TextStyle(color: Colors.grey[600], fontSize: 15),
                    ),
                  )
                : SingleChildScrollView(
                    padding: const EdgeInsets.all(20),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        _buildRow('الحالة', _statusLabel(_transaction!.status), isDark),
                        const SizedBox(height: 16),
                        _buildRow(
                          'المبلغ',
                          MoneyFormatter.format(_transaction!.amount, _transaction!.currencyCode),
                          isDark,
                        ),
                        if (_transaction!.beneficiaryAccount != null && _transaction!.beneficiaryAccount!.trim().isNotEmpty) ...[
                          const SizedBox(height: 16),
                          _buildRow('رقم المستفيد', _transaction!.beneficiaryAccount!, isDark),
                        ],
                        const SizedBox(height: 16),
                        _buildRow(
                          'التاريخ',
                          DateFormat('yyyy-MM-dd HH:mm').format(_transaction!.createdAt.toLocal()),
                          isDark,
                        ),
                        if (_transaction!.adminNotes != null && _transaction!.adminNotes!.trim().isNotEmpty) ...[
                          const SizedBox(height: 24),
                          Text(
                            'ملاحظات المسؤول',
                            style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: isDark ? Colors.white : Colors.black),
                          ),
                          const SizedBox(height: 8),
                          Container(
                            width: double.infinity,
                            padding: const EdgeInsets.all(14),
                            decoration: BoxDecoration(
                              color: isDark ? const Color(0xFF1D1D1D) : Colors.white,
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(color: Colors.grey.withValues(alpha: 0.2)),
                            ),
                            child: Text(
                              _transaction!.adminNotes!,
                              style: TextStyle(fontSize: 14, color: isDark ? Colors.grey[300] : Colors.grey[800]),
                            ),
                          ),
                        ],
                        if (_transaction!.adminAttachmentUrl != null && _transaction!.adminAttachmentUrl!.trim().isNotEmpty) ...[
                          const SizedBox(height: 24),
                          Text(
                            'إيصال التحويل من المسؤول',
                            style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: isDark ? Colors.white : Colors.black),
                          ),
                          const SizedBox(height: 8),
                          GestureDetector(
                            onTap: () => _openFullscreenImage(_transaction!.adminAttachmentUrl!),
                            child: ClipRRect(
                              borderRadius: BorderRadius.circular(16),
                              child: CachedNetworkImage(
                                imageUrl: _transaction!.adminAttachmentUrl!,
                                width: double.infinity,
                                height: 220,
                                fit: BoxFit.cover,
                                placeholder: (context, url) => Container(
                                  width: double.infinity,
                                  height: 220,
                                  color: Colors.grey[300],
                                  child: const Center(child: CircularProgressIndicator(strokeWidth: 2)),
                                ),
                                errorWidget: (context, url, error) => Container(
                                  width: double.infinity,
                                  height: 220,
                                  color: Colors.grey[300],
                                  child: const Icon(Icons.image_not_supported, color: Colors.grey, size: 40),
                                ),
                              ),
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
      ),
    );
  }

  Widget _buildRow(String label, String value, bool isDark) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label, style: TextStyle(color: Colors.grey[600], fontSize: 14)),
        Text(
          value,
          style: TextStyle(fontWeight: FontWeight.bold, fontSize: 15, color: isDark ? Colors.white : Colors.black),
        ),
      ],
    );
  }
}
