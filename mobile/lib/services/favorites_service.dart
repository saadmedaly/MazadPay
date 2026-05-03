import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:mezadpay/services/auth_service.dart';
import 'favorites_api.dart';

/// Service hybride pour les favoris
/// Stocke localement quand hors ligne, synchronise avec le backend quand connecté
class FavoritesService {
  static const String _localFavoritesKey = 'local_favorites';
  static const String _localAuctionsKey = 'local_favorite_auctions';
  static const String _pendingSyncKey = 'favorites_pending_sync';
  
  final FavoritesApi _favoritesApi = FavoritesApi();
  final AuthService _authService = AuthService();
  
  SharedPreferences? _prefs;
  
  // Singleton
  static final FavoritesService _instance = FavoritesService._internal();
  factory FavoritesService() => _instance;
  FavoritesService._internal();
  
  /// Initialiser SharedPreferences
  Future<void> _initPrefs() async {
    _prefs ??= await SharedPreferences.getInstance();
  }
  
  /// Vérifier si l'utilisateur est connecté
  Future<bool> _isAuthenticated() async {
    final token = await _authService.getToken();
    return token != null && token.isNotEmpty;
  }
  
  /// Récupérer tous les IDs des favoris
  Future<List<String>> getFavorites() async {
    await _initPrefs();
    
    // 1. Récupérer les favoris locaux
    final localFavorites = _getLocalFavorites();
    
    // 2. Si connecté, récupérer aussi du serveur et fusionner
    if (await _isAuthenticated()) {
      try {
        final response = await _favoritesApi.getFavorites();
        if (response.success && response.data != null) {
          final serverFavorites = _extractAuctionIds(response.data!);
          
          // Fusionner (serveur + local non encore synchronisé)
          final allFavorites = {...serverFavorites, ...localFavorites}.toList();
          
          // Synchroniser les différences
          await _syncFavorites(serverFavorites, localFavorites);
          
          return allFavorites;
        }
      } catch (e) {
        // En cas d'erreur réseau, retourner les favoris locaux
        print('Erreur récupération favoris serveur: $e');
      }
    }
    
    return localFavorites;
  }
  
  /// Récupérer les données complètes des enchères favorites (depuis le cache local)
  Future<List<Map<String, dynamic>>> getFavoriteAuctions() async {
    await _initPrefs();
    return _getLocalFavoriteAuctions();
  }
  
  /// Sauvegarder les données d'une enchère favorite en cache
  Future<void> cacheFavoriteAuction(Map<String, dynamic> auction) async {
    await _initPrefs();
    final auctionId = auction['id']?.toString() ?? auction['auction_id']?.toString();
    if (auctionId == null || auctionId.isEmpty) return;
    
    final cachedAuctions = _getLocalFavoriteAuctionsMap();
    cachedAuctions[auctionId] = auction;
    await _saveLocalFavoriteAuctions(cachedAuctions);
  }
  
  /// Ajouter un favori
  Future<bool> addFavorite(String auctionId) async {
    await _initPrefs();
    
    // Ajouter localement d'abord (toujours)
    final localFavorites = _getLocalFavorites();
    if (!localFavorites.contains(auctionId)) {
      localFavorites.add(auctionId);
      await _saveLocalFavorites(localFavorites);
    }
    
    // Si connecté, synchroniser avec le serveur
    if (await _isAuthenticated()) {
      try {
        final response = await _favoritesApi.addFavorite(auctionId);
        return response.success;
      } catch (e) {
        // Marquer pour synchronisation future
        await _addToPendingSync('add', auctionId);
        return true; // Considéré comme succès local
      }
    }
    
    return true;
  }
  
  /// Supprimer un favori
  Future<bool> removeFavorite(String auctionId) async {
    await _initPrefs();
    
    // Supprimer localement d'abord (toujours)
    final localFavorites = _getLocalFavorites();
    localFavorites.remove(auctionId);
    await _saveLocalFavorites(localFavorites);
    
    // Si connecté, synchroniser avec le serveur
    if (await _isAuthenticated()) {
      try {
        final response = await _favoritesApi.removeFavorite(auctionId);
        return response.success;
      } catch (e) {
        // Marquer pour synchronisation future
        await _addToPendingSync('remove', auctionId);
        return true; // Considéré comme succès local
      }
    }
    
    return true;
  }
  
  /// Vérifier si une enchère est en favori
  Future<bool> isFavorite(String auctionId) async {
    await _initPrefs();
    final favorites = await getFavorites();
    return favorites.contains(auctionId);
  }
  
  /// Basculer le statut favori (toggle)
  Future<bool> toggleFavorite(String auctionId) async {
    final isFav = await isFavorite(auctionId);
    if (isFav) {
      return await removeFavorite(auctionId);
    } else {
      return await addFavorite(auctionId);
    }
  }
  
  /// Synchroniser tous les favoris en attente
  /// À appeler après la connexion utilisateur
  Future<void> syncPendingFavorites() async {
    await _initPrefs();
    
    if (!await _isAuthenticated()) return;
    
    final pending = _getPendingSync();
    if (pending.isEmpty) return;
    
    for (final item in pending) {
      try {
        if (item['action'] == 'add') {
          await _favoritesApi.addFavorite(item['auctionId']);
        } else if (item['action'] == 'remove') {
          await _favoritesApi.removeFavorite(item['auctionId']);
        }
      } catch (e) {
        print('Erreur synchronisation favori: $e');
        // Continuer avec les autres
      }
    }
    
    // Vider la liste des synchronisations en attente
    await _clearPendingSync();
    
    // Récupérer les favoris du serveur et mettre à jour le local
    try {
      final response = await _favoritesApi.getFavorites();
      if (response.success && response.data != null) {
        final serverFavorites = _extractAuctionIds(response.data!);
        await _saveLocalFavorites(serverFavorites);
      }
    } catch (e) {
      print('Erreur récupération favoris après sync: $e');
    }
  }
  
