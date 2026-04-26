/// Modèle Wallet basé sur le backend Go
/// Correspond à `backend/internal/models/wallet.go`
class Wallet {
  final String userId;
  final double balance;
  final double frozenAmount;
  final int version;
  final DateTime updatedAt;

  Wallet({
    required this.userId,
    required this.balance,
    required this.frozenAmount,
    required this.version,
    required this.updatedAt,
  });

  factory Wallet.fromJson(Map<String, dynamic> json) {
    return Wallet(
      userId: json['user_id'] ?? '',
      balance: _parseDecimal(json['balance']),
      frozenAmount: _parseDecimal(json['frozen_amount']),
      version: json['version'] ?? 0,
      updatedAt: json['updated_at'] != null
          ? DateTime.parse(json['updated_at'])
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'user_id': userId,
      'balance': balance,
      'frozen_amount': frozenAmount,
      'version': version,
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  /// Solde disponible (balance - frozen)
  double get availableBalance => balance - frozenAmount;

  /// Parse un décimal depuis différents formats (String, num)
  static double _parseDecimal(dynamic value) {
    if (value == null) return 0.0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }
}

/// Types de transaction
class TransactionType {
  static const String deposit = 'deposit';
  static const String withdraw = 'withdraw';
  static const String bidHold = 'bid_hold';
  static const String bidRelease = 'bid_release';
  static const String payment = 'payment';
  static const String refund = 'refund';
  static const String fee = 'fee';
}

/// Statuts de transaction
class TransactionStatus {
  static const String pending = 'pending';
  static const String completed = 'completed';
  static const String failed = 'failed';
  static const String cancelled = 'cancelled';
  static const String underReview = 'under_review';
}

/// Modèle Transaction
class Transaction {
  final String id;
  final String userId;
  final String? auctionId;
  final String type;
  final double amount;
  final String? gateway;
  final String status;
  final String? reference;
  final String? receiptUrl;
  final String? adminNotes;
  final String? reviewedBy;
  final DateTime? reviewedAt;
  final String? walletHoldId;
  final String? receiptImageTemp;
  final String? paymentMethod;
  final double? feeAmount;
  final double? netAmount;
  final String? description;
  final String? failureReason;
  final DateTime createdAt;

  Transaction({
    required this.id,
    required this.userId,
    this.auctionId,
    required this.type,
    required this.amount,
    this.gateway,
    required this.status,
    this.reference,
    this.receiptUrl,
    this.adminNotes,
    this.reviewedBy,
    this.reviewedAt,
    this.walletHoldId,
    this.receiptImageTemp,
    this.paymentMethod,
    this.feeAmount,
    this.netAmount,
    this.description,
    this.failureReason,
    required this.createdAt,
  });

  factory Transaction.fromJson(Map<String, dynamic> json) {
    return Transaction(
      id: json['id'] ?? '',
      userId: json['user_id'] ?? '',
      auctionId: json['auction_id'],
      type: json['type'] ?? '',
      amount: Wallet._parseDecimal(json['amount']),
      gateway: json['gateway'],
      status: json['status'] ?? 'pending',
      reference: json['reference'],
      receiptUrl: json['receipt_url'],
      adminNotes: json['admin_notes'],
      reviewedBy: json['reviewed_by']?.toString(),
      reviewedAt: json['reviewed_at'] != null
          ? DateTime.parse(json['reviewed_at'])
          : null,
      walletHoldId: json['wallet_hold_id']?.toString(),
      receiptImageTemp: json['receipt_image_temp'],
      paymentMethod: json['payment_method'],
      feeAmount: json['fee_amount'] != null
          ? Wallet._parseDecimal(json['fee_amount'])
          : null,
      netAmount: json['net_amount'] != null
          ? Wallet._parseDecimal(json['net_amount'])
          : null,
      description: json['description'],
      failureReason: json['failure_reason'],
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'user_id': userId,
      'auction_id': auctionId,
      'type': type,
      'amount': amount,
      'gateway': gateway,
      'status': status,
      'reference': reference,
      'receipt_url': receiptUrl,
      'admin_notes': adminNotes,
      'reviewed_by': reviewedBy,
      'reviewed_at': reviewedAt?.toIso8601String(),
      'wallet_hold_id': walletHoldId,
      'receipt_image_temp': receiptImageTemp,
      'payment_method': paymentMethod,
      'fee_amount': feeAmount,
      'net_amount': netAmount,
      'description': description,
      'failure_reason': failureReason,
      'created_at': createdAt.toIso8601String(),
    };
  }

  /// Vérifie si c'est un crédit (argent entrant)
  bool get isCredit {
    return type == TransactionType.deposit ||
        type == TransactionType.bidRelease ||
        type == TransactionType.refund;
  }

  /// Vérifie si c'est un débit (argent sortant)
  bool get isDebit {
    return type == TransactionType.withdraw ||
        type == TransactionType.bidHold ||
        type == TransactionType.payment ||
        type == TransactionType.fee;
  }

  /// Montant signé (+ pour crédit, - pour débit)
  double get signedAmount => isCredit ? amount : -amount;

  /// Vérifie si la transaction est terminée avec succès
  bool get isCompleted => status == TransactionStatus.completed;

  /// Vérifie si la transaction est en attente
  bool get isPending => status == TransactionStatus.pending;
}
