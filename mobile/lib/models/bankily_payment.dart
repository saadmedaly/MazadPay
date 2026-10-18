/// Bankily B-PAY automatic top-up attempt, as returned by
/// /users/wallet/bankily/*. The backend only exposes PENDING_VERIFICATION,
/// SUCCESS and FAILED to the app (internal steps are folded into pending);
/// the phone arrives already masked.
enum BankilyPaymentStatus { pending, success, failed }

class BankilyPayment {
  final String id;
  final String operationId;
  final double amount;
  final String currencyCode;
  final String clientPhoneMasked;
  final BankilyPaymentStatus status;
  final String? bankilyTransactionId;
  final String? failureReason;
  final DateTime? createdAt;
  final DateTime? completedAt;

  const BankilyPayment({
    required this.id,
    required this.operationId,
    required this.amount,
    required this.currencyCode,
    required this.clientPhoneMasked,
    required this.status,
    this.bankilyTransactionId,
    this.failureReason,
    this.createdAt,
    this.completedAt,
  });

  static BankilyPaymentStatus parseStatus(String? raw) {
    switch (raw) {
      case 'SUCCESS':
        return BankilyPaymentStatus.success;
      case 'FAILED':
        return BankilyPaymentStatus.failed;
      default:
        return BankilyPaymentStatus.pending;
    }
  }

  factory BankilyPayment.fromJson(Map<String, dynamic> json) {
    return BankilyPayment(
      id: json['id']?.toString() ?? '',
      operationId: json['operation_id']?.toString() ?? '',
      amount: double.tryParse(json['amount']?.toString() ?? '') ?? 0,
      currencyCode: json['currency_code']?.toString() ?? 'MRU',
      clientPhoneMasked: json['client_phone']?.toString() ?? '',
      status: parseStatus(json['status']?.toString()),
      bankilyTransactionId: json['bankily_transaction_id']?.toString(),
      failureReason: json['failure_reason']?.toString(),
      createdAt: DateTime.tryParse(json['created_at']?.toString() ?? ''),
      completedAt: DateTime.tryParse(json['completed_at']?.toString() ?? ''),
    );
  }
}

/// Public B-PAY settings served by the backend (never any credential).
class BankilyConfig {
  final bool enabled;
  final String merchantCode;
  final double minAmount;
  final double maxAmount;
  final bool allowDecimals;

  const BankilyConfig({
    required this.enabled,
    required this.merchantCode,
    required this.minAmount,
    required this.maxAmount,
    required this.allowDecimals,
  });

  static const disabled = BankilyConfig(enabled: false, merchantCode: '', minAmount: 100, maxAmount: 100000, allowDecimals: false);

  factory BankilyConfig.fromJson(Map<String, dynamic> json) {
    return BankilyConfig(
      enabled: json['enabled'] == true,
      merchantCode: json['merchant_code']?.toString() ?? '',
      minAmount: (json['min_amount'] as num?)?.toDouble() ?? 100,
      maxAmount: (json['max_amount'] as num?)?.toDouble() ?? 100000,
      allowDecimals: json['allow_decimals'] == true,
    );
  }
}

/// Client-side input checks mirroring the backend (which stays authoritative).
class BankilyInputRules {
  static final RegExp _phone = RegExp(r'^[234][0-9]{7}$');

  static String normalizePhone(String raw) {
    var p = raw.replaceAll(RegExp(r'[\s-]'), '');
    if (p.startsWith('+222')) p = p.substring(4);
    if (p.startsWith('00222')) p = p.substring(5);
    return p;
  }

  static bool isValidPhone(String raw) => _phone.hasMatch(normalizePhone(raw));

  static bool isValidPasscode(String raw) => RegExp(r'^[A-Za-z0-9]{1,32}$').hasMatch(raw.trim());

  static bool isValidAmount(double? amount, BankilyConfig cfg) {
    if (amount == null || amount <= 0) return false;
    if (!cfg.allowDecimals && amount != amount.truncateToDouble()) return false;
    return amount >= cfg.minAmount && amount <= cfg.maxAmount;
  }
}
