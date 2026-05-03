/// Modèle Auction complet basé sur le backend Go
/// Correspond à `backend/internal/models/auction.go`
/// Remplace et étend le modèle auction.dart existant
class AuctionModel {
  final String id;
  final String sellerId;
  final int categoryId;
  final int? subCategoryId;
  final int? locationId;
  final String titleAr;
  final String? titleFr;
  final String? titleEn;
  final String? descriptionAr;
  final String? descriptionFr;
  final String? descriptionEn;
  final double startPrice;
  final double currentPrice;
  final double minIncrement;
  final double insuranceAmount;
  final double reservePrice;
  final DateTime startTime;
  final DateTime endTime;
  final String status;
  final String? lotNumber;
  final int views;
  final int bidderCount;
  final String? winnerId;
  final String? winningBidId;
  final DateTime? paymentDeadline;
  final bool isFeatured;
  final DateTime? featuredUntil;
  final String? rejectionReason;
  final String? phoneContact;
  final Map<String, dynamic>? itemDetails;
  final double? buyNowPrice;
  final int version;
  final DateTime createdAt;
  final String? condition;
  final String? brand;
  final bool isVerified;
  final DateTime? boostedUntil;
  final String? videoUrl;
  final int quantity;
  final String? categoryNameAr;
  final String? cityNameAr;
  final List<AuctionImage> images;

  // Propriétés calculées
  final bool isUserHighestBidder;

  AuctionModel({
    required this.id,
    required this.sellerId,
    required this.categoryId,
    this.subCategoryId,
    this.locationId,
    required this.titleAr,
    this.titleFr,
    this.titleEn,
    this.descriptionAr,
    this.descriptionFr,
    this.descriptionEn,
    required this.startPrice,
    required this.currentPrice,
    required this.minIncrement,
    required this.insuranceAmount,
    required this.reservePrice,
    required this.startTime,
    required this.endTime,
    required this.status,
    this.lotNumber,
    this.views = 0,
    this.bidderCount = 0,
    this.winnerId,
    this.winningBidId,
    this.paymentDeadline,
    this.isFeatured = false,
    this.featuredUntil,
    this.rejectionReason,
    this.phoneContact,
    this.itemDetails,
    this.buyNowPrice,
    this.version = 1,
    required this.createdAt,
    this.condition,
    this.brand,
    this.isVerified = false,
    this.boostedUntil,
    this.videoUrl,
    this.quantity = 1,
    this.categoryNameAr,
    this.cityNameAr,
    this.images = const [],
    this.isUserHighestBidder = false,
  });

  factory AuctionModel.fromJson(Map<String, dynamic> json) {
    final imagesList = json['images'] as List<dynamic>? ?? [];
    return AuctionModel(
      id: json['id'] ?? '',
      sellerId: json['seller_id'] ?? '',
      categoryId: json['category_id'] ?? 0,
      subCategoryId: json['sub_category_id'],
      locationId: json['location_id'],
      titleAr: json['title_ar'] ?? '',
      titleFr: json['title_fr'],
      titleEn: json['title_en'],
      descriptionAr: json['description_ar'],
      descriptionFr: json['description_fr'],
      descriptionEn: json['description_en'],
      startPrice: _parseDecimal(json['start_price']),
      currentPrice: _parseDecimal(json['current_price']),
      minIncrement: _parseDecimal(json['min_increment']),
      insuranceAmount: _parseDecimal(json['insurance_amount']),
      reservePrice: _parseDecimal(json['reserve_price']),
      startTime: json['start_time'] != null
          ? DateTime.parse(json['start_time'])
          : DateTime.now(),
      endTime: json['end_time'] != null
          ? DateTime.parse(json['end_time'])
          : DateTime.now().add(const Duration(days: 7)),
      status: json['status'] ?? 'active',
      lotNumber: json['lot_number'],
      views: json['views'] ?? 0,
      bidderCount: json['bidder_count'] ?? 0,
      winnerId: json['winner_id']?.toString(),
      winningBidId: json['winning_bid_id']?.toString(),
      paymentDeadline: json['payment_deadline'] != null
          ? DateTime.parse(json['payment_deadline'])
          : null,
      isFeatured: json['is_featured'] ?? false,
      featuredUntil: json['featured_until'] != null
          ? DateTime.parse(json['featured_until'])
          : null,
      rejectionReason: json['rejection_reason'],
      phoneContact: json['phone_contact'],
      itemDetails: json['item_details'] != null
          ? Map<String, dynamic>.from(json['item_details'])
          : null,
      buyNowPrice: json['buy_now_price'] != null
          ? _parseDecimal(json['buy_now_price'])
          : null,
      version: json['version'] ?? 1,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'])
          : DateTime.now(),
      condition: json['condition'],
      brand: json['brand'],
      isVerified: json['is_verified'] ?? false,
      boostedUntil: json['boosted_until'] != null
          ? DateTime.parse(json['boosted_until'])
          : null,
      videoUrl: json['video_url'],
      quantity: json['quantity'] ?? 1,
      categoryNameAr: json['category'],
      cityNameAr: json['city'],
      images: imagesList
          .map((e) => AuctionImage.fromJson(e as Map<String, dynamic>))
          .toList(),
      isUserHighestBidder: json['is_user_highest_bidder'] ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'seller_id': sellerId,
      'category_id': categoryId,
      'sub_category_id': subCategoryId,
      'location_id': locationId,
      'title_ar': titleAr,
      'title_fr': titleFr,
      'title_en': titleEn,
      'description_ar': descriptionAr,
      'description_fr': descriptionFr,
      'description_en': descriptionEn,
      'start_price': startPrice,
      'current_price': currentPrice,
      'min_increment': minIncrement,
      'insurance_amount': insuranceAmount,
      'reserve_price': reservePrice,
      'start_time': startTime.toIso8601String(),
      'end_time': endTime.toIso8601String(),
      'status': status,
      'lot_number': lotNumber,
      'views': views,
      'bidder_count': bidderCount,
      'winner_id': winnerId,
      'winning_bid_id': winningBidId,
      'payment_deadline': paymentDeadline?.toIso8601String(),
      'is_featured': isFeatured,
      'featured_until': featuredUntil?.toIso8601String(),
      'rejection_reason': rejectionReason,
      'phone_contact': phoneContact,
      'item_details': itemDetails,
      'buy_now_price': buyNowPrice,
      'version': version,
      'created_at': createdAt.toIso8601String(),
      'condition': condition,
      'brand': brand,
      'is_verified': isVerified,
      'boosted_until': boostedUntil?.toIso8601String(),
      'video_url': videoUrl,
      'quantity': quantity,
      'category': categoryNameAr,
      'city': cityNameAr,
      'images': images.map((e) => e.toJson()).toList(),
      'is_user_highest_bidder': isUserHighestBidder,
    };
  }

