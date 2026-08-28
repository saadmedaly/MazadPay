import 'package:mezadpay/models/api_response.dart';
import 'package:mezadpay/services/api_service.dart';
import 'package:mezadpay/services/auth_service.dart';

/// Construit l'URL absolue /v2/api/auth/... à partir d'une base API (API Versioning
/// Phase 1). Fonction pure et top-level (pas une méthode de classe) exprès, pour rester
/// testable sans instancier ApiService/AuthService (qui dépendent de dotenv).
///
/// Le Backend expose désormais deux contrats d'inscription/réinitialisation :
/// - /v1/api/auth/register|reset-password : contrat legacy (4 pays, PIN 4 chiffres),
///   conservé À L'IDENTIQUE pour la version déjà publiée sur Google Play.
/// - /v2/api/auth/register|reset-password : contrat strict (country_iso obligatoire,
///   mot de passe 8-72 caractères) — c'est CELUI-CI que cette version de l'app doit
///   utiliser exclusivement, pour ne jamais pouvoir contourner la politique v2 en
///   omettant simplement un champ.
///
/// login/otp/logout/change-password restent sur /v1/api (chemin PARTAGÉ entre les deux
/// contrats côté Backend, voir routes.go setupAuthRoutes/setupAuthRoutesV2).
String buildV2AuthUrl(String apiBaseUrl, String path) {
  final v2Base = apiBaseUrl.contains('/v1/api')
      ? apiBaseUrl.replaceFirst('/v1/api', '/v2/api')
      : '$apiBaseUrl/v2/api'; // fallback défensif si API_BASE_URL ne contient pas /v1/api
  return '$v2Base$path';
}

/// Service API pour l'authentification
class AuthApi {
  final ApiService _apiService = ApiService();
  final AuthService _authService = AuthService();

  static String _v2Url(String path) =>
      buildV2AuthUrl(ApiService.apiBaseUrl, path);

  /// Connexion utilisateur
  Future<ApiResponse<Map<String, dynamic>>> login({
    required String phone,
    required String pin,
    String? countryIso,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auth/login',
        data: {'phone': phone, 'pin': pin, 'country_iso': ?countryIso},
      );

      final apiResponse = ApiResponse<Map<String, dynamic>>.fromJson(response);

      // Si succès, sauvegarder le token
      if (apiResponse.success && apiResponse.data != null) {
        final token = apiResponse.data!['token'];
        final user = apiResponse.data!['user'];
        await _authService.saveToken(token);
        if (user != null && user['id'] != null) {
          await _authService.saveUserId(user['id'].toString());
        }
      }

      return apiResponse;
    } catch (e) {
      if (e is ApiException) return ApiResponse.error(e.message, code: e.code);
      return ApiResponse.error(e.toString(), code: 'connection_error');
    }
  }

  /// Inscription utilisateur
  Future<ApiResponse<Map<String, dynamic>>> register({
    required String phone,
    required String pin,
    required String fullName,
    required String countryIso,
    String? email,
    String? city,
    String? countryCode,
  }) async {
    try {
      // v2 (voir _v2Url) : country_iso est requis par le Backend et jamais optionnel
      // côté client — un appel qui l'omettrait serait rejeté par le contrat v2, pas
      // silencieusement dégradé vers le comportement legacy.
      final response = await _apiService.post<Map<String, dynamic>>(
        _v2Url('/auth/register'),
        data: {
          'phone': phone,
          'pin': pin,
          'full_name': fullName,
          'country_iso': countryIso,
          'email': ?email,
          'city': ?city,
          'country_code': ?countryCode,
        },
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      if (e is ApiException) return ApiResponse.error(e.message, code: e.code);
      return ApiResponse.error(e.toString(), code: 'connection_error');
    }
  }

  /// Envoyer OTP
  Future<ApiResponse<Map<String, dynamic>>> sendOTP({
    required String phone,
    required String purpose, // 'register' ou 'reset_password'
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auth/otp/send',
        data: {'phone': phone, 'purpose': purpose},
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      if (e is ApiException) return ApiResponse.error(e.message, code: e.code);
      return ApiResponse.error(e.toString(), code: 'connection_error');
    }
  }

  /// Vérifier OTP
  Future<ApiResponse<Map<String, dynamic>>> verifyOTP({
    required String phone,
    required String code,
    required String purpose,
  }) async {
    try {
      final response = await _apiService.post<Map<String, dynamic>>(
        '/auth/otp/verify',
        data: {'phone': phone, 'code': code, 'purpose': purpose},
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      if (e is ApiException) return ApiResponse.error(e.message, code: e.code);
      return ApiResponse.error(e.toString(), code: 'connection_error');
    }
  }

  /// Réinitialiser le mot de passe
  Future<ApiResponse<Map<String, dynamic>>> resetPassword({
    required String phone,
    required String newPin,
    required String otpCode,
  }) async {
    try {
      // v2 (voir _v2Url) : newPin doit respecter la politique 8-72 caractères imposée
      // par le Backend sur ce contrat — voir set_password_page/phone_password_page/
      // new_password_page pour la validation côté UI qui empêche déjà d'arriver ici
      // avec un mot de passe trop court.
      final response = await _apiService.post<Map<String, dynamic>>(
        _v2Url('/auth/reset-password'),
        data: {'phone': phone, 'new_pin': newPin, 'code': otpCode},
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      if (e is ApiException) return ApiResponse.error(e.message, code: e.code);
      return ApiResponse.error(e.toString(), code: 'connection_error');
    }
  }

  /// Changer le mot de passe
  Future<ApiResponse<Map<String, dynamic>>> changePassword({
    required String oldPin,
    required String newPin,
  }) async {
    try {
      final response = await _apiService.put<Map<String, dynamic>>(
        '/auth/change-password',
        data: {'old_pin': oldPin, 'new_pin': newPin},
      );

      return ApiResponse<Map<String, dynamic>>.fromJson(response);
    } catch (e) {
      if (e is ApiException) return ApiResponse.error(e.message, code: e.code);
      return ApiResponse.error(e.toString(), code: 'connection_error');
    }
  }

  /// Déconnexion
  Future<void> logout() async {
    try {
      await _apiService.post('/auth/logout');
    } catch (e) {
      // Ignorer l'erreur lors de la déconnexion
    } finally {
      await _authService.logout();
    }
  }
}
