import 'package:flutter/foundation.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import '../models/auction.dart';
import '../services/auction_api.dart';
import '../services/cache_service.dart';
import '../services/websocket_service.dart';

part 'auction_provider_api.g.dart';

/// Provider pour les enchères utilisant l'API backend
/// Remplace le provider mocké avec de vraies données
@riverpod
class AuctionNotifierApi extends _$AuctionNotifierApi {
  final AuctionApi _auctionApi = AuctionApi();
  
  @override
  Future<Auction> build(String id) async {
    // 1. Essayer de charger depuis le cache d'abord (Immédiat)
    final cachedData = await CacheService.instance.getCachedAuctionDetail(id);
    if (cachedData != null) {
      // Si on a du cache, on retourne ça tout de suite
      // On lance le refresh en arrière-plan
      _refreshInBackground(id);
      _listenToWebsocket(id);
      return _mapToAuction(cachedData);
    }

    // 2. Sinon, charger depuis l'API normalement
    _listenToWebsocket(id);
    return _fetchFromApi(id);
  }

  Future<void> _refreshInBackground(String id) async {
    try {
      final auction = await _fetchFromApi(id);
      state = AsyncValue.data(auction);
    } catch (e) {
      // Silencieux car on a déjà les données du cache
      debugPrint('Background refresh failed for auction $id: $e');
    }
  }

  Future<Auction> _fetchFromApi(String id) async {
    final response = await _auctionApi.getAuctionById(id);
    
    if (response.success && response.data != null) {
      final responseData = response.data!;
      Map<String, dynamic> auctionData;
      
      final dynamic auctionRaw = responseData['auction'] ?? responseData;
      if (auctionRaw is Map<String, dynamic>) {
        auctionData = Map<String, dynamic>.from(auctionRaw);
        if (responseData['images'] != null) {
          auctionData['images'] = responseData['images'];
        }
      } else {
        auctionData = responseData;
      }
    
      // Mettre en cache pour la prochaine fois
      await CacheService.instance.cacheAuctionDetail(id, auctionData);
      
      return _mapToAuction(auctionData);
    } else {
      throw Exception(response.message ?? 'Failed to load auction');
    }
  }

  void _listenToWebsocket(String id) {
    final wsService = WebsocketService();
    wsService.connect(id);
    
    // Écouter les mises à jour
    ref.onDispose(() {
      // Pas besoin de fermer ici car WebsocketService est un singleton 
      // mais on pourrait arrêter l'écoute spécifique si nécessaire
    });

    wsService.stream.listen((data) {
      if (data['type'] == 'bid_placed' || data['type'] == 'auction_update') {
        final update = data['data'];
        if (update != null && update is Map<String, dynamic>) {
          _handleWsUpdate(update);
        }
      }
    });
  }

  void _handleWsUpdate(Map<String, dynamic> update) {
    state.whenData((currentAuction) {
      // Mettre à jour uniquement les champs qui ont changé
      final updatedAuction = currentAuction.copyWith(
        currentPrice: (update['current_price'] ?? currentAuction.currentPrice).toDouble(),
        bidderCount: update['bid_count'] ?? update['bidder_count'] ?? currentAuction.bidderCount,
      );
      state = AsyncValue.data(updatedAuction);
      
      // Mettre à jour le cache aussi
      CacheService.instance.getCachedAuctionDetail(currentAuction.id).then((cached) {
        if (cached != null) {
          cached['current_price'] = updatedAuction.currentPrice;
          cached['bid_count'] = updatedAuction.bidderCount;
          CacheService.instance.cacheAuctionDetail(currentAuction.id, cached);
        }
      });
    });
  }

  Future<void> placeBidOptimistically(double amount) async {
    final previousState = state;
    
    // 1. Mise à jour optimiste (Feedback immédiat)
    state.whenData((currentAuction) {
      final updatedAuction = currentAuction.copyWith(
        currentPrice: amount,
        bidderCount: currentAuction.bidderCount + 1,
        isUserHighestBidder: true,
      );
      state = AsyncValue.data(updatedAuction);
    });

    try {
      // 2. Appel API
      final response = await _auctionApi.placeBid(
        auctionId: id,
        amount: amount,
      );

      if (!response.success) {
        // 3. Revenir en arrière si erreur
        state = previousState;
        throw Exception(response.message ?? 'Failed to place bid');
      }
      
      // 4. Mettre à jour le cache
      state.whenData((updated) {
        CacheService.instance.cacheAuctionDetail(id, {
          ...updated.toJson(), // On suppose que Auction a un toJson()
          'current_price': updated.currentPrice,
          'bid_count': updated.bidderCount,
        });
      });
    } catch (e) {
      // 3. Revenir en arrière si exception
      state = previousState;
      rethrow;
    }
  }

  /// Convertir la réponse API en modèle Auction
  Auction _mapToAuction(Map<String, dynamic> data) {
    return Auction.fromJson(data);
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
      sellerId: '',
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
      debugPrint('=== PROVIDER: BidHistory response success=${response.success}, data length=${response.data?.length} ===');
      
      if (response.success && response.data != null) {
        // API now returns List<dynamic> directly
        final bidsList = response.data!;
        debugPrint('=== PROVIDER: Extracted ${bidsList.length} bids ===');
        final result = bidsList.map((bid) => _mapToBidEntry(bid as Map<String, dynamic>)).toList();
        debugPrint('=== PROVIDER: Mapped to ${result.length} BidEntry objects ===');
        return result;
      }
      debugPrint('=== PROVIDER: Returning empty list (success=${response.success}, data=${response.data}) ===');
      return [];
    } catch (e, stackTrace) {
      debugPrint('=== PROVIDER: Error loading bid history: $e ===');
      debugPrint('=== PROVIDER: Stack trace: $stackTrace ===');
      return [];
    }
  }
  
  /// Convertir la réponse API en BidEntry
  BidEntry _mapToBidEntry(Map<String, dynamic> data) {
    // Parse amount which can be String or num from API
    final amountValue = data['amount'];
    double parsedAmount = 0.0;
    if (amountValue is String) {
      parsedAmount = double.tryParse(amountValue) ?? 0.0;
    } else if (amountValue is num) {
      parsedAmount = amountValue.toDouble();
    }
    
    return BidEntry(
      bidderName: data['bidder_name'] ?? data['user_name'] ?? 'Unknown',
      phoneNumber: data['bidder_phone'] ?? data['phone'] ?? '',
      amount: parsedAmount,
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
        final bidsList = response.data!;
        final bidList = bidsList.map((bid) => _mapToBidEntry(bid as Map<String, dynamic>)).toList();
        state = AsyncValue.data(bidList);
      } else {
        state = AsyncValue.data([]);
      }
    } catch (e) {
      debugPrint('Error refreshing bid history: $e');
      state = AsyncValue.error(e, StackTrace.current);
    }
  }
}
