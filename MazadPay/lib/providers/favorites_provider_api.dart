import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/user_api.dart';

/// Provider pour les favoris utilisant l'API backend
/// Remplace le provider mocké avec de vraies données
class FavoritesNotifierApi extends StateNotifier<Set<String>> {
  final UserApi _userApi = UserApi();
  
  FavoritesNotifierApi() : super({});
  
  /// Charger les favoris depuis l'API
  Future<void> loadFavorites() async {
    try {
      final response = await _userApi.getFavorites();
      
      if (response.success && response.data != null) {
        final auctions = response.data! as List;
        final favoriteIds = auctions
            .map((auction) => auction['auction_id']?.toString() ?? auction['id']?.toString() ?? '')
            .where((id) => id.isNotEmpty)
            .toSet();
        state = favoriteIds;
      }
    } catch (e) {
      // En cas d'erreur, garder l'état actuel
    }
  }
  
  /// Ajouter aux favoris
  Future<bool> addToFavorites(String auctionId) async {
    try {
      final response = await _userApi.addFavorite(auctionId);
      
      if (response.success) {
        state = {...state, auctionId};
        return true;
      }
      return false;
    } catch (e) {
      return false;
    }
  }
  
  /// Retirer des favoris
  Future<bool> removeFromFavorites(String auctionId) async {
    try {
      final response = await _userApi.removeFavorite(auctionId);
      
      if (response.success) {
        state = state.where((id) => id != auctionId).toSet();
        return true;
      }
      return false;
    } catch (e) {
      return false;
    }
  }
  
  /// Basculer le statut de favori
  Future<void> toggleFavorite(String auctionId) async {
    if (state.contains(auctionId)) {
      await removeFromFavorites(auctionId);
    } else {
      await addToFavorites(auctionId);
    }
  }
  
  /// Vérifier si c'est un favori
  bool isFavorite(String auctionId) {
    return state.contains(auctionId);
  }
}

/// Provider pour les favoris
final favoritesProviderApi = StateNotifierProvider<FavoritesNotifierApi, Set<String>>((ref) {
  final notifier = FavoritesNotifierApi();
  // Charger les favoris au démarrage
  notifier.loadFavorites();
  return notifier;
});
