import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../services/auth_api.dart';
import '../services/favorites_service.dart';

class LoginState {
  final bool isLoading;
  final String? error;

  LoginState({this.isLoading = false, this.error});

  LoginState copyWith({bool? isLoading, String? error}) {
    return LoginState(
      isLoading: isLoading ?? this.isLoading,
      error: error ?? this.error,
    );
  }
}

class LoginController extends StateNotifier<LoginState> {
  final AuthApi _authApi = AuthApi();
  
  LoginController() : super(LoginState());

  Future<bool> login(String phone, String password) async {
    state = state.copyWith(isLoading: true, error: null);
    
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
          print('Erreur synchronisation favoris: $e');
        }
        
        return true;
      } else {
        state = state.copyWith(
          isLoading: false,
          error: response.error?.message ?? 'Erreur de connexion',
        );
        return false;
      }
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }
}

final loginControllerProvider = StateNotifierProvider<LoginController, LoginState>((ref) {
  return LoginController();
});
