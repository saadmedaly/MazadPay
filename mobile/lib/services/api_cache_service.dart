import 'dart:io';
import 'dart:convert';
import 'dart:developer' as developer;
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:dio/dio.dart';

/// Provider global pour injecter le service de cache partout dans l'application
final apiCacheServiceProvider = Provider<ApiCacheService>((ref) {
  final service = ApiCacheService();
  // Nettoyage des listeners quand le provider est détruit
  ref.onDispose(() => service.dispose());
  return service;
});

/// Objet interne représentant un cache horodaté
class CachedResponse {
  final dynamic data;
  final DateTime timestamp;

  CachedResponse({required this.data, required this.timestamp});

  Map<String, dynamic> toJson() => {
        'data': data, // La donnée API brute (Map, List, etc.)
        'timestamp': timestamp.toIso8601String(),
      };

  factory CachedResponse.fromJson(Map<String, dynamic> json) {
    return CachedResponse(
      data: json['data'],
      timestamp: DateTime.parse(json['timestamp']),
    );
  }
}

/// Service de Cache intelligent (Stale-While-Revalidate, Deduplication, Offline)
class ApiCacheService extends WidgetsBindingObserver {
  late Box _box;
  bool _isInitialized = false;

  // Déduplication : empêche 2 appels réseaux identiques simultanés
  final Map<String, Future<dynamic>> _activeRequests = {};
  
  // Annulation : stocke les tokens pour pouvoir cancel les requêtes en background
  final Map<String, CancelToken> _activeTokens = {};

  ApiCacheService() {
    WidgetsBinding.instance.addObserver(this);
  }

  /// Initialisation indispensable de Hive
  Future<void> init() async {
    if (_isInitialized) return;
    _box = await Hive.openBox('api_cache');
    await _checkCacheSize();
    _isInitialized = true;
  }

  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _cancelAllRequests("Service disposed");
  }

  // --- Gestion du Lifecycle ---
  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    // Si l'application passe en arrière-plan (paused ou inactive)
    if (state == AppLifecycleState.paused || state == AppLifecycleState.inactive) {
      developer.log('App en background: Annulation de toutes les requêtes', name: 'ApiCache');
      _cancelAllRequests("App entered background");
    }
  }

  void _cancelAllRequests(String reason) {
    for (final token in _activeTokens.values) {
      if (!token.isCancelled) {
        token.cancel(reason);
      }
    }
    _activeTokens.clear();
    _activeRequests.clear();
  }

  // --- Règles Métier (TTL) ---
  Duration _getTtl(String key) {
    // Ordre d'importance : les matchers les plus spécifiques d'abord
    if (key.startsWith(RegExp(r'^/auctions/\w+'))) {
      return const Duration(seconds: 10); // /auctions/{id}
    }
    if (key.startsWith('/auctions')) return const Duration(seconds: 20);
    if (key == '/users/me') return const Duration(minutes: 5);
    if (key == '/notifications') return const Duration(seconds: 30);
    
    return const Duration(minutes: 1); // TTL par défaut
  }

  bool _isExpired(CachedResponse cached, Duration ttl) {
    return DateTime.now().difference(cached.timestamp) > ttl;
  }

  // --- Nettoyage (Limite 30MB) ---
  Future<void> _checkCacheSize() async {
    try {
      if (_box.path != null) {
        final file = File(_box.path!);
        if (await file.exists()) {
          final size = await file.length();
          if (size > 30 * 1024 * 1024) { // 30 MB limite dure
            developer.log('Cache Hive > 30MB. Vider la base de données.', name: 'ApiCache');
            await _box.clear();
          }
        }
      }
    } catch (e) {
      developer.log('Erreur lors de la vérification de la taille du cache', error: e, name: 'ApiCache');
    }
  }

  // --- MOTEUR CENTRAL (Stale-While-Revalidate) ---
  
  /// fetcher: Doit accepter un CancelToken (fourni par Dio)
  Future<dynamic> fetchWithCache(
    String key, 
    Future<dynamic> Function(CancelToken) fetcher
  ) async {
    if (!_isInitialized) await init();

    final state = WidgetsBinding.instance.lifecycleState;
    final bool isBackground = state == AppLifecycleState.paused || state == AppLifecycleState.inactive;

    final cachedStr = _box.get(key);
    CachedResponse? cached;

    if (cachedStr != null) {
      cached = CachedResponse.fromJson(jsonDecode(cachedStr));
    }

    // 1. Protection Lifecycle : Ne jamais appeler l'API en background
    if (isBackground) {
      if (cached != null) {
        developer.log('App en background, retour du cache forcé: $key', name: 'ApiCache');
        return cached.data;
      }
      throw Exception('Impossible de requêter API : app en background et aucun cache');
    }

    final ttl = _getTtl(key);

    // 2. Cache valide → Retourne immédiatement
    if (cached != null && !_isExpired(cached, ttl)) {
      developer.log('Cache HIT (valide) : $key', name: 'ApiCache');
      return cached.data;
    }

    // 3. Pattern Stale-While-Revalidate : Cache périmé
    if (cached != null) {
      developer.log('Cache Stale, retour rapide + refresh background : $key', name: 'ApiCache');
      _refreshInBackground(key, fetcher); // Lance le téléchargement sans await
      return cached.data; // Retourne l'ancienne donnée immédiatement à l'UI
    }

    // 4. Aucun cache : l'utilisateur doit attendre
    developer.log('Cache MISS, attente du réseau : $key', name: 'ApiCache');
    return _fetchDedup(key, fetcher);
  }

  // --- Méthodes privées ---

  void _refreshInBackground(String key, Future<dynamic> Function(CancelToken) fetcher) {
    _fetchDedup(key, fetcher).catchError((e) {
      developer.log('Erreur background refresh : $key', error: e, name: 'ApiCache');
    });
  }

  /// Gestionnaire de déduplication et de requêtes réseau
  Future<dynamic> _fetchDedup(String key, Future<dynamic> Function(CancelToken) fetcher) async {
    // Déduplication : si une requête est déjà en vol, retourner le même Future
    if (_activeRequests.containsKey(key)) {
      developer.log('Déduplication activée (appel déjà en cours) : $key', name: 'ApiCache');
      return _activeRequests[key];
    }

    final cancelToken = CancelToken();
    _activeTokens[key] = cancelToken;

    // Enregistrement de la promesse
    final future = fetcher(cancelToken).then((data) {
      // 1. Sauvegarde dans Hive
      final cachedObj = CachedResponse(data: data, timestamp: DateTime.now());
      _box.put(key, jsonEncode(cachedObj.toJson()));
      
      // 2. Nettoyage de l'état local
      _activeRequests.remove(key);
      _activeTokens.remove(key);
      
      return data;
    }).catchError((e) {
      _activeRequests.remove(key);
      _activeTokens.remove(key);
      throw e;
    });

    _activeRequests[key] = future;
    return future;
  }
}
