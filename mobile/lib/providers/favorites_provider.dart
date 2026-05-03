import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:mezadpay/services/favorites_service.dart';

part 'favorites_provider.g.dart';

/// Provider pour les favoris avec persistance locale
@riverpod
class Favorites extends _$Favorites {
  final FavoritesService _favoritesService = FavoritesService();

  @override
  Future<Set<String>> build() async {
    // Charger les favoris depuis le stockage local/serveur
    final favorites = await _favoritesService.getFavorites();
    return favorites.toSet();
  }

  /// Basculer le statut favori d'une enchère
  Future<void> toggleFavorite(String auctionId) async {
    final currentFavorites = state.value ?? {};
    
    // Mettre à jour localement d'abord pour UI réactive
    if (currentFavorites.contains(auctionId)) {
      state = AsyncValue.data({...currentFavorites}..remove(auctionId));
    } else {
      state = AsyncValue.data({...currentFavorites, auctionId});
    }
    
    // Persister le changement
    try {
      await _favoritesService.toggleFavorite(auctionId);
      // Recharger pour s'assurer de la cohérence
      ref.invalidateSelf();
    } catch (e) {
      // En cas d'erreur, recharger l'état précédent
      ref.invalidateSelf();
    }
  }

  /// Ajouter un favori
  Future<void> addFavorite(String auctionId) async {
    final currentFavorites = state.value ?? {};
    
    if (!currentFavorites.contains(auctionId)) {
      // Mise à jour UI immédiate
      state = AsyncValue.data({...currentFavorites, auctionId});
      
      // Persister
      try {
        await _favoritesService.addFavorite(auctionId);
      } catch (e) {
        ref.invalidateSelf();
      }
    }
  }

  /// Supprimer un favori
  Future<void> removeFavorite(String auctionId) async {
    final currentFavorites = state.value ?? {};
    
    if (currentFavorites.contains(auctionId)) {
      // Mise à jour UI immédiate
      state = AsyncValue.data({...currentFavorites}..remove(auctionId));
      
      // Persister
      try {
        await _favoritesService.removeFavorite(auctionId);
      } catch (e) {
        ref.invalidateSelf();
      }
    }
  }

  /// Vérifier si une enchère est en favori
  bool isFavorite(String auctionId) {
    return state.value?.contains(auctionId) ?? false;
  }

  /// Synchroniser les favoris avec le serveur
  /// À appeler après la connexion utilisateur
  Future<void> syncWithServer() async {
    state = const AsyncValue.loading();
    try {
      await _favoritesService.syncPendingFavorites();
      await _favoritesService.migrateLocalFavorites();
      final favorites = await _favoritesService.getFavorites();
      state = AsyncValue.data(favorites.toSet());
    } catch (e) {
      state = AsyncValue.error(e, StackTrace.current);
    }
  }

  /// Rafraîchir la liste des favoris
  Future<void> refresh() async {
    ref.invalidateSelf();
  }
}

/// Provider pour le nombre de favoris
@riverpod
Future<int> favoritesCount(FavoritesCountRef ref) async {
  final favorites = await ref.watch(favoritesProvider.future);
  return favorites.length;
}

/// Provider pour vérifier si une enchère spécifique est en favori
@riverpod
bool isAuctionFavorite(IsAuctionFavoriteRef ref, String auctionId) {
  final favoritesAsync = ref.watch(favoritesProvider);
  return favoritesAsync.value?.contains(auctionId) ?? false;
}
