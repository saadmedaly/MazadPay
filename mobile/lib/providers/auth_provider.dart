import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:mezadpay/models/models.dart';
import 'package:mezadpay/services/auth_api.dart';
import 'package:mezadpay/services/auth_service.dart';

part 'auth_provider.g.dart';

/// État de l'authentification
class AuthState {
  final User? user;
  final bool isLoading;
  final String? error;
  final bool isAuthenticated;

  AuthState({
    this.user,
    this.isLoading = false,
    this.error,
    this.isAuthenticated = false,
  });

  AuthState copyWith({
    User? user,
    bool? isLoading,
    String? error,
    bool? isAuthenticated,
  }) {
    return AuthState(
      user: user ?? this.user,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      isAuthenticated: isAuthenticated ?? this.isAuthenticated,
    );
  }
}

/// Provider pour l'état d'authentification
@riverpod
class AuthNotifier extends _$AuthNotifier {
  final AuthApi _authApi = AuthApi();
  final AuthService _authService = AuthService();

  @override
  AuthState build() {
    _checkAuthStatus();
    return AuthState();
  }

  /// Vérifie le statut d'authentification au démarrage
  Future<void> _checkAuthStatus() async {
    final isLoggedIn = await _authService.isLoggedIn();
    if (isLoggedIn) {
      state = state.copyWith(isAuthenticated: true);
    }
  }

  /// Connexion utilisateur
  Future<bool> login({
    required String phone,
    required String pin,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _authApi.login(
        phone: phone,
        pin: pin,
      );

      if (response.success && response.data != null) {
        final userData = response.data!['user'];
        final user = userData != null ? User.fromJson(userData) : null;

        state = state.copyWith(
          isLoading: false,
          isAuthenticated: true,
          user: user,
          error: null,
        );
        return true;
      } else {
        state = state.copyWith(
          isLoading: false,
          isAuthenticated: false,
          error: response.error?.message ?? 'Erreur de connexion',
        );
        return false;
      }
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        isAuthenticated: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Inscription utilisateur
  Future<bool> register({
    required String phone,
    required String pin,
    required String fullName,
    String? email,
    String? city,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _authApi.register(
        phone: phone,
        pin: pin,
        fullName: fullName,
        email: email,
        city: city,
      );

      state = state.copyWith(
        isLoading: false,
        error: response.success ? null : (response.error?.message ?? 'Erreur d\'inscription'),
      );

      return response.success;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Envoyer OTP
  Future<bool> sendOTP({
    required String phone,
    required String purpose,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _authApi.sendOTP(
        phone: phone,
        purpose: purpose,
      );

      state = state.copyWith(
        isLoading: false,
        error: response.success ? null : (response.error?.message ?? 'Erreur d\'envoi OTP'),
      );

      return response.success;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Vérifier OTP
  Future<bool> verifyOTP({
    required String phone,
    required String code,
    required String purpose,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _authApi.verifyOTP(
        phone: phone,
        code: code,
        purpose: purpose,
      );

      state = state.copyWith(
        isLoading: false,
        error: response.success ? null : (response.error?.message ?? 'Code invalide'),
      );

      return response.success;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Réinitialiser le mot de passe
  Future<bool> resetPassword({
    required String phone,
    required String newPin,
    required String otpCode,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _authApi.resetPassword(
        phone: phone,
        newPin: newPin,
        otpCode: otpCode,
      );

      state = state.copyWith(
        isLoading: false,
        error: response.success ? null : (response.error?.message ?? 'Erreur de réinitialisation'),
      );

      return response.success;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Changer le mot de passe
  Future<bool> changePassword({
    required String oldPin,
    required String newPin,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _authApi.changePassword(
        oldPin: oldPin,
        newPin: newPin,
      );

      state = state.copyWith(
        isLoading: false,
        error: response.success ? null : (response.error?.message ?? 'Erreur de changement de mot de passe'),
      );

      return response.success;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Déconnexion
  Future<void> logout() async {
    state = state.copyWith(isLoading: true);

    try {
      await _authApi.logout();
    } catch (e) {
      // Ignorer les erreurs de déconnexion
    } finally {
      state = AuthState();
    }
  }

  /// Effacer les erreurs
  void clearError() {
    state = state.copyWith(error: null);
  }
}