  /// Parse un décimal
  static double _parseDecimal(dynamic value) {
    if (value == null) return 0.0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }

  /// Récupère le titre dans la langue spécifiée
  String getTitle(String languageCode) {
    switch (languageCode) {
      case 'ar':
        return titleAr;
      case 'fr':
        return titleFr ?? titleAr;
      case 'en':
        return titleEn ?? titleAr;
      default:
        return titleAr;
    }
  }

  /// Récupère la description dans la langue spécifiée
  String? getDescription(String languageCode) {
    switch (languageCode) {
      case 'ar':
        return descriptionAr;
      case 'fr':
        return descriptionFr ?? descriptionAr;
      case 'en':
        return descriptionEn ?? descriptionAr;
      default:
        return descriptionAr;
    }
  }

  /// Prix minimum pour la prochaine enchère
  double get minBidAmount => currentPrice + minIncrement;

  /// Vérifie si l'enchère est active
  bool get isActive => status == 'active';

  /// Vérifie si l'enchère est terminée
  bool get isEnded => status == 'ended' || status == 'closed';

  /// Vérifie si l'enchère est en attente de validation
  bool get isPending => status == 'pending';

  /// Vérifie si l'enchère a un prix "Acheter maintenant"
  bool get hasBuyNow => buyNowPrice != null && buyNowPrice! > 0;

  /// Temps restant avant fin
  Duration get timeRemaining => endTime.difference(DateTime.now());

  /// Vérifie si l'enchère est en cours (active et pas encore terminée)
  bool get isOngoing => isActive && timeRemaining.inSeconds > 0;

  /// Image principale (première image)
  String? get mainImageUrl => images.isNotEmpty ? images.first.url : null;

