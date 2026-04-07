class Auction {
  final String id;
  final String title;
  final String description;
  final List<String> imageUrls;
  final double startPrice;
  double currentPrice;
  final double minIncrement;
  final DateTime endTime;
  final int bidderCount;
  final int views;
  final String lotNumber;
  final String phoneNumber;
  bool isUserHighestBidder;

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
    this.isUserHighestBidder = false,
    this.manufacturer,
    this.fuelType,
    this.transmission,
    this.year,
    this.mileage,
    this.model,
  });
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
