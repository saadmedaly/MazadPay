/// Modèle Bid (Enchère) basé sur le backend Go
/// Correspond à `backend/internal/models/bid.go`
class Bid {
  final String id;
  final String auctionId;
  final String userId;
  final double amount;
  final double? previousPrice;
  final bool isWinning;
  final String? bidderName;
  final String? bidderPhone;
  final bool isAnonymous;
  final DateTime createdAt;

  Bid({
    required this.id,
    required this.auctionId,
    required this.userId,
    required this.amount,
    this.previousPrice,
    this.isWinning = false,
    this.bidderName,
    this.bidderPhone,
    this.isAnonymous = false,
    required this.createdAt,
  });

  factory Bid.fromJson(Map<String, dynamic> json) {
    return Bid(
      id: json['id'] ?? '',
      auctionId: json['auction_id'] ?? '',
      userId: json['user_id'] ?? '',
      amount: _parseDecimal(json['amount']),
      previousPrice: json['previous_price'] != null
          ? _parseDecimal(json['previous_price'])
          : null,
      isWinning: json['is_winning'] ?? false,
      bidderName: json['bidder_name'],
      bidderPhone: json['bidder_phone'],
      isAnonymous: json['is_anonymous'] ?? false,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'auction_id': auctionId,
      'user_id': userId,
      'amount': amount,
      'previous_price': previousPrice,
      'is_winning': isWinning,
      'bidder_name': bidderName,
      'bidder_phone': bidderPhone,
      'is_anonymous': isAnonymous,
      'created_at': createdAt.toIso8601String(),
    };
  }

  /// Parse un décimal
  static double _parseDecimal(dynamic value) {
    if (value == null) return 0.0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }

  /// Nom affiché de l'enchérisseur (masqué si anonyme)
  String get displayBidderName {
    if (isAnonymous) return 'Anonyme';
    return bidderName ?? 'Utilisateur';
  }

  /// Téléphone affiché (masqué)
  String get displayPhone {
    if (bidderPhone == null || bidderPhone!.length < 4) return '####';
    return '####${bidderPhone!.substring(bidderPhone!.length - 4)}';
  }

  /// Incrément de l'enchère par rapport au prix précédent
  double? get increment {
    if (previousPrice == null) return null;
    return amount - previousPrice!;
  }

  /// Formate la date relative
  String getTimeAgo() {
    final now = DateTime.now();
    final difference = now.difference(createdAt);

    if (difference.inDays > 0) {
      return '${difference.inDays}j';
    } else if (difference.inHours > 0) {
      return '${difference.inHours}h';
    } else if (difference.inMinutes > 0) {
      return '${difference.inMinutes}m';
    } else {
      return 'maintenant';
    }
  }
}

/// Historique des enchères pour une auction
class BidHistory {
  final String auctionId;
  final List<Bid> bids;
  final double currentPrice;
  final int totalBids;
  final String? highestBidderId;

  BidHistory({
    required this.auctionId,
    required this.bids,
    required this.currentPrice,
    required this.totalBids,
    this.highestBidderId,
  });

  factory BidHistory.fromJson(Map<String, dynamic> json) {
    final bidsList = json['bids'] as List<dynamic>? ?? [];
    return BidHistory(
      auctionId: json['auction_id'] ?? '',
      bids: bidsList.map((e) => Bid.fromJson(e as Map<String, dynamic>)).toList(),
      currentPrice: Bid._parseDecimal(json['current_price']),
      totalBids: json['total_bids'] ?? 0,
      highestBidderId: json['highest_bidder_id']?.toString(),
    );
  }

  /// Enchère la plus haute
  Bid? get highestBid => bids.isNotEmpty ? bids.first : null;

  /// Nombre d'enchérisseurs uniques
  int get uniqueBidders => bids.map((b) => b.userId).toSet().length;
}

/// Modèle pour placer une enchère
class PlaceBidRequest {
  final double amount;
  final bool? isAutoBid;
  final double? maxAutoBidAmount;

  PlaceBidRequest({
    required this.amount,
    this.isAutoBid,
    this.maxAutoBidAmount,
  });

  Map<String, dynamic> toJson() {
    return {
      'amount': amount,
      if (isAutoBid != null) 'is_auto_bid': isAutoBid,
      if (maxAutoBidAmount != null) 'max_auto_bid_amount': maxAutoBidAmount,
    };
  }
}

/// Réponse après placement d'une enchère
class PlaceBidResponse {
  final bool success;
  final String? bidId;
  final double? newPrice;
  final String? message;
  final String? errorCode;

  PlaceBidResponse({
    required this.success,
    this.bidId,
    this.newPrice,
    this.message,
    this.errorCode,
  });

  factory PlaceBidResponse.fromJson(Map<String, dynamic> json) {
    return PlaceBidResponse(
      success: json['success'] ?? false,
      bidId: json['bid_id']?.toString(),
      newPrice: json['new_price'] != null
          ? Bid._parseDecimal(json['new_price'])
          : null,
      message: json['message'],
      errorCode: json['error_code'],
    );
  }
}
