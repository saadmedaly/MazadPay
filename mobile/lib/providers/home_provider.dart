import 'package:riverpod_annotation/riverpod_annotation.dart';
import '../services/auction_api.dart';
import '../services/cache_service.dart';
import '../models/auction.dart';

part 'home_provider.g.dart';

@riverpod
class HomeAuctions extends _$HomeAuctions {
  final _auctionApi = AuctionApi();

  @override
  FutureOr<List<Auction>> build() async {
    // 1. Charger depuis le cache (Instantané)
    final cachedAuctions = await CacheService.instance.getCachedAuctions();
    if (cachedAuctions != null && cachedAuctions.isNotEmpty) {
      // Retourner le cache immédiatement
      _refreshInBackground(); // Rafraîchir silencieusement
      return cachedAuctions.map((e) => Auction.fromJson(e)).toList();
    }

    // 2. Si pas de cache, charger de l'API
    return _fetchAuctions();
  }

  Future<List<Auction>> _fetchAuctions() async {
    final response = await _auctionApi.getAuctions(status: 'active', limit: 20);
    if (response.success && response.data != null) {
      final auctions = response.data!.map((e) => Auction.fromJson(e as Map<String, dynamic>)).toList();
      // Mettre à jour le cache
      final jsonAuctions = auctions.map((a) => a.toJson()).toList();
      await CacheService.instance.cacheAuctions(jsonAuctions);
      return auctions;
    }
    throw Exception(response.message ?? 'Failed to load auctions');
  }

  Future<void> _refreshInBackground() async {
    try {
      final auctions = await _fetchAuctions();
      state = AsyncValue.data(auctions);
    } catch (e) {
      // Ignorer les erreurs de rafraîchissement silencieux
      print('Silent refresh failed: $e');
    }
  }
}