  /// Copie avec modifications
  AuctionModel copyWith({
    String? id,
    String? sellerId,
    int? categoryId,
    int? subCategoryId,
    int? locationId,
    String? titleAr,
    String? titleFr,
    String? titleEn,
    String? descriptionAr,
    String? descriptionFr,
    String? descriptionEn,
    double? startPrice,
    double? currentPrice,
    double? minIncrement,
    double? insuranceAmount,
    double? reservePrice,
    DateTime? startTime,
    DateTime? endTime,
    String? status,
    String? lotNumber,
    int? views,
    int? bidderCount,
    String? winnerId,
    String? winningBidId,
    DateTime? paymentDeadline,
    bool? isFeatured,
    DateTime? featuredUntil,
    String? rejectionReason,
    String? phoneContact,
    Map<String, dynamic>? itemDetails,
    double? buyNowPrice,
    int? version,
    DateTime? createdAt,
    String? condition,
    String? brand,
    bool? isVerified,
    DateTime? boostedUntil,
    String? videoUrl,
    int? quantity,
    String? categoryNameAr,
    String? cityNameAr,
    List<AuctionImage>? images,
    bool? isUserHighestBidder,
  }) {
    return AuctionModel(
      id: id ?? this.id,
      sellerId: sellerId ?? this.sellerId,
      categoryId: categoryId ?? this.categoryId,
      subCategoryId: subCategoryId ?? this.subCategoryId,
      locationId: locationId ?? this.locationId,
      titleAr: titleAr ?? this.titleAr,
      titleFr: titleFr ?? this.titleFr,
      titleEn: titleEn ?? this.titleEn,
      descriptionAr: descriptionAr ?? this.descriptionAr,
      descriptionFr: descriptionFr ?? this.descriptionFr,
      descriptionEn: descriptionEn ?? this.descriptionEn,
      startPrice: startPrice ?? this.startPrice,
      currentPrice: currentPrice ?? this.currentPrice,
      minIncrement: minIncrement ?? this.minIncrement,
      insuranceAmount: insuranceAmount ?? this.insuranceAmount,
      reservePrice: reservePrice ?? this.reservePrice,
      startTime: startTime ?? this.startTime,
      endTime: endTime ?? this.endTime,
      status: status ?? this.status,
      lotNumber: lotNumber ?? this.lotNumber,
      views: views ?? this.views,
      bidderCount: bidderCount ?? this.bidderCount,
      winnerId: winnerId ?? this.winnerId,
      winningBidId: winningBidId ?? this.winningBidId,
      paymentDeadline: paymentDeadline ?? this.paymentDeadline,
      isFeatured: isFeatured ?? this.isFeatured,
      featuredUntil: featuredUntil ?? this.featuredUntil,
      rejectionReason: rejectionReason ?? this.rejectionReason,
      phoneContact: phoneContact ?? this.phoneContact,
      itemDetails: itemDetails ?? this.itemDetails,
      buyNowPrice: buyNowPrice ?? this.buyNowPrice,
      version: version ?? this.version,
      createdAt: createdAt ?? this.createdAt,
      condition: condition ?? this.condition,
      brand: brand ?? this.brand,
      isVerified: isVerified ?? this.isVerified,
      boostedUntil: boostedUntil ?? this.boostedUntil,
      videoUrl: videoUrl ?? this.videoUrl,
      quantity: quantity ?? this.quantity,
      categoryNameAr: categoryNameAr ?? this.categoryNameAr,
      cityNameAr: cityNameAr ?? this.cityNameAr,
      images: images ?? this.images,
      isUserHighestBidder: isUserHighestBidder ?? this.isUserHighestBidder,
    );
  }
}

/// Statuts d'enchère
class AuctionStatus {
  static const String pending = 'pending';
  static const String active = 'active';
  static const String ended = 'ended';
  static const String closed = 'closed';
  static const String cancelled = 'cancelled';
  static const String rejected = 'rejected';
}

/// Modèle Image d'enchère
class AuctionImage {
  final int id;
  final String auctionId;
  final String url;
  final String mediaType;
  final int displayOrder;

  AuctionImage({
    required this.id,
    required this.auctionId,
    required this.url,
    this.mediaType = 'image',
    this.displayOrder = 0,
  });

  factory AuctionImage.fromJson(Map<String, dynamic> json) {
    return AuctionImage(
      id: json['id'] ?? 0,
      auctionId: json['auction_id']?.toString() ?? '',
      url: json['url'] ?? '',
      mediaType: json['media_type'] ?? 'image',
      displayOrder: json['display_order'] ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'auction_id': auctionId,
      'url': url,
      'media_type': mediaType,
      'display_order': displayOrder,
    };
  }
}

/// Modèle pour créer une enchère
class CreateAuctionRequest {
  final String title;
  final String description;
  final double startingPrice;
  final int categoryId;
  final int? locationId;
  final DateTime endTime;
  final List<String> images;
  final double? buyNowPrice;
  final double? reservePrice;
  final Map<String, dynamic>? itemDetails;

  CreateAuctionRequest({
    required this.title,
    required this.description,
    required this.startingPrice,
    required this.categoryId,
    this.locationId,
    required this.endTime,
    required this.images,
    this.buyNowPrice,
    this.reservePrice,
    this.itemDetails,
  });

  Map<String, dynamic> toJson() {
    return {
      'title': title,
      'description': description,
      'starting_price': startingPrice,
      'category_id': categoryId,
      'location_id': locationId,
      'end_time': endTime.toIso8601String(),
      'images': images,
      if (buyNowPrice != null) 'buy_now_price': buyNowPrice,
      if (reservePrice != null) 'reserve_price': reservePrice,
      if (itemDetails != null) 'item_details': itemDetails,
    };
  }
}
