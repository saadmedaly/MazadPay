import 'package:flutter/foundation.dart' show debugPrint;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/auth_api.dart';
import '../services/favorites_service.dart';

class LoginState {
  final bool isLoading;
  final String? error;
  final String? errorCode;

  LoginState({this.isLoading = false, this.error, this.errorCode});

  // Sentinel utilisé pour distinguer "paramètre non fourni" de "null explicite"
  // (Mobile Auth Phase 1) : avec `String? error`, `error ?? this.error` ne pouvait
  // jamais effacer une erreur existante, puisque `null ?? this.error` retombe
  // toujours sur l'ancienne valeur — le message d'erreur restait donc affiché
  // indéfiniment après une nouvelle tentative réussie.
  static const _unset = Object();

  LoginState copyWith({
    bool? isLoading,
    Object? error = _unset,
    Object? errorCode = _unset,
  }) {
    return LoginState(
      isLoading: isLoading ?? this.isLoading,
      error: identical(error, _unset) ? this.error : error as String?,
      errorCode: identical(errorCode, _unset) ? this.errorCode : errorCode as String?,
    );
  }
}

class LoginController extends StateNotifier<LoginState> {
  final AuthApi _authApi = AuthApi();
  
  LoginController() : super(LoginState());

  Future<bool> login(String phone, String password) async {
    state = state.copyWith(isLoading: true, error: null, errorCode: null);
    
    try {
      final response = await _authApi.login(
        phone: phone,
        pin: password,
      );
      
      if (response.success) {
        state = state.copyWith(isLoading: false, error: null, errorCode: null);
        
        // Synchroniser les favoris locaux avec le serveur
        try {
          final favoritesService = FavoritesService();
          await favoritesService.syncPendingFavorites();
          await favoritesService.migrateLocalFavorites();
        } catch (e) {
          // Ne pas bloquer le login si la sync échoue
          debugPrint('Erreur synchronisation favoris: $e');
        }
        
        return true;
      } else {
        state = state.copyWith(
          isLoading: false,
          error: response.error?.message ?? 'Erreur de connexion',
          errorCode: response.error?.code,
        );
        return false;
      }
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
        errorCode: 'connection_error',
      );
      return false;
    }
  }
}

final loginControllerProvider = StateNotifierProvider<LoginController, LoginState>((ref) {
  return LoginController();
});
