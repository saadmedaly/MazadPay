class Auction {
  final String id;
  final String title;
  final String description;
  final List<String> imageUrls;
  final double startPrice;
  final double currentPrice;
  final double minIncrement;
  final DateTime endTime;
  final int bidderCount;
  final int views;
  final String lotNumber;
  final String phoneNumber;
  final String sellerId;
  final bool isUserHighestBidder;

  // Car details
  final String? manufacturer;
  final String? fuelType;
  final String? transmission;
  final String? year;
  final String? mileage;
  final String? model;

  Auction({
    required this.id,
    required this.title,
    required this.description,
    required this.imageUrls,
    required this.startPrice,
    required this.currentPrice,
    required this.minIncrement,
    required this.endTime,
    required this.bidderCount,
    required this.views,
    required this.lotNumber,
    required this.phoneNumber,
    required this.sellerId,
    this.isUserHighestBidder = false,
    this.manufacturer,
    this.fuelType,
    this.transmission,
    this.year,
    this.mileage,
    this.model,
  });

  Auction copyWith({
    String? id,
    String? title,
    String? description,
    List<String>? imageUrls,
    double? startPrice,
    double? currentPrice,
    double? minIncrement,
    DateTime? endTime,
    int? bidderCount,
    int? views,
    String? lotNumber,
    String? phoneNumber,
    String? sellerId,
    bool? isUserHighestBidder,
    String? manufacturer,
    String? fuelType,
    String? transmission,
    String? year,
    String? mileage,
    String? model,
  }) {
    return Auction(
      id: id ?? this.id,
      title: title ?? this.title,
      description: description ?? this.description,
      imageUrls: imageUrls ?? this.imageUrls,
      startPrice: startPrice ?? this.startPrice,
      currentPrice: currentPrice ?? this.currentPrice,
      minIncrement: minIncrement ?? this.minIncrement,
      endTime: endTime ?? this.endTime,
      bidderCount: bidderCount ?? this.bidderCount,
      views: views ?? this.views,
      lotNumber: lotNumber ?? this.lotNumber,
      phoneNumber: phoneNumber ?? this.phoneNumber,
      sellerId: sellerId ?? this.sellerId,
      isUserHighestBidder: isUserHighestBidder ?? this.isUserHighestBidder,
      manufacturer: manufacturer ?? this.manufacturer,
      fuelType: fuelType ?? this.fuelType,
      transmission: transmission ?? this.transmission,
      year: year ?? this.year,
      mileage: mileage ?? this.mileage,
      model: model ?? this.model,
    );
  }

  factory Auction.fromJson(Map<String, dynamic> json) {
    List<String> images = [];
    if (json['images'] != null && json['images'] is List) {
      images = (json['images'] as List).map((img) {
        if (img is Map<String, dynamic>) return img['url']?.toString() ?? '';
        return img.toString();
      }).where((s) => s.isNotEmpty).toList();
    } else if (json['image_urls'] != null && json['image_urls'] is List) {
      images = (json['image_urls'] as List).map((e) => e.toString()).toList();
    }

    return Auction(
      id: json['id']?.toString() ?? '',
      title: json['title_ar']?.toString() ?? json['title']?.toString() ?? '',
      description: json['description_ar']?.toString() ?? json['description']?.toString() ?? '',
      imageUrls: images,
      startPrice: double.tryParse(json['starting_price']?.toString() ?? json['start_price']?.toString() ?? '0') ?? 0,
      currentPrice: double.tryParse(json['current_price']?.toString() ?? json['current_bid']?.toString() ?? '0') ?? 0,
      minIncrement: double.tryParse(json['min_increment']?.toString() ?? '500') ?? 500,
      endTime: json['end_time'] != null ? DateTime.parse(json['end_time']) : DateTime.now(),
      bidderCount: json['bidder_count'] ?? json['bid_count'] ?? 0,
      views: json['views'] ?? json['view_count'] ?? 0,
      lotNumber: json['lot_number']?.toString() ?? 'N/A',
      phoneNumber: json['seller_phone']?.toString() ?? json['phone_number']?.toString() ?? 'N/A',
      sellerId: json['seller_id']?.toString() ?? '',
      isUserHighestBidder: json['is_user_highest_bidder'] ?? false,
      manufacturer: json['manufacturer']?.toString(),
      fuelType: json['fuel_type']?.toString(),
      transmission: json['transmission']?.toString(),
      year: json['year']?.toString(),
      mileage: json['mileage']?.toString(),
      model: json['model']?.toString(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'description': description,
      'images': imageUrls,
      'starting_price': startPrice,
      'current_price': currentPrice,
      'min_increment': minIncrement,
      'end_time': endTime.toIso8601String(),
      'bidder_count': bidderCount,
      'views': views,
      'lot_number': lotNumber,
      'seller_phone': phoneNumber,
      'seller_id': sellerId,
      'is_user_highest_bidder': isUserHighestBidder,
      'manufacturer': manufacturer,
      'fuel_type': fuelType,
      'transmission': transmission,
      'year': year,
      'mileage': mileage,
      'model': model,
    };
  }
}

class BidEntry {
  final String bidderName;
  final String phoneNumber;
  final double amount;
  final DateTime timestamp;
  final bool isWinner;

  BidEntry({
    required this.bidderName,
    required this.phoneNumber,
    required this.amount,
    required this.timestamp,
    this.isWinner = false,
  });
}
