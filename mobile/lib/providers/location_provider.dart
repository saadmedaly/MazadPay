import 'dart:convert';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../services/location_service.dart';
import '../services/user_api.dart';

/// État de la localisation
class LocationState {
  final LocationData? location;
  final bool isLoading;
  final String? error;
  final bool hasDetected;

  LocationState({
    this.location,
    this.isLoading = false,
    this.error,
    this.hasDetected = false,
  });

  LocationState copyWith({
    LocationData? location,
    bool? isLoading,
    String? error,
    bool? hasDetected,
  }) {
    return LocationState(
      location: location ?? this.location,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      hasDetected: hasDetected ?? this.hasDetected,
    );
  }
}

/// Notifier pour gérer la localisation
class LocationNotifier extends StateNotifier<LocationState> {
  final UserApi _userApi = UserApi();

  LocationNotifier() : super(LocationState());

  /// Détecter la localisation au démarrage
  Future<void> detectLocation() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      // Essayer de charger depuis le cache d'abord
      final cached = await _loadFromCache();
      if (cached != null) {
        state = state.copyWith(
          location: cached,
          isLoading: false,
          hasDetected: true,
        );
      }

      // Détecter la vraie position
      final location = await LocationService.detectLocation();

      if (location != null) {
        // Sauvegarder en cache
        await _saveToCache(location);

        state = state.copyWith(
          location: location,
          isLoading: false,
          hasDetected: true,
        );

        // Mettre à jour le profil si la ville a changé
        await _updateProfileIfNeeded(location);
      } else {
        // Utiliser la localisation par défaut (Nouakchott)
        final defaultLocation = LocationService.getDefaultLocation();
        if (state.location == null) {
          state = state.copyWith(
            location: defaultLocation,
            isLoading: false,
            hasDetected: true,
          );
        }
      }
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: 'Erreur de localisation',
        hasDetected: true,
      );
    }
  }

  /// Rafraîchir la localisation manuellement
  Future<void> refreshLocation() async {
    state = state.copyWith(isLoading: true, error: null);

    final location = await LocationService.detectLocation();

    if (location != null) {
      await _saveToCache(location);
      state = state.copyWith(
        location: location,
        isLoading: false,
      );
      await _updateProfileIfNeeded(location);
    } else {
      state = state.copyWith(
        isLoading: false,
        error: 'Impossible de détecter la localisation',
      );
    }
  }

  /// Mettre à jour le profil utilisateur si la ville est différente
  Future<void> _updateProfileIfNeeded(LocationData location) async {
    if (location.city == null || location.city!.isEmpty) return;

    try {
      // Vérifier le profil actuel
      final response = await _userApi.getProfile();
      if (response.success && response.data != null) {
        final userData = response.data!['user'] ?? response.data!;
        final currentCity = userData['city']?.toString();

        // Mettre à jour seulement si différent
        if (currentCity != location.city) {
          await _userApi.updateProfile(city: location.city);
        }
      }
    } catch (e) {
      // Silencieux - pas critique si le profil n'est pas mis à jour
    }
  }

  /// Sauvegarder en cache
  Future<void> _saveToCache(LocationData location) async {
    try {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString('cached_location', jsonEncode(location.toJson()));
    } catch (e) {
      // Ignorer les erreurs de cache
    }
  }

  /// Charger depuis le cache
  Future<LocationData?> _loadFromCache() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final cached = prefs.getString('cached_location');
      if (cached != null) {
        return LocationData.fromJson(jsonDecode(cached));
      }
    } catch (e) {
      // Ignorer les erreurs de cache
    }
    return null;
  }

  /// Définir manuellement une ville (fallback si permission refusée)
  void setManualLocation(String city, String country) {
    final location = LocationData(
      city: city,
      country: country,
      latitude: 0.0,
      longitude: 0.0,
    );
    _saveToCache(location);
    state = state.copyWith(location: location, hasDetected: true);
  }
}

/// Provider global pour la localisation
final locationProvider = StateNotifierProvider<LocationNotifier, LocationState>(
  (ref) => LocationNotifier(),
);

/// Provider pour accéder facilement à la ville actuelle
final currentCityProvider = Provider<String?>((ref) {
  final locationState = ref.watch(locationProvider);
  return locationState.location?.city;
});

/// Provider pour savoir si la localisation est en cours de détection
final isDetectingLocationProvider = Provider<bool>((ref) {
  final locationState = ref.watch(locationProvider);
  return locationState.isLoading;
});
