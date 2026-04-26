import 'package:riverpod_annotation/riverpod_annotation.dart';
import 'package:mezadpay/models/models.dart';
import 'package:mezadpay/services/user_api.dart';

part 'user_provider.g.dart';

/// État de l'utilisateur
class UserState {
  final User? user;
  final bool isLoading;
  final String? error;
  final bool isInitialized;

  UserState({
    this.user,
    this.isLoading = false,
    this.error,
    this.isInitialized = false,
  });

  UserState copyWith({
    User? user,
    bool? isLoading,
    String? error,
    bool? isInitialized,
  }) {
    return UserState(
      user: user ?? this.user,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      isInitialized: isInitialized ?? this.isInitialized,
    );
  }
}

/// Provider pour la gestion du profil utilisateur
@riverpod
class UserNotifier extends _$UserNotifier {
  final UserApi _userApi = UserApi();

  @override
  UserState build() {
    return UserState();
  }

  /// Charger le profil utilisateur
  Future<void> loadProfile() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _userApi.getProfile();

      if (response.success && response.data != null) {
        final user = User.fromJson(response.data!);
        state = state.copyWith(
          isLoading: false,
          user: user,
          isInitialized: true,
          error: null,
        );
      } else {
        state = state.copyWith(
          isLoading: false,
          error: response.error?.message ?? 'Erreur de chargement du profil',
          isInitialized: true,
        );
      }
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
        isInitialized: true,
      );
    }
  }

  /// Mettre à jour le profil
  Future<bool> updateProfile({
    String? fullName,
    String? email,
    String? phone,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _userApi.updateProfile(
        fullName: fullName,
        email: email,
        phone: phone,
      );

      if (response.success) {
        // Recharger le profil pour obtenir les données à jour
        await loadProfile();
        return true;
      } else {
        state = state.copyWith(
          isLoading: false,
          error: response.error?.message ?? 'Erreur de mise à jour',
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

  /// Mettre à jour l'avatar
  Future<bool> updateAvatar(String avatarUrl) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _userApi.updateAvatar(avatarUrl: avatarUrl);

      if (response.success) {
        await loadProfile();
        return true;
      } else {
        state = state.copyWith(
          isLoading: false,
          error: response.error?.message ?? 'Erreur de mise à jour de l\'avatar',
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

  /// Changer la langue
  Future<bool> updateLanguage(String language) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _userApi.updateLanguage(language: language);

      if (response.success) {
        await loadProfile();
        return true;
      } else {
        state = state.copyWith(
          isLoading: false,
          error: response.error?.message ?? 'Erreur de changement de langue',
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

  /// Mettre à jour les préférences de notifications
  Future<bool> updateNotificationPrefs(Map<String, dynamic> preferences) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _userApi.updateNotificationPrefs(
        preferences: preferences,
      );

      if (response.success) {
        await loadProfile();
        return true;
      } else {
        state = state.copyWith(
          isLoading: false,
          error: response.error?.message ?? 'Erreur de mise à jour des préférences',
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

  /// Récupérer le statut KYC
  Future<Map<String, dynamic>?> getKYCStatus() async {
    try {
      final response = await _userApi.getKYCStatus();

      if (response.success && response.data != null) {
        return response.data;
      }
      return null;
    } catch (e) {
      return null;
    }
  }

  /// Soumettre KYC
  Future<bool> submitKYC({
    required String idCardFront,
    required String idCardBack,
    required String selfie,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final response = await _userApi.submitKYC(
        idCardFront: idCardFront,
        idCardBack: idCardBack,
        selfie: selfie,
      );

      state = state.copyWith(isLoading: false);
      return response.success;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      return false;
    }
  }

  /// Effacer les erreurs
  void clearError() {
    state = state.copyWith(error: null);
  }

  /// Rafraîchir le profil
  Future<void> refresh() async {
    await loadProfile();
  }
}
