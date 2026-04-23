// lib/domain/repositories/auth_repository.dart
// Interface et implémentation du repository d'authentification

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'dart:convert';

import '../../core/errors/exceptions.dart';
import '../../data/datasources/remote/api_client.dart';
import '../../data/models/auth_models.dart';
import '../../core/constants/api_constants.dart';

// Provider pour le repository
final authRepositoryProvider = Provider<AuthRepository>((ref) {
  final dio = ref.watch(apiClientProvider);
  return AuthRepositoryImpl(dio: dio);
});

// Provider pour l'état d'authentification
final authStateProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  final repo = ref.watch(authRepositoryProvider);
  return AuthNotifier(repository: repo);
});

// Interface
abstract class AuthRepository {
  Future<LoginResponse> login(LoginRequest request);
  Future<void> register(RegisterRequest request);
  Future<void> logout();
  Future<bool> refreshToken();
  Future<void> sendOtp(OtpRequest request);
  Future<void> verifyOtp(OtpVerifyRequest request);
  Future<void> resetPassword(ResetPasswordRequest request);
  Future<void> changePin(ChangePinRequest request);
  Future<String?> getToken();
  Future<String?> getRefreshToken();
  Future<UserSession?> getSession();
  Future<void> saveSession(UserSession session);
  Future<void> clearSession();
  Future<bool> isAuthenticated();
  Future<String> getLanguagePreference();
  Future<void> setLanguagePreference(String lang);
}

// État d'authentification
class AuthState {
  final bool isAuthenticated;
  final bool isLoading;
  final UserSession? user;
  final String? error;

  AuthState({
    this.isAuthenticated = false,
    this.isLoading = false,
    this.user,
    this.error,
  });

  AuthState copyWith({
    bool? isAuthenticated,
    bool? isLoading,
    UserSession? user,
    String? error,
  }) {
    return AuthState(
      isAuthenticated: isAuthenticated ?? this.isAuthenticated,
      isLoading: isLoading ?? this.isLoading,
      user: user ?? this.user,
      error: error ?? this.error,
    );
  }
}

// Notifier pour la gestion d'état
class AuthNotifier extends StateNotifier<AuthState> {
  final AuthRepository repository;

  AuthNotifier({required this.repository}) : super(AuthState()) {
    checkAuthStatus();
  }

  Future<void> checkAuthStatus() async {
    final isAuth = await repository.isAuthenticated();
    if (isAuth) {
      final session = await repository.getSession();
      state = state.copyWith(isAuthenticated: true, user: session);
    } else {
      state = state.copyWith(isAuthenticated: false, user: null);
    }
  }

  Future<void> login(LoginRequest request) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final response = await repository.login(request);
      final session = UserSession(
        token: response.token,
        refreshToken: response.refreshToken,
        userId: response.user.id,
        phone: response.user.phone,
        fullName: response.user.fullName,
        loginTime: DateTime.now().toIso8601String(),
        language: 'ar',
      );
      await repository.saveSession(session);
      state = state.copyWith(
        isAuthenticated: true,
        isLoading: false,
        user: session,
      );
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  Future<void> logout() async {
    state = state.copyWith(isLoading: true);
    try {
      await repository.logout();
      await repository.clearSession();
      state = AuthState();
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> changeLanguage(String lang) async {
    await repository.setLanguagePreference(lang);
    if (state.user != null) {
      final updatedSession = state.user!.copyWith(language: lang);
      await repository.saveSession(updatedSession);
      state = state.copyWith(user: updatedSession);
    }
  }
}

// Implémentation
class AuthRepositoryImpl implements AuthRepository {
  final Dio dio;
  final FlutterSecureStorage secureStorage = const FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
    iOptions: IOSOptions(accessibility: KeychainAccessibility.first_unlock),
  );

  static const String _tokenKey = 'auth_token';
  static const String _refreshTokenKey = 'refresh_token';
  static const String _sessionKey = 'user_session';
  static const String _languageKey = 'language_pref';

  AuthRepositoryImpl({required this.dio});

  @override
  Future<LoginResponse> login(LoginRequest request) async {
    try {
      final response = await dio.post(
        ApiConstants.login,
        data: request.toJson(),
      );
      
      if (response.statusCode == 200 && response.data['success'] == true) {
        return LoginResponse.fromJson(response.data['data']);
      }
      
      throw UnauthorizedException(response.data['message'] ?? 'Login failed');
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<void> register(RegisterRequest request) async {
    try {
      final response = await dio.post(
        ApiConstants.register,
        data: request.toJson(),
      );
      
      if (response.statusCode != 200 && response.statusCode != 201) {
        throw ServerException(response.data['message'] ?? 'Registration failed');
      }
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<void> logout() async {
    try {
      await dio.post(ApiConstants.logout);
    } catch (e) {
      // Ignorer les erreurs de logout
    }
  }

  @override
  Future<bool> refreshToken() async {
    try {
      final refreshToken = await getRefreshToken();
      if (refreshToken == null) return false;

      final response = await dio.post(
        ApiConstants.refreshToken,
        data: {'refresh_token': refreshToken},
      );

      if (response.statusCode == 200) {
        final newToken = response.data['data']['token'];
        await secureStorage.write(key: _tokenKey, value: newToken);
        
        final newRefreshToken = response.data['data']['refresh_token'];
        if (newRefreshToken != null) {
          await secureStorage.write(key: _refreshTokenKey, value: newRefreshToken);
        }
        
        return true;
      }
      return false;
    } catch (e) {
      return false;
    }
  }

  @override
  Future<void> sendOtp(OtpRequest request) async {
    try {
      await dio.post(ApiConstants.sendOtp, data: request.toJson());
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<void> verifyOtp(OtpVerifyRequest request) async {
    try {
      await dio.post(ApiConstants.verifyOtp, data: request.toJson());
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<void> resetPassword(ResetPasswordRequest request) async {
    try {
      await dio.post(ApiConstants.resetPassword, data: request.toJson());
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<void> changePin(ChangePinRequest request) async {
    try {
      await dio.put(ApiConstants.changePassword, data: request.toJson());
    } catch (e) {
      throw ApiClient.handleError(e);
    }
  }

  @override
  Future<String?> getToken() async {
    return await secureStorage.read(key: _tokenKey);
  }

  @override
  Future<String?> getRefreshToken() async {
    return await secureStorage.read(key: _refreshTokenKey);
  }

  @override
  Future<UserSession?> getSession() async {
    final sessionJson = await secureStorage.read(key: _sessionKey);
    if (sessionJson != null) {
      return UserSession.fromJson(jsonDecode(sessionJson));
    }
    return null;
  }

  @override
  Future<void> saveSession(UserSession session) async {
    await secureStorage.write(key: _tokenKey, value: session.token);
    if (session.refreshToken != null) {
      await secureStorage.write(key: _refreshTokenKey, value: session.refreshToken!);
    }
    await secureStorage.write(
      key: _sessionKey,
      value: jsonEncode(session.toJson()),
    );
  }

  @override
  Future<void> clearSession() async {
    await secureStorage.delete(key: _tokenKey);
    await secureStorage.delete(key: _refreshTokenKey);
    await secureStorage.delete(key: _sessionKey);
  }

  @override
  Future<bool> isAuthenticated() async {
    final token = await getToken();
    return token != null && token.isNotEmpty;
  }

  @override
  Future<String> getLanguagePreference() async {
    return await secureStorage.read(key: _languageKey) ?? 'ar';
  }

  @override
  Future<void> setLanguagePreference(String lang) async {
    await secureStorage.write(key: _languageKey, value: lang);
  }
}
