import 'package:mezadpay/l10n/app_localizations.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import '../models/auction.dart';

part 'auction_provider.g.dart';

@riverpod
class AuctionNotifier extends _$AuctionNotifier {
  @override
  Auction build(String id) {
    // Dummy initial data for the demo
    return Auction(
      id: id,
      title: 'Toyota Corolla 2018',
      description: 'سيارة نظيفة جدا، صيانة دورية، محرك ممتاز.',
      imageUrls: ['assets/corolla.png', 'assets/corolla.png', 'assets/corolla.png'], // Simulated multi-images
      startPrice: 300000,
      currentPrice: 307000,
      minIncrement: 500,
      endTime: DateTime.now().add(const Duration(hours: 13, minutes: 50, seconds: 23)),
      bidderCount: 5,
      views: 106,
      lotNumber: '49495',
      phoneNumber: '36601175',
      manufacturer: 'تويوتا',
      fuelType: 'بنزين',
      transmission: 'أوتوماتيكي',
      year: '2011',
      mileage: '257000',
      model: 'تويوتا برادو',
    );
  }

  void placeBid(double amount) {
    state = Auction(
      id: state.id,
      title: state.title,
      description: state.description,
      imageUrls: state.imageUrls,
      startPrice: state.startPrice,
      currentPrice: amount,
      minIncrement: state.minIncrement,
      endTime: state.endTime,
      bidderCount: state.bidderCount + 1,
      views: state.views,
      lotNumber: state.lotNumber,
      phoneNumber: state.phoneNumber,
      isUserHighestBidder: true,
    );
  }
}

@riverpod
class AuctionHistory extends _$AuctionHistory {
  @override
  List<BidEntry> build(String auctionId) {
    return [
      BidEntry(bidderName: 'محمد احمد', phoneNumber: '####4709', amount: 307000, timestamp: DateTime.now().subtract(const Duration(minutes: 10)), isWinner: true),
      BidEntry(bidderName: 'مريم سيدي', phoneNumber: '####4709', amount: 306500, timestamp: DateTime.now().subtract(const Duration(minutes: 39))),
      BidEntry(bidderName: 'علي صمبا', phoneNumber: '####4709', amount: 306000, timestamp: DateTime.now().subtract(const Duration(minutes: 40))),
      BidEntry(bidderName: 'محمد احمد', phoneNumber: '####4709', amount: 305500, timestamp: DateTime.now().subtract(const Duration(hours: 1, minutes: 52))),
    ];
  }

  void addBid(BidEntry bid) {
    state = [bid, ...state.map((e) => BidEntry(bidderName: e.bidderName, phoneNumber: e.phoneNumber, amount: e.amount, timestamp: e.timestamp, isWinner: false))];
  }
}
