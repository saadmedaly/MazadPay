import 'package:riverpod_annotation/riverpod_annotation.dart';
import '../models/auction.dart';
import '../services/auction_api.dart';

part 'auction_provider_api.g.dart';

/// Provider pour les enchères utilisant l'API backend
/// Remplace le provider mocké avec de vraies données
@riverpod
class AuctionNotifierApi extends _$AuctionNotifierApi {
  final AuctionApi _auctionApi = AuctionApi();
  
  @override
  Future<Auction> build(String id) async {
    try {
      final response = await _auctionApi.getAuctionById(id);
      
      if (response.success && response.data != null) {
        // L'API retourne {"data": {"auction": {...}, "images": [...]}}
        final responseData = response.data!;
        if (responseData is Map<String, dynamic>) {
          // Extraire l'enchère (peut être dans 'auction' ou directement dans 'data')
          final dynamic auctionRaw = responseData['auction'] ?? responseData;
          if (auctionRaw is Map<String, dynamic>) {
            // Créer une copie pour éviter de modifier l'original
            final auctionData = Map<String, dynamic>.from(auctionRaw);
            // Fusionner les images si présentes séparément
            if (responseData['images'] != null) {
              auctionData['images'] = responseData['images'];
            }
            return _mapToAuction(auctionData);
          }
        }
        // Fallback: essayer de mapper directement
        if (responseData is Map<String, dynamic>) {
          return _mapToAuction(responseData);
        }
      } else {
        // En cas d'erreur, retourner une enchère vide par défaut
        return _getDefaultAuction(id);
      }
    } catch (e) {
      // En cas d'erreur, retourner une enchère vide par défaut
      return _getDefaultAuction(id);
    }
  }
  
  /// Convertir la réponse API en modèle Auction
  Auction _mapToAuction(Map<String, dynamic> data) {
    // Gérer les images - format objet (detail API) ou string (list API)
    List<String> imageUrls = ['assets/corolla.png'];
    if (data['images'] != null && data['images'] is List) {
      imageUrls = (data['images'] as List).map((img) {
        if (img is Map<String, dynamic>) {
          return img['url']?.toString() ?? 'assets/corolla.png';
        } else {
          return img.toString();
        }
      }).toList();
    }
    
    return Auction(
      id: data['id']?.toString() ?? '',
      title: data['title_ar'] ?? data['title'] ?? 'Unknown',
      description: data['description_ar'] ?? data['description'] ?? '',
      imageUrls: imageUrls.isNotEmpty ? imageUrls : ['assets/corolla.png'],
      startPrice: (data['starting_price'] ?? 0).toDouble(),
      currentPrice: (data['current_price'] ?? data['starting_price'] ?? 0).toDouble(),
      minIncrement: (data['min_increment'] ?? 500).toDouble(),
      endTime: data['end_time'] != null 
          ? DateTime.parse(data['end_time']) 
          : DateTime.now().add(const Duration(hours: 13)),
      bidderCount: data['bidder_count'] ?? data['bid_count'] ?? 0,
      views: data['views'] ?? 0,
      lotNumber: data['lot_number'] ?? 'N/A',
      phoneNumber: data['seller_phone'] ?? 'N/A',
      manufacturer: data['manufacturer'] ?? '',
      fuelType: data['fuel_type'] ?? '',
      transmission: data['transmission'] ?? '', 
      year: data['year']?.toString() ?? '',
      mileage: data['mileage']?.toString() ?? '',
      model: data['model'] ?? '',
    );
  }
  
  /// Enchère par défaut en cas d'erreur
  Auction _getDefaultAuction(String id) {
    return Auction(
      id: id,
      title: 'Chargement...',
      description: '',
      imageUrls: ['assets/corolla.png'],
      startPrice: 0,
      currentPrice: 0,
      minIncrement: 500,
      endTime: DateTime.now().add(const Duration(hours: 13)),
      bidderCount: 0,
      views: 0,
      lotNumber: 'N/A',
      phoneNumber: 'N/A',
      manufacturer: '',
      fuelType: '',
      transmission: '',
      year: '',
      mileage: '',
      model: '',
    );
  }
  
}

@riverpod
class AuctionHistoryApi extends _$AuctionHistoryApi {
  final AuctionApi _auctionApi = AuctionApi();
  
  @override
  Future<List<BidEntry>> build(String auctionId) async {
    try {
      final response = await _auctionApi.getBidHistory(auctionId);
      
      if (response.success && response.data != null) {
        final bids = response.data! as List;
        return bids.map((bid) => _mapToBidEntry(bid as Map<String, dynamic>)).toList();
      }
      return [];
    } catch (e) {
      return [];
    }
  }
  
  /// Convertir la réponse API en BidEntry
  BidEntry _mapToBidEntry(Map<String, dynamic> data) {
    return BidEntry(
      bidderName: data['bidder_name'] ?? data['user_name'] ?? 'Unknown',
      phoneNumber: data['bidder_phone'] ?? data['phone'] ?? '',
      amount: (data['amount'] ?? 0).toDouble(),
      timestamp: data['created_at'] != null
          ? DateTime.parse(data['created_at'])
          : DateTime.now(),
      isWinner: data['is_winning'] ?? data['is_winner'] ?? false,
    );
  }

  /// Rafraîchir la liste des bids depuis l'API
  Future<void> refresh(String auctionId) async {
    state = const AsyncValue.loading();
    try {
      final response = await _auctionApi.getBidHistory(auctionId);
      if (response.success && response.data != null) {
        final bids = response.data! as List;
        final bidList = bids.map((bid) => _mapToBidEntry(bid as Map<String, dynamic>)).toList();
        state = AsyncValue.data(bidList);
      } else {
        state = AsyncValue.data([]);
      }
    } catch (e) {
      state = AsyncValue.error(e, StackTrace.current);
    }
  }
}