  /// Migrer les favoris locaux vers le serveur
  Future<void> migrateLocalFavorites() async {
    await _initPrefs();
    
    if (!await _isAuthenticated()) return;
    
    final localFavorites = _getLocalFavorites();
    if (localFavorites.isEmpty) return;
    
    // Récupérer d'abord les favoris serveur existants
    List<String> serverFavorites = [];
    try {
      final response = await _favoritesApi.getFavorites();
      if (response.success && response.data != null) {
        serverFavorites = _extractAuctionIds(response.data!);
      }
    } catch (e) {
      print('Erreur récupération favoris serveur: $e');
    }
    
    // Ajouter chaque favori local qui n'est pas déjà sur le serveur
    for (final auctionId in localFavorites) {
      if (!serverFavorites.contains(auctionId)) {
        try {
          await _favoritesApi.addFavorite(auctionId);
        } catch (e) {
          print('Erreur migration favori $auctionId: $e');
        }
      }
    }
    
    // Mettre à jour le stockage local avec la liste fusionnée
    final allFavorites = {...serverFavorites, ...localFavorites}.toList();
    await _saveLocalFavorites(allFavorites);
  }
  
  /// Récupérer le nombre de favoris locaux
  Future<int> getLocalFavoritesCount() async {
    await _initPrefs();
    return _getLocalFavorites().length;
  }
  
  /// Vider tous les favoris locaux
  Future<void> clearLocalFavorites() async {
    await _initPrefs();
    await _prefs!.remove(_localFavoritesKey);
    await _prefs!.remove(_localAuctionsKey);
    await _prefs!.remove(_pendingSyncKey);
  }
  
  // ==================== Méthodes privées ====================
  
  List<String> _getLocalFavorites() {
    final jsonString = _prefs!.getString(_localFavoritesKey);
    if (jsonString == null || jsonString.isEmpty) return [];
    
    try {
      final List<dynamic> decoded = jsonDecode(jsonString);
      return decoded.cast<String>();
    } catch (e) {
      return [];
    }
  }
  
  Future<void> _saveLocalFavorites(List<String> favorites) async {
    final jsonString = jsonEncode(favorites);
    await _prefs!.setString(_localFavoritesKey, jsonString);
  }
  
  List<Map<String, dynamic>> _getPendingSync() {
    final jsonString = _prefs!.getString(_pendingSyncKey);
    if (jsonString == null || jsonString.isEmpty) return [];
    
    try {
      final List<dynamic> decoded = jsonDecode(jsonString);
      return decoded.cast<Map<String, dynamic>>();
    } catch (e) {
      return [];
    }
  }
  
  Future<void> _addToPendingSync(String action, String auctionId) async {
    final pending = _getPendingSync();
    pending.add({'action': action, 'auctionId': auctionId, 'timestamp': DateTime.now().toIso8601String()});
    await _prefs!.setString(_pendingSyncKey, jsonEncode(pending));
  }
  
  Future<void> _clearPendingSync() async {
    await _prefs!.remove(_pendingSyncKey);
  }
  
  List<String> _extractAuctionIds(Map<String, dynamic> data) {
    final List<dynamic> favorites = data['favorites'] ?? [];
    return favorites.map((f) {
      if (f is Map<String, dynamic>) {
        return f['auction_id']?.toString() ?? f['id']?.toString() ?? '';
      }
      return f.toString();
    }).where((id) => id.isNotEmpty).toList();
  }
  
  Future<void> _syncFavorites(List<String> serverFavorites, List<String> localFavorites) async {
    // Les favoris locaux qui ne sont pas sur le serveur doivent être ajoutés
    final toAdd = localFavorites.where((id) => !serverFavorites.contains(id)).toList();
    
    for (final auctionId in toAdd) {
      try {
        await _favoritesApi.addFavorite(auctionId);
      } catch (e) {
        print('Erreur sync ajout favori $auctionId: $e');
      }
    }
  }
  
  // ==================== Cache des enchères complètes ====================
  
  /// Récupérer les enchères favorites en cache sous forme de liste
  List<Map<String, dynamic>> _getLocalFavoriteAuctions() {
    final auctionsMap = _getLocalFavoriteAuctionsMap();
    return auctionsMap.values.toList();
  }
  
  /// Récupérer les enchères favorites en cache sous forme de Map
  Map<String, Map<String, dynamic>> _getLocalFavoriteAuctionsMap() {
    final jsonString = _prefs!.getString(_localAuctionsKey);
    if (jsonString == null || jsonString.isEmpty) return {};
    
    try {
      final Map<String, dynamic> decoded = jsonDecode(jsonString);
      return decoded.map((key, value) => MapEntry(key, value as Map<String, dynamic>));
    } catch (e) {
      return {};
    }
  }
  
  /// Sauvegarder les enchères favorites en cache
  Future<void> _saveLocalFavoriteAuctions(Map<String, Map<String, dynamic>> auctions) async {
    final jsonString = jsonEncode(auctions);
    await _prefs!.setString(_localAuctionsKey, jsonString);
  }
}
