import 'package:flutter/foundation.dart' show debugPrint;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/auth_api.dart';
import '../services/favorites_service.dart';

class LoginState {
  final bool isLoading;
  final String? error;
  final String? errorCode;

  LoginState({this.isLoading = false, this.error, this.errorCode});

  LoginState copyWith({bool? isLoading, String? error, String? errorCode}) {
    return LoginState(
      isLoading: isLoading ?? this.isLoading,
      error: error ?? this.error,
      errorCode: errorCode ?? this.errorCode,
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
        state = state.copyWith(isLoading: false);
        
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
